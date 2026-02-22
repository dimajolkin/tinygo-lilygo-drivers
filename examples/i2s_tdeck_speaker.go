//go:build esp32s3

// T-Deck speaker: I2S 16 kHz, 16 bit. Если мелодия слишком быстрая — увеличь speedFactor.
// Pins: POWERON=10, BCK=7, DOUT=6, WS=5.
package main

import (
	"machine"
	"math"
	"time"
)

const (
	pinPowerOn = 10
	pinBCK     = 7
	pinDOUT    = 6
	pinWS      = 5

	hwSampleRate = 16000
	speedFactor  = 300
	bufSamples   = 1023
	amp          = 24000
)

var (
	stereoBuf [bufSamples]uint32
	sineTable [256]int16
)

type note struct {
	freqHz uint16
	ms     uint32
}

func init() {
	for i := 0; i < 256; i++ {
		sineTable[i] = int16(math.Sin(2*math.Pi*float64(i)/256) * float64(amp))
	}
}

func main() {
	powerOn := machine.Pin(pinPowerOn)
	powerOn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	powerOn.High()
	time.Sleep(1 * time.Second)

	machine.I2S0.Configure(machine.I2SConfig{
		SCK:            machine.Pin(pinBCK),
		WS:             machine.Pin(pinWS),
		SDO:            machine.Pin(pinDOUT),
		SDI:            machine.NoPin,
		Mode:           machine.I2SModeSource,
		Standard:       machine.I2StandardPhilips,
		ClockSource:    machine.I2SClockSourceInternal,
		DataFormat:     machine.I2SDataFormat16bit,
		AudioFrequency: hwSampleRate,
		Stereo:         true,
	})
	machine.I2S0.Enable(true)

	imperialMarch := []note{
		{196, 280}, {196, 280}, {196, 280}, {155, 600}, {233, 180},
		{196, 320}, {155, 560}, {233, 180}, {196, 720},
		{294, 280}, {294, 280}, {294, 280}, {311, 560}, {233, 180},
		{196, 320}, {155, 560}, {233, 180}, {196, 720},
		{0, 300},
	}

	const pauseBetweenNotes = 40 * time.Millisecond
	for {
		t0 := time.Now()
		for _, n := range imperialMarch {
			playNote(n.freqHz, n.ms)
			time.Sleep(pauseBetweenNotes)
		}
		elapsed := time.Since(t0)
		println("melody_ms:", elapsed.Milliseconds())
		time.Sleep(2 * time.Second)
	}
}

func playNote(freqHz uint16, ms uint32) {
	if freqHz == 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return
	}
	genRate := uint32(hwSampleRate) * uint32(speedFactor)
	phaseInc := uint32(freqHz) * 256 * 65536 / genRate
	phase := uint32(0)
	samplesLeft := ms * genRate / 1000
	const fadeSamples = 24
	for samplesLeft > 0 {
		n := bufSamples
		if samplesLeft < uint32(n) {
			n = int(samplesLeft)
		}
		n &^= 1
		if n == 0 {
			break
		}
		lastChunk := samplesLeft <= uint32(n)
		for i := 0; i < n; i++ {
			idx := (phase >> 16) & 0xff
			s := sineTable[idx]
			env := uint32(0xffff)
			if i < fadeSamples {
				env = uint32(i) * 0xffff / fadeSamples
			}
			if lastChunk && i >= n-fadeSamples && n > fadeSamples {
				e := uint32(n-i) * 0xffff / fadeSamples
				if e < env {
					env = e
				}
			}
			s = int16(int32(s) * int32(env) >> 16)
			u := uint32(uint16(s))
			stereoBuf[i] = (u << 16) | u
			phase += phaseInc
		}
		machine.I2S0.WriteStereo(stereoBuf[:n])
		samplesLeft -= uint32(n)
	}
}
