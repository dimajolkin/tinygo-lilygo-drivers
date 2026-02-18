package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
)

const (
	TFT_SCLK machine.Pin = 40
	TFT_MOSI machine.Pin = 41
	TFT_CS   machine.Pin = 12
	TFT_DC   machine.Pin = 11
	TFT_RST  machine.Pin = 10
	TFT_BL   machine.Pin = 42

	boardI2CSCL = machine.GPIO8
	boardI2CSDA = machine.GPIO18

	screenW  = 320
	screenH  = 240
	cellSz   = 16
	gridCols = screenW / cellSz
	gridRows = screenH / cellSz
	tickMs   = 150
)

var (
	nokiaBg    = color.RGBA{0x0d, 0x37, 0x0d, 255}
	nokiaSnake = color.RGBA{0x33, 0x99, 0x33, 255}
	nokiaHead  = color.RGBA{0x55, 0xbb, 0x55, 255}
	nokiaFood  = color.RGBA{0xcc, 0x22, 0x22, 255}
	nokiaGrid  = color.RGBA{0x1a, 0x4a, 0x1a, 255}
)

type vec2 struct{ x, y int }

type game struct {
	snake   []vec2
	dir     vec2
	next    vec2
	food    vec2
	score   int
	over    bool
	display *st7789.Device
}

func main() {
	time.Sleep(1 * time.Second)

	spi := machine.SPI1
	spi.Configure(machine.SPIConfig{
		Frequency: 80000000,
		SCK:       TFT_SCLK,
		SDO:       TFT_MOSI,
		Mode:      0,
	})

	display := st7789.New(spi, TFT_RST, TFT_DC, TFT_CS, TFT_BL)
	err := display.Configure(st7789.Config{
		Width:    240,
		Height:   320,
		Rotation: st7789.Rotation90,
	})
	if err != nil {
		println("display:", err.Error())
		return
	}
	display.EnableBacklight(true)

	err = machine.I2C0.Configure(machine.I2CConfig{SCL: boardI2CSCL, SDA: boardI2CSDA})
	if err != nil {
		println("i2c:", err.Error())
		return
	}

	kb := tdeck.New(machine.I2C0, TFT_RST)
	kb.PowerOn()
	time.Sleep(100 * time.Millisecond)
	_ = kb.SetBrightness(127)

	g := &game{display: display}
	g.reset()

	for {
		code, _ := kb.ReadKey()
		if code != 0 {
			if g.over && (code == ' ' || code == 'r' || code == 'R') {
				g.reset()
			} else {
				g.input(byte(code))
			}
		}

		if !g.over {
			g.tick()
		} else {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		g.draw()
		time.Sleep(tickMs * time.Millisecond)
	}
}

func (g *game) reset() {
	cx, cy := gridCols/2, gridRows/2
	g.snake = []vec2{{cx, cy}, {cx - 1, cy}, {cx - 2, cy}}
	g.dir = vec2{1, 0}
	g.next = g.dir
	g.score = 0
	g.over = false
	g.placeFood()
}

func (g *game) input(key byte) {
	switch key {
	case 'w', 'W':
		if g.dir.y == 0 {
			g.next = vec2{0, -1}
		}
	case 's', 'S':
		if g.dir.y == 0 {
			g.next = vec2{0, 1}
		}
	case 'a', 'A':
		if g.dir.x == 0 {
			g.next = vec2{-1, 0}
		}
	case 'd', 'D':
		if g.dir.x == 0 {
			g.next = vec2{1, 0}
		}
	}
}

func (g *game) placeFood() {
	for {
		g.food.x = int(randUint(uint32(gridCols)))
		g.food.y = int(randUint(uint32(gridRows)))
		ok := true
		for _, s := range g.snake {
			if s.x == g.food.x && s.y == g.food.y {
				ok = false
				break
			}
		}
		if ok {
			return
		}
	}
}

func (g *game) tick() {
	g.dir = g.next
	head := g.snake[0]
	head.x += g.dir.x
	head.y += g.dir.y

	if head.x < 0 || head.x >= gridCols || head.y < 0 || head.y >= gridRows {
		g.over = true
		return
	}
	for _, s := range g.snake {
		if s.x == head.x && s.y == head.y {
			g.over = true
			return
		}
	}

	g.snake = append([]vec2{head}, g.snake...)
	if head.x == g.food.x && head.y == g.food.y {
		g.score++
		g.placeFood()
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}
}

func (g *game) draw() {
	g.display.FillScreen(nokiaBg)

	for row := int16(0); row <= gridRows; row++ {
		g.display.FillRectangle(0, row*cellSz, screenW, 1, nokiaGrid)
	}
	for col := int16(0); col <= gridCols; col++ {
		g.display.FillRectangle(col*cellSz, 0, 1, screenH, nokiaGrid)
	}

	for i, s := range g.snake {
		x := int16(s.x) * cellSz
		y := int16(s.y) * cellSz
		c := nokiaSnake
		if i == 0 {
			c = nokiaHead
		}
		g.display.FillRectangle(x+2, y+2, cellSz-4, cellSz-4, c)
	}

	fx := int16(g.food.x) * cellSz
	fy := int16(g.food.y) * cellSz
	g.display.FillRectangle(fx+2, fy+2, cellSz-4, cellSz-4, nokiaFood)
}

var seed uint32 = 1

func randUint(max uint32) uint32 {
	seed = seed*1103515245 + 12345
	return (seed >> 16) % max
}
