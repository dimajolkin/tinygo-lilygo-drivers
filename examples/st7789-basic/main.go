// Basic ST7789 example for LilyGo T-Deck
// Demonstrates basic display functionality with optimized performance
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo-lilygo-drivers/st7789"
)

// LilyGo T-Deck pin configuration
const (
	TFT_SCLK machine.Pin = 40 // SCK - Serial Clock
	TFT_MOSI machine.Pin = 41 // MOSI - Master Out Slave In
	TFT_CS   machine.Pin = 12 // Chip Select
	TFT_DC   machine.Pin = 11 // Data/Command
	TFT_RST  machine.Pin = 10 // Reset
	TFT_BL   machine.Pin = 42 // Backlight
)

// Predefined colors for testing
var (
	RED     = color.RGBA{255, 0, 0, 255}
	GREEN   = color.RGBA{0, 255, 0, 255}
	BLUE    = color.RGBA{0, 0, 255, 255}
	WHITE   = color.RGBA{255, 255, 255, 255}
	BLACK   = color.RGBA{0, 0, 0, 255}
	YELLOW  = color.RGBA{255, 255, 0, 255}
	CYAN    = color.RGBA{0, 255, 255, 255}
	MAGENTA = color.RGBA{255, 0, 255, 255}
)

func main() {
	// Configure SPI for maximum performance
	spi := machine.SPI1
	spi.Configure(machine.SPIConfig{
		Frequency: 80000000, // 80 MHz - maximum speed for ST7789
		SCK:       TFT_SCLK,
		SDO:       TFT_MOSI,
		Mode:      0, // SPI Mode 0
	})

	// Create display driver
	display := st7789.New(&spi, TFT_RST, TFT_DC, TFT_CS, TFT_BL)

	// Configure display with LilyGo T-Deck settings
	err := display.Configure(st7789.Config{
		Width:    240, // T-Deck display width
		Height:   320, // T-Deck display height
		Rotation: st7789.Rotation0,
	})
	if err != nil {
		println("Display configuration error:", err.Error())
		return
	}

	println("LilyGo T-Deck ST7789 Display Test")

	// Run display tests
	testBasicColors(display)
	testGeometry(display)
	testRotations(display)
	testPerformance(display)
}

// Test basic color fills
func testBasicColors(display *st7789.Device) {
	println("Testing basic colors...")

	colors := []color.RGBA{RED, GREEN, BLUE, WHITE, BLACK, YELLOW, CYAN, MAGENTA}
	names := []string{"Red", "Green", "Blue", "White", "Black", "Yellow", "Cyan", "Magenta"}

	for i, col := range colors {
		println("Color:", names[i])
		display.FillScreen(col)
		time.Sleep(1 * time.Second)
	}
}

// Test geometric shapes
func testGeometry(display *st7789.Device) {
	println("Testing geometry...")
	display.FillScreen(BLACK)

	w, h := display.Size()
	println("Display size:", int(w), "x", int(h))

	// Draw rectangles
	rectW := w / 4
	display.FillRectangle(10, 10, rectW, 50, RED)
	display.FillRectangle(10+rectW+10, 10, rectW, 50, GREEN)
	display.FillRectangle(10+2*(rectW+10), 10, rectW, 50, BLUE)

	// Draw full-width stripes
	fullW := w - 20
	display.FillRectangle(10, 70, fullW, 30, YELLOW)
	display.FillRectangle(10, 110, fullW, 30, WHITE)

	// Draw lines
	display.DrawFastHLine(0, w-1, 150, CYAN)
	display.DrawFastVLine(w/2, 160, h-160, MAGENTA)

	time.Sleep(3 * time.Second)
}

// Test screen rotations
func testRotations(display *st7789.Device) {
	println("Testing rotations...")

	rotations := []st7789.Rotation{
		st7789.Rotation0, st7789.Rotation90,
		st7789.Rotation180, st7789.Rotation270,
	}

	for i, rotation := range rotations {
		println("Rotation:", i*90, "degrees")
		display.SetRotation(rotation)
		display.FillScreen(BLACK)

		w, h := display.Size()
		display.FillRectangle(0, 0, w/4, h/4, RED)
		display.FillRectangle(3*w/4, 3*h/4, w/4, h/4, GREEN)

		time.Sleep(2 * time.Second)
	}

	// Return to original orientation
	display.SetRotation(st7789.Rotation0)
}

// Performance test
func testPerformance(display *st7789.Device) {
	println("Performance test...")
	w, h := display.Size()

	// Test 1: Full screen fills
	println("Test 1: Full screen fills")
	start := time.Now()
	for i := 0; i < 10; i++ {
		display.FillScreen(RED)
		display.FillScreen(GREEN)
		display.FillScreen(BLUE)
	}
	elapsed := time.Since(start)
	println("30 screen fills in:", elapsed.String())
	println("Average per fill:", (elapsed / 30).String())

	// Test 2: Small rectangles
	println("Test 2: Small rectangles")
	display.FillScreen(BLACK)
	start = time.Now()
	for i := int16(0); i < 100; i++ {
		x := (i * 3) % (w - 10)
		y := (i * 2) % (h - 10)
		display.FillRectangle(x, y, 10, 10, RED)
	}
	elapsed = time.Since(start)
	println("100 small rectangles in:", elapsed.String())

	// Calculate throughput
	totalPixels := int32(w) * int32(h) * 30 // 30 screen fills
	bytesPerPixel := int32(2)               // RGB565 = 2 bytes per pixel
	totalBytes := totalPixels * bytesPerPixel
	throughputMBps := float64(totalBytes) / (elapsed.Seconds() / 30) / 1024 / 1024

	println("Throughput:", throughputMBps, "MB/s")

	// Final blinking animation
	println("Final animation...")
	for {
		display.FillScreen(RED)
		time.Sleep(200 * time.Millisecond)
		display.FillScreen(GREEN)
		time.Sleep(200 * time.Millisecond)
		display.FillScreen(BLUE)
		time.Sleep(200 * time.Millisecond)
	}
}
