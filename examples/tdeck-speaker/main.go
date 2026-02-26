// T-Deck speaker: I2S, проигрывание имперского марша.
package main

import (
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

const (
	sampleRate = 16000
	noteGapMs  = 300
)

var (
	imperialMarchTones = []int{
		392, 392, 392, 311, 466, 392, 311, 466, 392,
		587, 587, 587, 622, 466, 369, 311, 466, 392,
		784, 392, 392, 784, 739, 698, 659, 622, 659,
		415, 554, 523, 493, 466, 440, 466,
		311, 369, 311, 466, 392,
	}
	imperialMarchDurations = []int{
		350, 350, 350, 250, 100, 350, 250, 100, 700,
		350, 350, 350, 250, 100, 350, 250, 100, 700,
		350, 250, 100, 350, 250, 100, 100, 100, 450,
		150, 350, 250, 100, 100, 100, 450,
		150, 350, 250, 100, 750,
	}
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

	spkEn.High()
	time.Sleep(50 * time.Millisecond)

	for {
		for i := 0; i < len(imperialMarchTones); i++ {
			dur := time.Duration(imperialMarchDurations[i]) * time.Millisecond
			samples := makeBeepSamples(imperialMarchTones[i], dur)
			mono := make([]uint16, len(samples)/2)
			for j := range mono {
				mono[j] = uint16(samples[j*2])
			}
			_, _ = machine.I2S0.WriteMono(mono)
			time.Sleep(dur)
		}
		time.Sleep(2 * time.Second)
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
