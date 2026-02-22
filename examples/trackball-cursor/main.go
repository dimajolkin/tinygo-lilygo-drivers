package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
	"tinygo.org/x/drivers"
)

const (
	TFT_SCLK machine.Pin = 40
	TFT_MOSI machine.Pin = 41
	TFT_CS   machine.Pin = 12
	TFT_DC   machine.Pin = 11
	TFT_RST  machine.Pin = 10
	TFT_BL   machine.Pin = 42

	screenW   = 320
	screenH   = 240
	cursorW   = 10
	cursorH   = 10
	step      = 4
	edgeMargin = 2
)

var (
	bgColor     = color.RGBA{0x10, 0x10, 0x20, 255}
	cursorColor = color.RGBA{0xff, 0xff, 0xff, 255}
	clickColor  = color.RGBA{0xff, 0x80, 0x00, 255}
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

	blPWM := machine.PWM0
	blPWM.Configure(machine.PWMConfig{Period: uint64(time.Second / 5000)})
	display.ConfigureBacklightPWM(blPWM)
	display.SetBacklightBrightness(180)

	tb := tdeck.NewTrackballDefault()
	x := int16(screenW/2 - cursorW/2)
	y := int16(screenH/2 - cursorH/2)

	display.FillScreen(bgColor)
	display.FillRectangle(x, y, cursorW, cursorH, cursorColor)

	for {
		dx, dy := tb.ReadMotion() // крути трекбол — импульсы по пинам
		s := tb.Read()
		prevX, prevY := x, y
		x += int16(dx) * step
		y += int16(dy) * step
		if x < edgeMargin {
			x = edgeMargin
		}
		if x > screenW-cursorW-edgeMargin {
			x = screenW - cursorW - edgeMargin
		}
		if y < edgeMargin {
			y = edgeMargin
		}
		if y > screenH-cursorH-edgeMargin {
			y = screenH - cursorH - edgeMargin
		}

		if x != prevX || y != prevY {
			display.FillRectangle(prevX, prevY, cursorW, cursorH, bgColor)
		}
		c := cursorColor
		if s.OK {
			c = clickColor
		}
		display.FillRectangle(x, y, cursorW, cursorH, c)

		time.Sleep(10 * time.Millisecond)
	}
}
