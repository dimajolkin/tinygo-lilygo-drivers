// T-Deck: список файлов с SD-карты (FAT). Вывод в UART.
package main

import (
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
	"tinygo.org/x/drivers/sdcard"
	"tinygo.org/x/tinyfs/fatfs"
)

const boardPower = machine.GPIO10

func main() {
	time.Sleep(1 * time.Second)

	boardPower.Configure(machine.PinConfig{Mode: machine.PinOutput})
	boardPower.High()
	time.Sleep(200 * time.Millisecond)

	spi := machine.SPI1
	spi.Configure(machine.SPIConfig{
		SCK: machine.Pin(tdeck.SDCardSCK),
		SDO: machine.Pin(tdeck.SDCardMOSI),
		SDI: machine.Pin(tdeck.SDCardMISO),
		// Frequency: 250_000,
		// Mode:      0,
	})

	sd := sdcard.New(spi, machine.Pin(tdeck.SDCardSCK), machine.Pin(tdeck.SDCardMOSI), machine.Pin(tdeck.SDCardMISO), machine.Pin(tdeck.SDCardCS))
	err := sd.Configure()
	if err != nil {
		println("SD Configure:", err.Error())
		for {
			time.Sleep(time.Hour)
		}
	}
	println("SD card OK, size:", sd.Size())

	fat := fatfs.New(&sd)
	fat.Configure(&fatfs.Config{SectorSize: 512})
	err = fat.Mount()
	if err != nil {
		println("FAT Mount:", err.Error())
		for {
			time.Sleep(time.Hour)
		}
	}
	defer fat.Unmount()
	println("FAT mounted")

	dir, err := fat.Open("/")
	if err != nil {
		println("Open /:", err.Error())
		for {
			time.Sleep(time.Hour)
		}
	}
	defer dir.Close()

	infos, err := dir.Readdir(0)
	if err != nil {
		println("Readdir:", err.Error())
		for {
			time.Sleep(time.Hour)
		}
	}

	println("--- Files on SD ---")
	for _, info := range infos {
		name := info.Name()
		if name == "" || name == "." || name == ".." {
			continue
		}
		var kind string
		if info.IsDir() {
			kind = "<DIR>"
		} else {
			kind = ""
		}
		println(name, info.Size(), kind)
	}
	println("--- Done ---")

	for {
		time.Sleep(time.Hour)
	}
}
