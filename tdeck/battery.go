package tdeck

import (
	"machine"
	"strconv"
	"time"
)

const (
	adcBits         = 16
	adcMax          = (1 << adcBits) - 1 // на ESP32-S3 в TinyGo сырые значения 0..65535 (наблюдаемо до ~65520)
	adcRefMV        = 3300
	smoothAlpha     = 0.12 // при разряде — меньше дрожания
	smoothAlphaChg  = 0.45 // при зарядке — быстрее виден рост напряжения
	rawAdcChgEnter  = 62000 // зарядка: сырой АЦП достиг этого значения (близко к макс 65520)
	rawAdcChgExit   = 55000 // выход из зарядки: сырой АЦП ниже этого (гистерезис)
)

type SoCMethod int

const (
	SoCLinear SoCMethod = iota
	SoCLiIon
)

type BatteryConfig struct {
	EmptyMV   int32
	FullMV    int32
	ChargedMV int32
	Divider   int32
	SoC       SoCMethod
}

func DefaultBatteryConfig() BatteryConfig {
	return BatteryConfig{
		EmptyMV:   3000,
		FullMV:    3700,
		ChargedMV: 4200,
		Divider:   2,
		SoC:       SoCLiIon,
	}
}

// Типовая разрядная кривая Li-ion (одна ячейка), напряжение мВ -> условный % (0–100 для 3000–4200 мВ).
var liIonCurve = []struct {
	mv  int32
	pct int
}{
	{3000, 0}, {3300, 2}, {3400, 5}, {3520, 10}, {3600, 15}, {3640, 20},
	{3680, 30}, {3720, 40}, {3780, 50}, {3850, 60}, {3920, 70}, {4000, 80},
	{4080, 90}, {4150, 97}, {4200, 100},
}

func liIonCurvePct(mv int32) int {
	if mv <= liIonCurve[0].mv {
		return 0
	}
	if mv >= liIonCurve[len(liIonCurve)-1].mv {
		return 100
	}
	for i := 0; i < len(liIonCurve)-1; i++ {
		v0, p0 := liIonCurve[i].mv, liIonCurve[i].pct
		v1, p1 := liIonCurve[i+1].mv, liIonCurve[i+1].pct
		if mv >= v0 && mv <= v1 {
			dx := v1 - v0
			if dx <= 0 {
				return p0
			}
			return p0 + (p1-p0)*int(mv-v0)/int(dx)
		}
	}
	return 0
}

func voltageToPctLiIon(mv int32, emptyMV, fullMV int32) int {
	if mv <= emptyMV {
		return 0
	}
	if mv >= fullMV {
		return 100
	}
	emptyPct := liIonCurvePct(emptyMV)
	fullPct := liIonCurvePct(fullMV)
	if fullPct <= emptyPct {
		return int((mv - emptyMV) * 100 / (fullMV - emptyMV))
	}
	pct := (liIonCurvePct(mv)-emptyPct)*100 / (fullPct - emptyPct)
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

type BatteryReading struct {
	VoltageMV int32  // сглаженное напряжение, мВ
	RawADC    uint32 // сырое значение АЦП (0..65535)
	Pct       int
	Charging  bool
	TimeLeft  string
}

type Battery struct {
	adc         machine.ADC
	cfg         BatteryConfig
	lastVMV     int32
	lastAt      time.Time
	smoothedVMV float64
	charging    bool // зарядка только по сырому АЦП: raw >= rawAdcChgEnter
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
	r.RawADC = uint32(raw)
	rawVMV := int32(uint32(raw)*uint32(adcRefMV)*uint32(b.cfg.Divider)) / int32(adcMax)
	if rawVMV < b.cfg.EmptyMV {
		rawVMV = b.cfg.EmptyMV
	}
	if rawVMV > b.cfg.ChargedMV {
		rawVMV = b.cfg.ChargedMV
	}

	if r.RawADC >= rawAdcChgEnter {
		b.charging = true
	} else if r.RawADC < rawAdcChgExit {
		b.charging = false
	}
	r.Charging = b.charging

	alpha := smoothAlpha
	if r.Charging {
		alpha = smoothAlphaChg
	}
	v := float64(rawVMV)
	if b.smoothedVMV == 0 {
		b.smoothedVMV = v
	} else {
		b.smoothedVMV = alpha*v + (1-alpha)*b.smoothedVMV
	}
	r.VoltageMV = int32(b.smoothedVMV)

	if r.Charging {
		if rawVMV >= b.cfg.ChargedMV {
			r.Pct = 100
		} else {
			r.Pct = int((rawVMV - b.cfg.FullMV) * 100 / (b.cfg.ChargedMV - b.cfg.FullMV))
		}
	} else {
		switch b.cfg.SoC {
		case SoCLiIon:
			r.Pct = voltageToPctLiIon(r.VoltageMV, b.cfg.EmptyMV, b.cfg.FullMV)
		default:
			r.Pct = int((r.VoltageMV - b.cfg.EmptyMV) * 100 / (b.cfg.FullMV - b.cfg.EmptyMV))
		}
	}
	if r.Pct < 0 {
		r.Pct = 0
	}
	if r.Pct > 100 {
		r.Pct = 100
	}

	if r.Charging {
		now := time.Now()
		if b.lastVMV >= 0 && rawVMV > b.lastVMV {
			elapsed := now.Sub(b.lastAt).Seconds()
			if elapsed >= 4 {
				rateMVps := float64(rawVMV-b.lastVMV) / elapsed
				if rateMVps > 2 {
					remaining := int32(b.cfg.ChargedMV - rawVMV)
					if remaining > 0 {
						secLeft := float64(remaining) / rateMVps
						r.TimeLeft = formatBatteryTimeLeft(int64(secLeft))
					}
				}
			}
		}
		b.lastVMV = rawVMV
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
