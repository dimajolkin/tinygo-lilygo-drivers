# Snake

Classic Nokia-style Snake for LilyGo T-Deck. 320×240 display, controls via T-Deck keyboard.

## Requirements

- [TinyGo](https://tinygo.org/getting-started/install/)
- LilyGo T-Deck (ST7789 display + I2C keyboard)

## Build and flash

From the example directory:

```bash
cd examples/game-snake
tinygo build -target=esp32-c3 -o game-snake.uf2 .
```

Flash the device (T-Deck in bootloader mode):

```bash
tinygo flash -target=esp32-c3
```

Or copy `game-snake.uf2` onto the T-Deck drive that appears (BOOT/RENAME_ME etc. — see LilyGo docs).

> Target may differ by T-Deck model (e.g. `esp32-s3`). Check your board documentation.

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
