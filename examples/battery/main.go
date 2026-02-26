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

	batteryADC = machine.GPIO4
	fontScale  = 2
	margin     = 10
	batteryMAh = 1400
)

const (
	adcBits       = 16
	adcMax        = (1 << adcBits) - 1
	adcRefMV      = 3300
	vBatEmptyMV   = 3000
	vBatFullMV    = 3700
	vBatChargedMV = 4200
	dividerRatio  = 2
)

var (
	boardPower    = machine.GPIO10
	bgColor       = color.RGBA{0x10, 0x10, 0x18, 255}
	textColor     = color.RGBA{0xe0, 0xe0, 0xe0, 255}
	barBgColor    = color.RGBA{0x30, 0x30, 0x40, 255}
	barFgColor    = color.RGBA{0x50, 0xa0, 0xff, 255}
	lastVBatMV    int32 = -1
	lastChargeAt  time.Time
	lastPct       int   = -1
	lastVbatMV    int32 = -1
	lastCharging  bool
	lastTimeLeft  string
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

	machine.InitADC()
	adc := machine.ADC{Pin: batteryADC}
	adc.Configure(machine.ADCConfig{})

	for {
		raw := adc.Get()
		vBatMV := int32(uint32(raw)*uint32(adcRefMV)*dividerRatio) / int32(adcMax)
		if vBatMV < vBatEmptyMV {
			vBatMV = vBatEmptyMV
		}
		if vBatMV > vBatChargedMV {
			vBatMV = vBatChargedMV
		}

		charging := vBatMV > vBatFullMV+50

		var pct int
		if charging {
			pct = int((vBatMV - vBatFullMV) * 100 / (vBatChargedMV - vBatFullMV))
		} else {
			pct = int((vBatMV - vBatEmptyMV) * 100 / (vBatFullMV - vBatEmptyMV))
		}
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}

		timeLeftStr := ""
		if charging {
			now := time.Now()
			if lastVBatMV >= 0 && vBatMV > lastVBatMV {
				elapsed := now.Sub(lastChargeAt).Seconds()
				if elapsed >= 4 {
					rateMVps := float64(vBatMV-lastVBatMV) / elapsed
					if rateMVps > 2 {
						remainingMV := int32(vBatChargedMV - vBatMV)
						if remainingMV > 0 {
							secLeft := float64(remainingMV) / rateMVps
							timeLeftStr = formatTimeLeft(int64(secLeft))
						}
					}
				}
			}
			lastVBatMV = vBatMV
			lastChargeAt = now
		} else {
			lastVBatMV = -1
		}

		changed := lastPct != pct || lastVbatMV != vBatMV || lastCharging != charging || lastTimeLeft != timeLeftStr
		if !changed {
			time.Sleep(2 * time.Second)
			continue
		}
		lastPct = pct
		lastVbatMV = vBatMV
		lastCharging = charging
		lastTimeLeft = timeLeftStr

		display.FillScreen(bgColor)

		display.DrawString(margin, margin, "Battery "+strconv.Itoa(batteryMAh)+" mAh", textColor, fontScale)
		if charging {
			display.DrawString(margin, margin+30, "Charge: "+strconv.Itoa(pct)+"%", textColor, fontScale)
		} else {
			display.DrawString(margin, margin+30, strconv.Itoa(pct)+"%", textColor, fontScale)
		}
		display.DrawString(margin, margin+60, "Voltage: "+formatVoltage(vBatMV)+" V", textColor, fontScale)

		barX := int16(margin)
		barY := int16(margin + 90)
		barW := int16(screenW - 2*margin)
		barH := int16(24)
		display.FillRectangle(barX, barY, barW, barH, barBgColor)
		display.FillRectangle(barX+2, barY+2, int16(pct)*(barW-4)/100, barH-4, barFgColor)

		if charging {
			display.DrawString(margin, margin+130, "CHARGING", textColor, fontScale)
			if pct >= 100 {
				display.DrawString(margin, margin+158, "Full", textColor, fontScale)
			} else if timeLeftStr != "" {
				display.DrawString(margin, margin+158, "~"+timeLeftStr+" left", textColor, fontScale)
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

func formatTimeLeft(sec int64) string {
	if sec < 60 {
		return "<1 min"
	}
	min := sec / 60
	if min < 60 {
		return strconv.FormatInt(min, 10) + " min"
	}
	h := min / 60
	m := min % 60
	if m == 0 {
		return strconv.FormatInt(h, 10) + " h"
	}
	return strconv.FormatInt(h, 10) + " h " + strconv.FormatInt(m, 10) + " min"
}
