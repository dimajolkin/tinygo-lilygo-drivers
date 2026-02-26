package main

import (
	"image/color"
	"machine"
	"strconv"
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

	boardI2CSCL = machine.GPIO8
	boardI2CSDA = machine.GPIO18

	screenW = 320
	screenH = 240

	fontScale  = 2
	margin     = 10
	batteryMAh = 1400
)

var (
	boardPower   = machine.GPIO10
	bgColor      = color.RGBA{0x10, 0x10, 0x18, 255}
	textColor    = color.RGBA{0xe0, 0xe0, 0xe0, 255}
	barBgColor   = color.RGBA{0x30, 0x30, 0x40, 255}
	barFgColor   = color.RGBA{0x50, 0xa0, 0xff, 255}
	lastPct      int   = -1
	lastVbatMV   int32 = -1
	lastCharging bool
	lastTimeLeft string
)

func main() {
	time.Sleep(1 * time.Second)

	boardPower.Configure(machine.PinConfig{Mode: machine.PinOutput})
	boardPower.High()
	time.Sleep(200 * time.Millisecond)

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

	err := machine.I2C0.Configure(machine.I2CConfig{SCL: boardI2CSCL, SDA: boardI2CSDA})
	if err != nil {
		println("i2c:", err.Error())
		return
	}

	kb := tdeck.New(machine.I2C0, TFT_RST)
	kb.PowerOn()
	time.Sleep(100 * time.Millisecond)
	_ = kb.SetBrightness(127)

	bat := tdeck.NewBattery(tdeck.BatteryADCPin, tdeck.DefaultBatteryConfig())
	bat.Configure()

	for {
		r := bat.Read()

		changed := lastPct != r.Pct || lastVbatMV != r.VoltageMV || lastCharging != r.Charging || lastTimeLeft != r.TimeLeft
		if !changed {
			time.Sleep(2 * time.Second)
			continue
		}
		lastPct = r.Pct
		lastVbatMV = r.VoltageMV
		lastCharging = r.Charging
		lastTimeLeft = r.TimeLeft

		display.FillScreen(bgColor)

		display.DrawString(margin, margin, "Battery "+strconv.Itoa(batteryMAh)+" mAh", textColor, fontScale)
		if r.Charging {
			display.DrawString(margin, margin+30, "Charge: "+strconv.Itoa(r.Pct)+"%", textColor, fontScale)
		} else {
			display.DrawString(margin, margin+30, strconv.Itoa(r.Pct)+"%", textColor, fontScale)
		}
		display.DrawString(margin, margin+60, "Voltage: "+formatVoltage(r.VoltageMV)+" V", textColor, fontScale)

		barX := int16(margin)
		barY := int16(margin + 90)
		barW := int16(screenW - 2*margin)
		barH := int16(24)
		display.FillRectangle(barX, barY, barW, barH, barBgColor)
		display.FillRectangle(barX+2, barY+2, int16(r.Pct)*(barW-4)/100, barH-4, barFgColor)

		if r.Charging {
			display.DrawString(margin, margin+130, "CHARGING", textColor, fontScale)
			if r.Pct >= 100 {
				display.DrawString(margin, margin+158, "Full", textColor, fontScale)
			} else if r.TimeLeft != "" {
				display.DrawString(margin, margin+158, "~"+r.TimeLeft+" left", textColor, fontScale)
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func formatVoltage(mv int32) string {
	v := mv / 1000
	frac := (mv % 1000) / 100
	return strconv.Itoa(int(v)) + "." + strconv.Itoa(int(frac))
}
