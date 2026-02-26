package tdeck

import (
	"machine"
	"strconv"
	"time"
)

const (
	adcBits  = 16
	adcMax   = (1 << adcBits) - 1
	adcRefMV = 3300
)

type BatteryConfig struct {
	EmptyMV   int32
	FullMV    int32
	ChargedMV int32
	Divider   int32
}

func DefaultBatteryConfig() BatteryConfig {
	return BatteryConfig{
		EmptyMV:   3000,
		FullMV:    3700,
		ChargedMV: 4200,
		Divider:   2,
	}
}

type BatteryReading struct {
	VoltageMV int32
	Pct       int
	Charging  bool
	TimeLeft  string
}

type Battery struct {
	adc        machine.ADC
	cfg        BatteryConfig
	lastVMV    int32
	lastAt     time.Time
}

func NewBattery(pin machine.Pin, cfg BatteryConfig) *Battery {
	return &Battery{
		adc:     machine.ADC{Pin: pin},
		cfg:     cfg,
		lastVMV: -1,
	}
}

func (b *Battery) Configure() {
	machine.InitADC()
	b.adc.Configure(machine.ADCConfig{})
}

func (b *Battery) Read() BatteryReading {
	var r BatteryReading
	raw := b.adc.Get()
	r.VoltageMV = int32(uint32(raw)*uint32(adcRefMV)*uint32(b.cfg.Divider)) / int32(adcMax)
	if r.VoltageMV < b.cfg.EmptyMV {
		r.VoltageMV = b.cfg.EmptyMV
	}
	if r.VoltageMV > b.cfg.ChargedMV {
		r.VoltageMV = b.cfg.ChargedMV
	}

	r.Charging = r.VoltageMV > b.cfg.FullMV+50

	if r.Charging {
		r.Pct = int((r.VoltageMV - b.cfg.FullMV) * 100 / (b.cfg.ChargedMV - b.cfg.FullMV))
	} else {
		r.Pct = int((r.VoltageMV - b.cfg.EmptyMV) * 100 / (b.cfg.FullMV - b.cfg.EmptyMV))
	}
	if r.Pct < 0 {
		r.Pct = 0
	}
	if r.Pct > 100 {
		r.Pct = 100
	}

	if r.Charging {
		now := time.Now()
		if b.lastVMV >= 0 && r.VoltageMV > b.lastVMV {
			elapsed := now.Sub(b.lastAt).Seconds()
			if elapsed >= 4 {
				rateMVps := float64(r.VoltageMV-b.lastVMV) / elapsed
				if rateMVps > 2 {
					remaining := int32(b.cfg.ChargedMV - r.VoltageMV)
					if remaining > 0 {
						secLeft := float64(remaining) / rateMVps
						r.TimeLeft = formatBatteryTimeLeft(int64(secLeft))
					}
				}
			}
		}
		b.lastVMV = r.VoltageMV
		b.lastAt = now
	} else {
		b.lastVMV = -1
	}

	return r
}

func formatBatteryTimeLeft(sec int64) string {
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
