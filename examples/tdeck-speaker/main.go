// T-Deck speaker: перебор вариантов вывода — по логу понять, что сработало.
// Вариант 1 написан так, как будто в TinyGo есть machine.I2S (целевой API).
package main

import (
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

const (
	sampleRate = 16000
	testHz     = 1000
	testDur    = 2 * time.Second
)

var boardPower = machine.GPIO10

// I2SConfig — как могла бы выглядеть конфигурация I2S в machine (ESP32-S3).
type I2SConfig struct {
	BCK        machine.Pin
	WS         machine.Pin
	DOUT       machine.Pin
	SampleRate uint32
	Bits       uint8
}

// I2S — интерфейс вывода звука (целевой API для TinyGo).
// Когда в TinyGo появится machine.I2S, можно заменить i2sDriver на реальный драйвер.
type I2S interface {
	Configure(cfg I2SConfig) error
	Write(samples []int16) (n int, err error)
}

func main() {
	println("=== tdeck-speaker: перебор вариантов ===")
	time.Sleep(1 * time.Second)

	boardPower.Configure(machine.PinConfig{Mode: machine.PinOutput})
	boardPower.High()
	time.Sleep(200 * time.Millisecond)

	g22 := machine.Pin(tdeck.SpeakerPin)
	g22.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g22.Low()

	println("")
	println(">>> Вариант 1: как будто в TinyGo есть I2S — Configure + Write(samples)")
	var i2s I2S = newI2SDriver()
	cfg := I2SConfig{
		BCK:        machine.Pin(tdeck.I2SBCK),
		WS:         machine.Pin(tdeck.I2SWS),
		DOUT:       machine.Pin(tdeck.I2SDOUT),
		SampleRate: sampleRate,
		Bits:       16,
	}
	if err := i2s.Configure(cfg); err != nil {
		println("  I2S Configure:", err)
	} else {
		samples := makeBeepSamples(testHz, testDur)
		g22.High()
		time.Sleep(50 * time.Millisecond)
		n, err := i2s.Write(samples)
		g22.Low()
		if err != nil {
			println("  I2S Write err:", err)
		} else {
			println("  I2S Write ok, samples:", n)
		}
	}
	time.Sleep(500 * time.Millisecond)

	println(">>> Вариант 2: только GPIO22, меандр", testHz, "Hz,", testDur)
	beepGPIO(g22, testHz, testDur)
	time.Sleep(500 * time.Millisecond)

	println(">>> Вариант 3: только GPIO22, PWM", testHz, "Hz,", testDur)
	beepPWM(g22, testHz, testDur)
	time.Sleep(500 * time.Millisecond)

	println(">>> Вариант 4: GPIO22=High + I2S bit-bang 16bit")
	g22.High()
	time.Sleep(50 * time.Millisecond)
	bb := newI2SOut(machine.Pin(tdeck.I2SWS), machine.Pin(tdeck.I2SBCK), machine.Pin(tdeck.I2SDOUT), false)
	bb.configure()
	beepI2S(bb, testHz, testDur)
	g22.Low()
	time.Sleep(500 * time.Millisecond)

	println(">>> Вариант 5: GPIO22=High + I2S bit-bang 32bit slot")
	g22.High()
	time.Sleep(50 * time.Millisecond)
	bb32 := newI2SOut(machine.Pin(tdeck.I2SWS), machine.Pin(tdeck.I2SBCK), machine.Pin(tdeck.I2SDOUT), true)
	bb32.configure()
	beepI2S(bb32, testHz, testDur)
	g22.Low()
	time.Sleep(500 * time.Millisecond)

	println(">>> Вариант 6: I2S без GPIO22")
	bbNoEn := newI2SOut(machine.Pin(tdeck.I2SWS), machine.Pin(tdeck.I2SBCK), machine.Pin(tdeck.I2SDOUT), false)
	bbNoEn.configure()
	beepI2S(bbNoEn, testHz, testDur)
	time.Sleep(500 * time.Millisecond)

	println("")
	println("=== Конец. Если ни один вариант не запищал — спикер ждёт аппаратный I2S; в TinyGo для ESP32-S3 его нет, нужно добавлять. ===")
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
		if (i%period) < period/2 {
			v = amp
		} else {
			v = -amp
		}
		s[i*2] = v
		s[i*2+1] = v
	}
	return s
}

type i2sDriver struct {
	*i2sOut
}

func newI2SDriver() *i2sDriver {
	return &i2sDriver{}
}

func (d *i2sDriver) Configure(cfg I2SConfig) error {
	d.i2sOut = newI2SOut(cfg.WS, cfg.BCK, cfg.DOUT, false)
	d.configure()
	return nil
}

func (d *i2sDriver) Write(samples []int16) (int, error) {
	if d.i2sOut == nil {
		return 0, nil
	}
	slot := time.Second / time.Duration(sampleRate)
	for i := 0; i < len(samples); i += 2 {
		start := time.Now()
		l, r := samples[i], samples[i]
		if i+1 < len(samples) {
			r = samples[i+1]
		}
		d.writeSample(l, r)
		elapsed := time.Since(start)
		if elapsed < slot {
			time.Sleep(slot - elapsed)
		}
	}
	return len(samples), nil
}

type i2sOut struct {
	ws, bck, dout machine.Pin
	halfBit       time.Duration
	slot32        bool
}

func newI2SOut(ws, bck, dout machine.Pin, slot32 bool) *i2sOut {
	return &i2sOut{
		ws: ws, bck: bck, dout: dout,
		halfBit: time.Microsecond,
		slot32:  slot32,
	}
}

func (i *i2sOut) configure() {
	i.ws.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i.bck.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i.dout.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i.ws.Low()
	i.bck.Low()
	i.dout.Low()
}

func (i *i2sOut) writeBit(bit bool) {
	if bit {
		i.dout.High()
	} else {
		i.dout.Low()
	}
	time.Sleep(i.halfBit)
	i.bck.High()
	time.Sleep(i.halfBit)
	i.bck.Low()
}

func (i *i2sOut) write16(v int16) {
	for b := 15; b >= 0; b-- {
		i.writeBit((uint16(v)>>b)&1 != 0)
	}
}

func (i *i2sOut) writeSample(left, right int16) {
	i.ws.Low()
	i.write16(left)
	if i.slot32 {
		for n := 0; n < 16; n++ {
			i.writeBit(false)
		}
	}
	i.ws.High()
	i.write16(right)
	if i.slot32 {
		for n := 0; n < 16; n++ {
			i.writeBit(false)
		}
	}
}

func beepI2S(i2s *i2sOut, hz int, dur time.Duration) {
	samples := int(sampleRate * dur.Milliseconds() / 1000)
	period := sampleRate / hz
	if period < 2 {
		period = 2
	}
	amp := int16(0x7FFF)
	slot := time.Second / time.Duration(sampleRate)
	for n := 0; n < samples; n++ {
		start := time.Now()
		var v int16
		if (n % period) < period/2 {
			v = amp
		} else {
			v = -amp
		}
		i2s.writeSample(v, v)
		elapsed := time.Since(start)
		if elapsed < slot {
			time.Sleep(slot - elapsed)
		}
	}
}

func beepPWM(pin machine.Pin, hz int, dur time.Duration) {
	pwm := machine.PWM1
	pwm.Configure(machine.PWMConfig{Period: uint64(time.Second) / uint64(hz)})
	ch, err := pwm.Channel(pin)
	if err != nil {
		beepGPIO(pin, hz, dur)
		return
	}
	top := pwm.Top()
	pwm.Set(ch, top/2)
	time.Sleep(dur)
	pwm.Set(ch, 0)
}

func beepGPIO(pin machine.Pin, hz int, dur time.Duration) {
	half := time.Second / time.Duration(hz*2)
	deadline := time.Now().Add(dur)
	for time.Now().Before(deadline) {
		pin.High()
		time.Sleep(half)
		pin.Low()
		time.Sleep(half)
	}
}
