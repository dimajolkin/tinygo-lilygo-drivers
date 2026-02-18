package main

import (
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

const (
	boardPowerOn = machine.GPIO10
	boardI2CSCL  = machine.GPIO8
	boardI2CSDA  = machine.GPIO18
)

func main() {
	time.Sleep(2 * time.Second)

	i2c := machine.I2C0
	err := i2c.Configure(machine.I2CConfig{
		SCL: boardI2CSCL,
		SDA: boardI2CSDA,
	})
	if err != nil {
		println("I2C configure:", err)
		return
	}

	kb := tdeck.New(i2c, boardPowerOn)
	kb.PowerOn()
	time.Sleep(100 * time.Millisecond)

	_ = kb.SetBrightness(127)
	time.Sleep(100 * time.Millisecond)

	println("Polling keys...")
	for {
		code, err := kb.ReadKey()
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if code != 0 {
			println("key:", code, "('", string(rune(code)), "')")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
