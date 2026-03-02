# Snake

Classic Nokia-style Snake for LilyGo T-Deck. 320×240 display, controls via T-Deck keyboard and trackball, battery and brightness/sound indicators in the top panel.

## Requirements

- [TinyGo](https://tinygo.org/getting-started/install/)
- LilyGo T-Deck (ST7789 display + I2C keyboard + trackball + speaker)

## Build and flash

From the example directory:

```bash
cd examples/game-snake
tinygo build -target=esp32-s3-devkitc -o game-snake.uf2 .
```

Then flash the device (T-Deck in bootloader mode):

```bash
tinygo flash -target=esp32-s3-devkitc
```

## Controls

Keyboard:

| Key | Action |
|-----|--------|
| **W** | Up |
| **S** | Down |
| **A** | Left |
| **D** | Right |
| **Space** | Pause / resume |
| **R** or **Space** (after game over) | New game |
| **+ / -** (after game over) | Change backlight brightness |
| **key code 4** | Toggle sound on/off |

Trackball:

- Move the ball to change direction (dominant axis: horizontal/vertical).
- Press trackball button to restart after game over.

## Rules

- Eat the red food — the snake grows and new food appears.
- Don’t run into your own tail (screen wraps at the edges).
- After game over, use **Space / R / trackball press** to restart.
