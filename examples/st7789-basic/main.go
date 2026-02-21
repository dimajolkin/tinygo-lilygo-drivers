// Basic ST7789 example for LilyGo T-Deck (official TinyGo driver: tinygo.org/x/drivers/st7789).
// Color test and basic display check: bars, grays, geometry, rotations.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
)

const (
	TFT_SCLK machine.Pin = 40
	TFT_MOSI machine.Pin = 41
	TFT_CS   machine.Pin = 12
	TFT_DC   machine.Pin = 11
	TFT_RST  machine.Pin = 10
	TFT_BL   machine.Pin = 42
)

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
	time.Sleep(1 * time.Second)

	spi := machine.SPI1
	spi.Configure(machine.SPIConfig{
		Frequency: 80000000,
		SCK:       TFT_SCLK,
		SDO:       TFT_MOSI,
		Mode:      0,
	})

	display := st7789.New(spi, TFT_RST, TFT_DC, TFT_CS, TFT_BL)
	display.Configure(st7789.Config{
		Width:    240,
		Height:   320,
		Rotation: drivers.Rotation90,
	})

	println("ST7789 color test")
	testColorPalette(&display)
}

func testColorPalette(display *st7789.Device) {
	println("Test: full color palette")
	w, h := display.Size()
	const cols, rows = 32, 24
	cw := w / cols
	ch := h / rows
	for row := int16(0); row < rows; row++ {
		g := uint8((uint16(row) * 255) / (rows - 1))
		for col := int16(0); col < cols; col++ {
			r := uint8((uint16(col) * 255) / (cols - 1))
			c := color.RGBA{R: r, G: g, B: 128, A: 255}
			x := col * cw
			y := row * ch
			display.FillRectangle(x, y, cw, ch, c)
		}
	}
	time.Sleep(5 * time.Second)
}

func testColorBars(display *st7789.Device) {
	println("Test: color bars")
	w, h := display.Size()
	colors := []color.RGBA{RED, GREEN, BLUE, WHITE, YELLOW, CYAN, MAGENTA, BLACK}
	n := int16(len(colors))
	barH := h / n
	for i := int16(0); i < n; i++ {
		y0 := i * barH
		display.FillRectangle(0, y0, w, barH, colors[i])
	}
	time.Sleep(4 * time.Second)
}

func testGrayRamp(display *st7789.Device) {
	println("Test: gray ramp")
	w, h := display.Size()
	display.FillScreen(BLACK)
	stripH := h / 4
	nSteps := int16(16)
	stepW := w / nSteps
	for i := int16(0); i < nSteps; i++ {
		v := uint8((uint16(i) * 255) / uint16(nSteps))
		c := color.RGBA{v, v, v, 255}
		display.FillRectangle(i*stepW, 0, stepW, stripH, c)
	}
	display.FillRectangle(0, stripH, w, stripH*3, color.RGBA{80, 80, 100, 255})
	time.Sleep(3 * time.Second)
}

func testFullScreenColors(display *st7789.Device) {
	println("Test: full-screen colors")
	colors := []color.RGBA{RED, GREEN, BLUE, WHITE, BLACK, YELLOW, CYAN, MAGENTA}
	for _, c := range colors {
		display.FillScreen(c)
		time.Sleep(1 * time.Second)
	}
}

func testGeometry(display *st7789.Device) {
	println("Test: geometry")
	display.FillScreen(BLACK)
	w, h := display.Size()

	rectW := w / 4
	display.FillRectangle(10, 10, rectW, 50, RED)
	display.FillRectangle(10+rectW+10, 10, rectW, 50, GREEN)
	display.FillRectangle(10+2*(rectW+10), 10, rectW, 50, BLUE)

	display.FillRectangle(10, 70, w-20, 30, YELLOW)
	display.FillRectangle(10, 110, w-20, 30, WHITE)

	display.DrawFastHLine(0, w-1, 150, CYAN)
	display.DrawFastVLine(w/2, 160, h-160, MAGENTA)

	time.Sleep(3 * time.Second)
}

func testRotations(display *st7789.Device) {
	println("Test: rotations")
	rotations := []drivers.Rotation{
		drivers.Rotation0, drivers.Rotation90,
		drivers.Rotation180, drivers.Rotation270,
	}
	for i, rot := range rotations {
		println("Rotation", i*90)
		display.SetRotation(rot)
		display.FillScreen(BLACK)
		w, h := display.Size()
		display.FillRectangle(0, 0, w/4, h/4, RED)
		display.FillRectangle(3*w/4, 3*h/4, w/4, h/4, GREEN)
		time.Sleep(2 * time.Second)
	}
	display.SetRotation(drivers.Rotation90)
}

func testPerformance(display *st7789.Device) {
	println("Test: performance")
	w, h := display.Size()
	start := time.Now()
	for i := 0; i < 10; i++ {
		display.FillScreen(RED)
		display.FillScreen(GREEN)
		display.FillScreen(BLUE)
	}
	elapsed := time.Since(start)
	println("30 full-screen fills:", elapsed.String())

	display.FillScreen(BLACK)
	start = time.Now()
	for i := int16(0); i < 100; i++ {
		x := (i * 3) % (w - 10)
		y := (i * 2) % (h - 10)
		display.FillRectangle(x, y, 10, 10, RED)
	}
	println("100 small rects:", time.Since(start).String())

	for {
		display.FillScreen(RED)
		time.Sleep(300 * time.Millisecond)
		display.FillScreen(GREEN)
		time.Sleep(300 * time.Millisecond)
		display.FillScreen(BLUE)
		time.Sleep(300 * time.Millisecond)
	}
}
