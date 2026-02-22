// T-Deck speaker: machine.I2S0, конфиг как в рабочем Arduino (arduino.md): моно, 16 kHz, I2S.
package main

import (
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

const (
	sampleRate = 16000
	beepHz     = 1000
	beepDur    = 2 * time.Second
)

var boardPower = machine.GPIO10

func main() {
	time.Sleep(1 * time.Second)

	boardPower.Configure(machine.PinConfig{Mode: machine.PinOutput})
	boardPower.High()
	time.Sleep(200 * time.Millisecond)

	spkEn := machine.Pin(tdeck.SpeakerPin)
	spkEn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	spkEn.Low()

	err := machine.I2S0.Configure(machine.I2SConfig{
		SCK:            machine.Pin(tdeck.I2SBCK),
		WS:             machine.Pin(tdeck.I2SWS),
		SDO:            machine.Pin(tdeck.I2SDOUT),
		SDI:            machine.NoPin,
		Mode:           machine.I2SModeSource,
		Standard:       machine.I2StandardPhilips,
		ClockSource:    machine.I2SClockSourceInternal,
		DataFormat:     machine.I2SDataFormat16bit,
		AudioFrequency: sampleRate,
		Stereo:         false,
	})
	if err != nil {
		println("I2S0 Configure:", err)
		for {
			time.Sleep(time.Hour)
		}
	}

	machine.I2S0.Enable(true)

	time.Sleep(1 * time.Second)

	samples := makeBeepSamples(beepHz, beepDur)
	mono := make([]uint16, len(samples)/2)
	for i := range mono {
		mono[i] = uint16(samples[i*2])
	}

	spkEn.High()
	time.Sleep(50 * time.Millisecond)
	n, err := machine.I2S0.WriteMono(mono)
	spkEn.Low()
	if err != nil {
		println("I2S0 WriteMono:", err)
	} else {
		println("I2S0 WriteMono ok, samples:", n)
	}

	for {
		time.Sleep(time.Hour)
	}
}

func makeBeepSamples(hz int, dur time.Duration) []int16 {
	n := int(sampleRate * dur.Milliseconds() / 1000)
	s := make([]int16, n*2)
	period := sampleRate / hz
	if period < 2 {
		period = 2
	}
	amp := int16(0x7FFF)
	for i := 0; i < n; i++ {
		var v int16
		if (i % period) < period/2 {
			v = amp
		} else {
			v = -amp
		}
		s[i*2] = v
		s[i*2+1] = v
	}
	return s
}
