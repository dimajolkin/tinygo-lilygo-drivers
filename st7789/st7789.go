// Package st7789 implements a driver for the ST7789 TFT displays optimized for LilyGo devices,
// it comes in various screen sizes and includes LilyGo-specific initialization sequences.
//
// Datasheets: https://cdn-shop.adafruit.com/product-files/3787/3787_tft_QT154H2201__________20190228182902.pdf
//
//	http://www.newhavendisplay.com/appnotes/datasheets/LCDs/ST7789V.pdf
package st7789

import (
	"image/color"
	"machine"
	"time"

	"errors"

	drivers "github.com/dimajolkin/tinygo-lilygo-drivers"
)

// Rotation controls the rotation used by the display.
type Rotation uint8

const (
	Rotation0   Rotation = 0
	Rotation90  Rotation = 1
	Rotation180 Rotation = 2
	Rotation270 Rotation = 3

	// Old rotations consts
	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)

// Device wraps an SPI connection to ST7789 display
type Device struct {
	spi      drivers.SPI
	dcPin    machine.Pin
	resetPin machine.Pin
	csPin    machine.Pin
	blPin    machine.Pin
	width    int16
	height   int16
	rotation Rotation

	// Optimization buffers
	largeBuffer []uint8           // Large buffer for DMA-like transfer
	colorCache  map[uint32]uint16 // Cache for converted RGB565 colors
	lastWindow  [4]int16          // Cache for last window [x, y, w, h]
}

// Config is the configuration for the display
type Config struct {
	Width    int16
	Height   int16
	Rotation Rotation

	// LilyGo-specific gamma control. Look in the LCD panel datasheet or provided example code
	// to find these values. If not set, the LilyGo optimized defaults will be used.
	PVGAMCTRL []uint8 // Positive voltage gamma control (14 bytes)
	NVGAMCTRL []uint8 // Negative voltage gamma control (14 bytes)
}

var (
	errOutOfBounds = errors.New("rectangle coordinates outside display area")
)

// New creates a new ST7789 connection optimized for LilyGo devices. The SPI wire must already be configured.
func New(spi drivers.SPI, resetPin, dcPin, csPin, blPin machine.Pin) *Device {
	// Configure pins
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Initial pin states
	csPin.High()    // CS inactive (HIGH)
	dcPin.High()    // DC in data mode
	resetPin.High() // RST inactive (HIGH)
	blPin.Low()     // Backlight off initially

	return &Device{
		spi:         spi,
		dcPin:       dcPin,
		resetPin:    resetPin,
		csPin:       csPin,
		blPin:       blPin,
		width:       240, // Default width
		height:      320, // Default height
		rotation:    Rotation0,
		largeBuffer: make([]uint8, 4096),         // 4KB buffer (2048 pixels)
		colorCache:  make(map[uint32]uint16, 16), // Cache for 16 colors
		lastWindow:  [4]int16{-1, -1, -1, -1},    // Invalid cache
	}
}

// Configure initializes the display with LilyGo-optimized configuration
func (d *Device) Configure(cfg Config) error {
	// Set dimensions
	if cfg.Width != 0 {
		d.width = cfg.Width
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	}
	d.rotation = cfg.Rotation

	// Hardware reset sequence
	d.resetPin.Low()
	time.Sleep(20 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(150 * time.Millisecond)

	// Begin initialization sequence
	d.startWrite()

	// LilyGo T-Deck optimized initialization sequence (based on 20240726 changes)

	// Sleep Out
	d.sendCommand(SLPOUT, nil)
	time.Sleep(120 * time.Millisecond)

	// Color Mode - 16bit (RGB565)
	d.sendCommand(COLMOD, []byte{0x55})
	time.Sleep(10 * time.Millisecond)

	// Memory Access Control - RGB color order
	d.sendCommand(MADCTL, []byte{0x00})

	// ST7789V Frame rate setting
	d.sendCommand(PORCTRL, []byte{0x0c, 0x0c, 0x00, 0x33, 0x33})

	// Voltages: VGH / VGL
	d.sendCommand(GCTRL, []byte{0x85})

	// VCOMS
	d.sendCommand(VCOMS, []byte{0x3F})

	// LCMCTRL
	d.sendCommand(LCMCTRL, []byte{0x2c})

	// VDVVRHEN
	d.sendCommand(VDVVRHEN, []byte{0x01})

	// VRHS voltage
	d.sendCommand(VRHS, []byte{0x13})

	// VDVSET
	d.sendCommand(VDVS, []byte{0x20})

	// Frame Rate Control in Normal Mode
	d.sendCommand(FRCTRL2, []byte{0x0f})

	// Power Control 1
	d.sendCommand(PWCTRL1, []byte{0xa4, 0xa1})

	// Set gamma tables
	if len(cfg.PVGAMCTRL) == 14 {
		d.sendCommand(GMCTRP1, cfg.PVGAMCTRL)
	} else {
		// LilyGo optimized positive gamma
		d.sendCommand(GMCTRP1, []byte{
			0xd0, 0x00, 0x05, 0x0e, 0x15, 0x0d, 0x37,
			0x43, 0x47, 0x09, 0x15, 0x12, 0x16, 0x19,
		})
	}

	if len(cfg.NVGAMCTRL) == 14 {
		d.sendCommand(GMCTRN1, cfg.NVGAMCTRL)
	} else {
		// LilyGo optimized negative gamma
		d.sendCommand(GMCTRN1, []byte{
			0xd0, 0x00, 0x02, 0x07, 0x0a, 0x28, 0x31,
			0x54, 0x47, 0x0e, 0x1c, 0x17, 0x1b, 0x1e,
		})
	}

	// Display Inversion On - critical for LilyGo T-Deck!
	d.sendCommand(INVON, nil)

	// Set initial window to full screen
	d.setWindow(0, 0, d.width, d.height)
	d.fillScreen(color.RGBA{0, 0, 0, 255}) // Clear screen

	// Display Brightness Control (maximum brightness)
	d.sendCommand(WRDISBV, []byte{0xFF}) // Maximum brightness 255

	// Display Control (enable brightness control)
	d.sendCommand(WRCTRLD, []byte{0x2C}) // Enable Brightness Control

	// Content Adaptive Brightness Control
	d.sendCommand(WRCABC, []byte{0x01}) // Enable CABC for UI images

	// Normal mode ON
	d.sendCommand(NORON, nil)
	time.Sleep(10 * time.Millisecond)

	// Display ON
	d.sendCommand(DISPON, nil)
	time.Sleep(100 * time.Millisecond)

	d.endWrite()

	// Enable backlight
	d.blPin.High()

	return nil
}

// Send a command with data to the display. It does not change the chip select
// pin (it must be low when calling). The DC pin is left high after return,
// meaning that data can be sent right away.
func (d *Device) sendCommand(command uint8, data []byte) error {
	d.dcPin.Low()
	_, err := d.spi.Transfer(command)
	d.dcPin.High()
	if len(data) != 0 {
		for _, b := range data {
			d.spi.Transfer(b)
		}
	}
	return err
}

// startWrite must be called at the beginning of all exported methods to set the
// chip select pin low.
func (d *Device) startWrite() {
	if d.csPin != machine.NoPin {
		d.csPin.Low()
	}
}

// endWrite must be called at the end of all exported methods to set the chip
// select pin high.
func (d *Device) endWrite() {
	if d.csPin != machine.NoPin {
		d.csPin.High()
	}
}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *Device) Display() error {
	return nil
}

// SetPixel sets a pixel in the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || y < 0 ||
		(((d.rotation == Rotation0 || d.rotation == Rotation180) && (x >= d.width || y >= d.height)) ||
			((d.rotation == Rotation90 || d.rotation == Rotation270) && (x >= d.height || y >= d.width))) {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	d.sendCommand(CASET, []byte{uint8(x >> 8), uint8(x), uint8((x + w - 1) >> 8), uint8(x + w - 1)})
	d.sendCommand(RASET, []byte{uint8(y >> 8), uint8(y), uint8((y + h - 1) >> 8), uint8(y + h - 1)})
	d.sendCommand(RAMWR, nil)
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	d.startWrite()
	err := d.fillRectangle(x, y, width, height, c)
	d.endWrite()
	return err
}

func (d *Device) fillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errOutOfBounds
	}
	d.setWindow(x, y, width, height)

	// Convert color to RGB565 with caching
	color565 := d.colorToRGB565(c)
	colorHi := uint8(color565 >> 8)
	colorLo := uint8(color565 & 0xFF)

	// Send color data
	d.dcPin.High() // Data mode
	totalPixels := int32(width) * int32(height)

	// Choose optimal sending method
	if totalPixels > 2000 {
		// Super large areas - use maximum buffer
		d.fillSuperLargeArea(totalPixels, colorHi, colorLo)
	} else if totalPixels > 500 {
		// Large areas - use standard buffering
		d.fillLargeArea(totalPixels, colorHi, colorLo)
	} else {
		// Small areas - direct send
		for i := int32(0); i < totalPixels; i++ {
			d.spi.Transfer(colorHi)
			d.spi.Transfer(colorLo)
		}
	}

	return nil
}

// Optimized fill for large areas
func (d *Device) fillLargeArea(totalPixels int32, colorHi, colorLo uint8) {
	// Create buffer for sending in blocks
	bufferSize := int32(512) // 512 byte buffer (256 pixels)
	if totalPixels*2 < bufferSize {
		bufferSize = totalPixels * 2
	}

	// Fill buffer with color
	buffer := make([]uint8, bufferSize)
	for i := int32(0); i < bufferSize; i += 2 {
		buffer[i] = colorHi
		buffer[i+1] = colorLo
	}

	// Send full buffers
	remainingBytes := totalPixels * 2
	for remainingBytes > 0 {
		sendSize := bufferSize
		if remainingBytes < bufferSize {
			sendSize = remainingBytes
		}

		// Send block of data
		for i := int32(0); i < sendSize; i++ {
			d.spi.Transfer(buffer[i])
		}

		remainingBytes -= sendSize
	}
}

// Super-optimized fill for maximum areas (using large buffer)
func (d *Device) fillSuperLargeArea(totalPixels int32, colorHi, colorLo uint8) {
	// Use large pre-allocated buffer
	bufferSize := int32(len(d.largeBuffer))
	if bufferSize > totalPixels*2 {
		bufferSize = totalPixels * 2
	}

	// Fill large buffer with color
	for i := int32(0); i < bufferSize; i += 2 {
		d.largeBuffer[i] = colorHi
		d.largeBuffer[i+1] = colorLo
	}

	// Send with maximum block size
	remainingBytes := totalPixels * 2
	for remainingBytes > 0 {
		sendSize := bufferSize
		if remainingBytes < bufferSize {
			sendSize = remainingBytes
		}

		// Send block of data in one piece
		for i := int32(0); i < sendSize; i++ {
			d.spi.Transfer(d.largeBuffer[i])
		}

		remainingBytes -= sendSize
	}
}

// Fast color conversion to RGB565 with caching
func (d *Device) colorToRGB565(c color.RGBA) uint16 {
	// Create cache key from RGBA
	key := (uint32(c.R) << 24) | (uint32(c.G) << 16) | (uint32(c.B) << 8) | uint32(c.A)

	// Check cache
	if cached, exists := d.colorCache[key]; exists {
		return cached
	}

	// Convert to RGB565
	r := uint16(c.R) >> 3
	g := uint16(c.G) >> 2
	b := uint16(c.B) >> 3
	color565 := (r << 11) | (g << 5) | b

	// Save to cache
	d.colorCache[key] = color565

	return color565
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *Device) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *Device) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	d.FillRectangle(x0, y, x1-x0+1, 1, c)
}

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	d.startWrite()
	d.fillScreen(c)
	d.endWrite()
}

func (d *Device) fillScreen(c color.RGBA) {
	// Force clear window cache for full screen
	d.lastWindow = [4]int16{-1, -1, -1, -1}

	if d.rotation == Rotation0 || d.rotation == Rotation180 {
		d.fillRectangle(0, 0, d.width, d.height, c)
	} else {
		d.fillRectangle(0, 0, d.height, d.width, c)
	}
}

// Rotation returns the current rotation of the device.
func (d *Device) Rotation() Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device (clock-wise)
func (d *Device) SetRotation(rotation Rotation) error {
	d.rotation = rotation
	d.startWrite()
	err := d.setRotation(rotation)
	d.endWrite()
	return err
}

func (d *Device) setRotation(rotation Rotation) error {
	madctl := uint8(0)
	switch rotation % 4 {
	case Rotation0:
		madctl = 0x00
	case Rotation90:
		madctl = MADCTL_MX | MADCTL_MV
	case Rotation180:
		madctl = MADCTL_MX | MADCTL_MY
	case Rotation270:
		madctl = MADCTL_MY | MADCTL_MV
	}

	return d.sendCommand(MADCTL, []byte{madctl})
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == Rotation90 || d.rotation == Rotation270 {
		return d.height, d.width
	}
	return d.width, d.height
}

// EnableBacklight enables or disables the backlight
func (d *Device) EnableBacklight(enable bool) {
	if enable {
		d.blPin.High()
	} else {
		d.blPin.Low()
	}
}

// Sleep sets the sleep mode for this LCD panel. When sleeping, the panel uses a lot
// less power. The LCD won't display an image anymore, but the memory contents
// will be kept.
func (d *Device) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		d.startWrite()
		d.sendCommand(SLPIN, nil)
		d.endWrite()
		time.Sleep(5 * time.Millisecond) // 5ms required by the datasheet
	} else {
		// Turn the LCD panel back on.
		d.startWrite()
		d.sendCommand(SLPOUT, nil)
		d.endWrite()
		// Note: the st7789 documentation says that it is needed to wait at
		// least 120ms before going to sleep again. Sleeping here would not be
		// practical (delays turning on the screen too much), so just hope the
		// screen won't need to sleep again for at least 120ms.
		// In practice, it's unlikely the user will set the display to sleep
		// again within 120ms.
	}
	return nil
}

// InvertColors inverts the colors of the screen
func (d *Device) InvertColors(invert bool) {
	d.startWrite()
	if invert {
		d.sendCommand(INVON, nil)
	} else {
		d.sendCommand(INVOFF, nil)
	}
	d.endWrite()
}
