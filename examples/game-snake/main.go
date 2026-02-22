package main

import (
	"image/color"
	"machine"
	"time"

	"github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
	"github.com/dimajolkin/tinygo-lilygo-drivers/tdeck"
	"tinygo.org/x/drivers"
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

	screenW     = 320
	screenH     = 240
	cellSz      = 16
	panelHeight = cellSz
	fieldY      = panelHeight
	fieldH      = screenH - panelHeight
	gridCols    = screenW / cellSz
	gridRows    = fieldH / cellSz
	tickMs      = 150

	speakerSampleRate = 16000
	beepHz            = 1000
	beepDur           = 80 * time.Millisecond
)

var (
	nokiaBg    = color.RGBA{0x0d, 0x37, 0x0d, 255}
	nokiaSnake = color.RGBA{0x33, 0x99, 0x33, 255}
	nokiaHead  = color.RGBA{0x55, 0xbb, 0x55, 255}
	nokiaFood  = color.RGBA{0xcc, 0x22, 0x22, 255}
	nokiaGrid  = color.RGBA{0x1a, 0x4a, 0x1a, 255}
	panelBg    = color.RGBA{0x08, 0x28, 0x08, 255}
	scoreColor = color.RGBA{0xaa, 0xcc, 0xaa, 255}
)

const cellPixels = cellSz * cellSz * 2

var (
	rgb565Bg    = rgbaTo565(nokiaBg)
	rgb565Snake = rgbaTo565(nokiaSnake)
	rgb565Head  = rgbaTo565(nokiaHead)
	rgb565Food  = rgbaTo565(nokiaFood)
	rgb565Grid  = rgbaTo565(nokiaGrid)
)

type vec2 struct{ x, y int }

type game struct {
	snake          []vec2
	dir            vec2
	next           vec2
	food           vec2
	score          int
	lastScore      int
	lastBrightness int
	brightness     uint8
	over           bool
	display        *st7789.Device
	needFullDraw   bool
	dirty          [8]vec2
	ndirty         int
	cellBuf        [cellPixels]uint8
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
	display.Configure(st7789.Config{
		Width:    240,
		Height:   320,
		Rotation: drivers.Rotation90,
	})

	blPWM := machine.PWM0
	blPWM.Configure(machine.PWMConfig{Period: uint64(time.Second / 5000)})
	display.ConfigureBacklightPWM(blPWM)

	display.SetBacklightBrightness(100)

	err := machine.I2C0.Configure(machine.I2CConfig{SCL: boardI2CSCL, SDA: boardI2CSDA})
	if err != nil {
		println("i2c:", err.Error())
		return
	}

	kb := tdeck.New(machine.I2C0, TFT_RST)
	kb.PowerOn()
	time.Sleep(100 * time.Millisecond)
	_ = kb.SetBrightness(127)

	tb := tdeck.NewTrackballDefault()

	initSpeaker()

	g := &game{display: &display, brightness: 128, lastBrightness: -1}
	g.reset()

	for {
		code, _ := kb.ReadKey()
		if code != 0 {
			if g.over && (code == '+' || code == '-') {
				step := uint8(25)
				if code == '+' && g.brightness < 255 {
					if 255-g.brightness < step {
						g.brightness = 255
					} else {
						g.brightness += step
					}
				} else if code == '-' && g.brightness > 0 {
					if g.brightness < step {
						g.brightness = 0
					} else {
						g.brightness -= step
					}
				}
				display.SetBacklightBrightness(g.brightness)
			} else if g.over && (code == ' ' || code == 'r' || code == 'R') {
				g.reset()
			} else if !g.over {
				g.input(byte(code))
			}
		}

		dx, dy := tb.ReadMotion()
		if dx != 0 || dy != 0 {
			g.inputTrackball(dx, dy)
		}
		if g.over {
			if s := tb.Read(); s.OK {
				g.reset()
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

func rgbaTo565(c color.RGBA) uint16 {
	r := uint16(c.R) >> 3
	gr := uint16(c.G) >> 2
	b := uint16(c.B) >> 3
	return (r << 11) | (gr << 5) | b
}

var (
	spkEn       machine.Pin
	speakerInit bool
)

func initSpeaker() {
	if speakerInit {
		return
	}
	spkEn = machine.Pin(tdeck.SpeakerPin)
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
		AudioFrequency: speakerSampleRate,
		Stereo:         false,
	})
	if err != nil {
		return
	}
	machine.I2S0.Enable(true)
	speakerInit = true
}

func playBeep() {
	if !speakerInit {
		return
	}
	samples := makeBeepSamples(beepHz, beepDur)
	mono := make([]uint16, len(samples)/2)
	for i := range mono {
		mono[i] = uint16(samples[i*2])
	}
	spkEn.High()
	time.Sleep(10 * time.Millisecond)
	_, _ = machine.I2S0.WriteMono(mono)
	spkEn.Low()
}

func makeBeepSamples(hz int, dur time.Duration) []int16 {
	n := int(speakerSampleRate * dur.Milliseconds() / 1000)
	s := make([]int16, n*2)
	period := speakerSampleRate / hz
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

func (g *game) reset() {
	cx, cy := gridCols/2, gridRows/2
	g.snake = []vec2{{cx, cy}, {cx - 1, cy}, {cx - 2, cy}}
	g.dir = vec2{1, 0}
	g.next = g.dir
	g.score = 0
	g.over = false
	g.placeFood()
	g.needFullDraw = true
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *game) inputTrackball(dx, dy int) {
	if abs(dx) >= abs(dy) {
		if dx > 0 && g.dir.x == 0 {
			g.next = vec2{1, 0}
		} else if dx < 0 && g.dir.x == 0 {
			g.next = vec2{-1, 0}
		}
	} else {
		if dy > 0 && g.dir.y == 0 {
			g.next = vec2{0, 1}
		} else if dy < 0 && g.dir.y == 0 {
			g.next = vec2{0, -1}
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

	g.ndirty = 0
	addDirty := func(v vec2) {
		if g.ndirty < len(g.dirty) {
			g.dirty[g.ndirty] = v
			g.ndirty++
		}
	}

	oldTail := g.snake[len(g.snake)-1]
	g.snake = append([]vec2{head}, g.snake...)
	if head.x == g.food.x && head.y == g.food.y {
		g.score++
		playBeep()
		oldFood := g.food
		g.placeFood()
		addDirty(vec2{head.x, head.y})
		addDirty(oldFood)
		addDirty(g.food)
	} else {
		g.snake = g.snake[:len(g.snake)-1]
		addDirty(oldTail)
		addDirty(vec2{head.x, head.y})
	}
}

func (g *game) draw() {
	if g.needFullDraw {
		g.drawFull()
		g.needFullDraw = false
		g.lastScore = g.score
		g.lastBrightness = int(g.brightness)
		return
	}
	if g.score != g.lastScore || int(g.brightness) != g.lastBrightness {
		g.drawPanel()
		g.lastScore = g.score
		g.lastBrightness = int(g.brightness)
	}
	for i := 0; i < g.ndirty; i++ {
		c := g.dirty[i]
		g.drawCell(c.x, c.y)
	}
}

func (g *game) drawPanel() {
	g.display.FillRectangle(0, 0, screenW, panelHeight, panelBg)
	g.display.FillRectangle(0, panelHeight-1, screenW, 1, nokiaGrid)
	drawScore(g.display, g.score)
	drawBrightness(g.display, g.brightness)
}

func (g *game) drawFull() {
	g.display.FillRectangle(0, 0, screenW, panelHeight, panelBg)
	g.display.FillRectangle(0, panelHeight-1, screenW, 1, nokiaGrid)
	drawScore(g.display, g.score)
	drawBrightness(g.display, g.brightness)
	g.display.FillRectangle(0, fieldY, screenW, fieldH, nokiaBg)
	for row := int16(0); row <= gridRows; row++ {
		g.display.FillRectangle(0, fieldY+row*cellSz, screenW, 1, nokiaGrid)
	}
	for col := int16(0); col <= gridCols; col++ {
		g.display.FillRectangle(col*cellSz, fieldY, 1, fieldH, nokiaGrid)
	}
	for i, s := range g.snake {
		x := int16(s.x) * cellSz
		y := fieldY + int16(s.y)*cellSz
		c := nokiaSnake
		if i == 0 {
			c = nokiaHead
		}
		g.display.FillRectangle(x+2, y+2, cellSz-4, cellSz-4, c)
	}
	fx := int16(g.food.x) * cellSz
	fy := fieldY + int16(g.food.y)*cellSz
	g.display.FillRectangle(fx+2, fy+2, cellSz-4, cellSz-4, nokiaFood)
}

var digitGlyphs = [10][6]uint8{
	{0x0E, 0x09, 0x09, 0x09, 0x09, 0x0E},
	{0x04, 0x0C, 0x04, 0x04, 0x04, 0x0E},
	{0x0E, 0x01, 0x0E, 0x08, 0x08, 0x0E},
	{0x0E, 0x01, 0x06, 0x01, 0x01, 0x0E},
	{0x09, 0x09, 0x0F, 0x01, 0x01, 0x01},
	{0x0F, 0x08, 0x0E, 0x01, 0x01, 0x0E},
	{0x0E, 0x08, 0x0E, 0x09, 0x09, 0x0E},
	{0x0F, 0x01, 0x02, 0x04, 0x04, 0x04},
	{0x0E, 0x09, 0x0E, 0x09, 0x09, 0x0E},
	{0x0E, 0x09, 0x09, 0x0F, 0x01, 0x0E},
}

const digitW, digitH = 4, 6

func drawScore(d *st7789.Device, score int) {
	drawNumberAt(d, 4, 5, score, scoreColor)
}

func drawBrightness(d *st7789.Device, b uint8) {
	pct := int(b) * 100 / 255
	digits := 1
	if pct >= 100 {
		digits = 3
	} else if pct >= 10 {
		digits = 2
	}
	x := screenW - int16((digitW+1)*digits) - 2
	drawNumberAt(d, x, 5, pct, scoreColor)
}

func drawNumberAt(d *st7789.Device, x, y int16, n int, c color.RGBA) {
	if n == 0 {
		drawDigitAt(d, x, y, 0, c)
		return
	}
	digits := []int{}
	for n > 0 {
		digits = append([]int{n % 10}, digits...)
		n /= 10
	}
	for i, dg := range digits {
		drawDigitAt(d, x+int16(i)*(digitW+1), y, dg, c)
	}
}

func drawDigitAt(d *st7789.Device, x, y int16, digit int, c color.RGBA) {
	if digit < 0 || digit > 9 {
		return
	}
	glyph := digitGlyphs[digit]
	for row := 0; row < digitH; row++ {
		bits := glyph[row]
		for col := 0; col < digitW; col++ {
			if (bits>>uint(4-1-col))&1 != 0 {
				d.FillRectangle(x+int16(col), y+int16(row), 1, 1, c)
			}
		}
	}
}

func (g *game) drawCell(cx, cy int) {
	buf := g.cellBuf[:]
	for i := 0; i < cellSz*cellSz; i++ {
		pix := rgb565Bg
		buf[i*2] = uint8(pix >> 8)
		buf[i*2+1] = uint8(pix)
	}
	for x := 0; x < cellSz; x++ {
		buf[x*2] = uint8(rgb565Grid >> 8)
		buf[x*2+1] = uint8(rgb565Grid)
	}
	for y := 1; y < cellSz; y++ {
		buf[(y*cellSz)*2] = uint8(rgb565Grid >> 8)
		buf[(y*cellSz)*2+1] = uint8(rgb565Grid)
	}
	for i, s := range g.snake {
		if s.x != cx || s.y != cy {
			continue
		}
		pix := rgb565Snake
		if i == 0 {
			pix = rgb565Head
		}
		for dy := 2; dy < cellSz-2; dy++ {
			for dx := 2; dx < cellSz-2; dx++ {
				off := (dy*cellSz + dx) * 2
				buf[off] = uint8(pix >> 8)
				buf[off+1] = uint8(pix)
			}
		}
		break
	}
	if g.food.x == cx && g.food.y == cy {
		for dy := 2; dy < cellSz-2; dy++ {
			for dx := 2; dx < cellSz-2; dx++ {
				off := (dy*cellSz + dx) * 2
				buf[off] = uint8(rgb565Food >> 8)
				buf[off+1] = uint8(rgb565Food)
			}
		}
	}
	px := int16(cx) * cellSz
	py := fieldY + int16(cy)*cellSz
	g.display.DrawRGBBitmap8(px, py, buf, cellSz, cellSz)
}

var seed uint32 = 1

func randUint(max uint32) uint32 {
	seed = seed*1103515245 + 12345
	return (seed >> 16) % max
}
