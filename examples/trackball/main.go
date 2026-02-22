package main

import (
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

func main() {
	time.Sleep(1 * time.Second)

	tb := tdeck.NewTrackballDefault()

	println("Trackball: крути шарик — вывод dx, dy. Нажатие OK = клик.")
	for {
		dx, dy := tb.ReadMotion()
		s := tb.Read()
		if dx != 0 || dy != 0 {
			println("motion: dx=", dx, " dy=", dy)
		}
		if s.OK {
			println("OK pressed")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
