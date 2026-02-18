package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
)

const (
	TFT_SCLK machine.Pin = 40
	TFT_MOSI machine.Pin = 41
	TFT_CS   machine.Pin = 12
	TFT_DC   machine.Pin = 11
	TFT_RST  machine.Pin = 10
	TFT_BL   machine.Pin = 42

	gridStep = 40
)

var (
	bgColor   = color.RGBA{20, 20, 30, 255}
	gridColor = color.RGBA{60, 60, 80, 255}
	axisColor = color.RGBA{100, 100, 255, 255}
	textColor = color.RGBA{220, 220, 255, 255}
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
	// Rotation90 = landscape 320x240, (0,0) top-left; corners: yellow TL, red TR, green BL, blue BR
	err := display.Configure(st7789.Config{
		Width:    240,
		Height:   320,
		Rotation: st7789.Rotation90,
	})
	if err != nil {
		println("display:", err.Error())
		return
	}
	display.EnableBacklight(true)

	w, h := display.Size()
	println("Size:", int(w), "x", int(h))

	display.FillScreen(bgColor)

	for x := int16(0); x <= w; x += gridStep {
		c := gridColor
		if x == 0 {
			c = axisColor
		}
		display.DrawFastVLine(x, 0, h-1, c)
		if x < w {
			drawNumber(display, x+2, 2, int(x), textColor)
		}
	}
	for y := int16(0); y <= h; y += gridStep {
		c := gridColor
		if y == 0 {
			c = axisColor
		}
		display.DrawFastHLine(0, w-1, y, c)
		if y < h {
			drawNumber(display, 2, y+2, int(y), textColor)
		}
	}

	display.FillRectangle(0, 0, 4, 4, color.RGBA{255, 255, 0, 255})
	display.FillRectangle(w-4, 0, 4, 4, color.RGBA{255, 0, 0, 255})
	display.FillRectangle(0, h-4, 4, 4, color.RGBA{0, 255, 0, 255})
	display.FillRectangle(w-4, h-4, 4, 4, color.RGBA{0, 0, 255, 255})

	drawNumber(display, w/2-18, h/2-6, int(w), textColor)
	display.FillRectangle(w/2-2, h/2-4, 4, 8, textColor)
	drawNumber(display, w/2+4, h/2-6, int(h), textColor)

	for {
		time.Sleep(10 * time.Second)
	}
}

var digitGlyphs = [10][6]uint8{
	{0x0E, 0x09, 0x09, 0x09, 0x09, 0x0E},
	{0x04, 0x0C, 0x04, 0x04, 0x04, 0x0E},
	{0x0E, 0x01, 0x0E, 0x08, 0x08, 0x0E},
	{0x0E, 0x01, 0x06, 0x01, 0x01, 0x0E},
	{0x09, 0x09, 0x0F, 0x01, 0x01, 0x01},
	{0x0F, 0x08, 0x0E, 0x01, 0x01, 0x0E},
	{0x0E, 0x08, 0x0E, 0x09, 0x09, 0x0E},
	{0x0F, 0x01, 0x02, 0x04, 0x04, 0x04},
	{0x0E, 0x09, 0x0E, 0x09, 0x09, 0x0E},
	{0x0E, 0x09, 0x09, 0x0F, 0x01, 0x0E},
}

const digitW, digitH = 4, 6

func drawNumber(d *st7789.Device, x, y int16, n int, c color.RGBA) {
	if n == 0 {
		drawDigit(d, x, y, 0, c)
		return
	}
	digits := []int{}
	for n > 0 {
		digits = append([]int{n % 10}, digits...)
		n /= 10
	}
	for i, dg := range digits {
		drawDigit(d, x+int16(i)*(digitW+1), y, dg, c)
	}
}

func drawDigit(d *st7789.Device, x, y int16, digit int, c color.RGBA) {
	if digit < 0 || digit > 9 {
		return
	}
	glyph := digitGlyphs[digit]
	for row := 0; row < digitH; row++ {
		bits := glyph[row]
		for col := 0; col < digitW; col++ {
			if (bits>>uint(4-1-col))&1 != 0 {
				d.FillRectangle(x+int16(col), y+int16(row), 1, 1, c)
			}
		}
	}
}
