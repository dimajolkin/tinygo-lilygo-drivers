// Package tdeck реализует драйвер клавиатуры LilyGo T-Deck (I2C).
// Чтение кодов символов клавиш, управление подсветкой и питанием.
package tdeck

import (
	"machine"
	"time"

	drivers "github.com/dimajolkin/tinygo-lilygo-drivers"
)

const (
	DefaultAddress = 0x55

	regBrightness        = 0x01
	regDefaultBrightness = 0x02
	minDefaultBrightness = 30
)

type Keyboard struct {
	bus      drivers.I2C
	addr     uint16
	powerPin machine.Pin
	readBuf  [1]byte
}

func New(bus drivers.I2C, powerPin machine.Pin) *Keyboard {
	return NewWithAddress(bus, DefaultAddress, powerPin)
}

func NewWithAddress(bus drivers.I2C, addr uint16, powerPin machine.Pin) *Keyboard {
	if powerPin != machine.NoPin {
		powerPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		powerPin.Low()
	}
	return &Keyboard{
		bus:      bus,
		addr:     addr,
		powerPin: powerPin,
	}
}

func (k *Keyboard) PowerOn() {
	if k.powerPin != machine.NoPin {
		k.powerPin.High()
		time.Sleep(500 * time.Millisecond)
	}
}

func (k *Keyboard) PowerOff() {
	if k.powerPin != machine.NoPin {
		k.powerPin.Low()
	}
}

func (k *Keyboard) SetBrightness(value uint8) error {
	return k.bus.Tx(k.addr, []byte{regBrightness, value}, nil)
}

func (k *Keyboard) SetDefaultBrightness(value uint8) error {
	if value < minDefaultBrightness {
		value = minDefaultBrightness
	}
	return k.bus.Tx(k.addr, []byte{regDefaultBrightness, value}, nil)
}

var emptyWrite = []byte{}

func (k *Keyboard) ReadKey() (byte, error) {
	k.readBuf[0] = 0
	err := k.bus.Tx(k.addr, emptyWrite, k.readBuf[:])
	if err != nil {
		return 0, err
	}
	return k.readBuf[0], nil
}
