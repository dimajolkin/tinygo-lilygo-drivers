# Snake

Classic Nokia-style Snake for LilyGo T-Deck. 320×240 display, controls via T-Deck keyboard.

## Requirements

- [TinyGo](https://tinygo.org/getting-started/install/)
- LilyGo T-Deck (ST7789 display + I2C keyboard)

## Build and flash

From the example directory:

```bash
cd examples/game-snake
tinygo build -target=esp32s3-wroom1 -o game-snake.uf2 .
```

Flash the device (T-Deck in bootloader mode):

```bash
tinygo flash -target=esp32s3-wroom1
```


## Controls

| Key | Action |
|-----|--------|
| **W** | Up |
| **S** | Down |
| **A** | Left |
| **D** | Right |
| **Space** or **R** | New game (after game over) |

## Rules

- Eat the red food — the snake grows and new food appears.
- Don’t hit the walls or your own tail.
- After game over, press Space or R to restart.
