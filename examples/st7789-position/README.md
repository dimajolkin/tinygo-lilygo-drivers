# ST7789 position / calibration

Firmware that draws a coordinate grid on the display so you can verify resolution and orientation (e.g. for LilyGo T-Deck).

- **Grid** every 40 px, axes at (0,0) in a distinct color
- **Coordinates** along X and Y (0, 40, 80, …)
- **Corners** marked with colored squares: yellow (0,0), red (w,0), green (0,h), blue (w,h)
- **Center** shows logical size `W x H`

If the picture looks wrong (e.g. portrait vs landscape, or wrong size), change `Width`, `Height`, and `Rotation` in `main.go` and rebuild. Serial prints `Size: W x H` at startup.

## Build and flash

```bash
cd examples/st7789-position
tinygo build -target=esp32-c3 -o position.uf2 .
tinygo flash -target=esp32-c3
```

Use the target that matches your board (e.g. `esp32-s3`).
