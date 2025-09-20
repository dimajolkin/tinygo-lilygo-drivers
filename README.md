# TinyGo LilyGo Drivers

TinyGo drivers for LilyGo devices.

📖 **[Documentation](https://dimajolkin.github.io/tinygo-lilygo-drivers/)**

## Supported Drivers

- **ST7789** - TFT display driver optimized for LilyGo T-Deck

## Installation

```bash
go get github.com/dimajolkin/tinygo-lilygo-drivers
```

## Quick Example

```go
package main

import (
    "image/color"
    "machine"
    "github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
)

func main() {
    spi := machine.SPI1
    spi.Configure(machine.SPIConfig{
        Frequency: 80000000,
        SCK: 40, SDO: 41, Mode: 0,
    })

    display := st7789.New(&spi, 10, 11, 12, 42)
    display.Configure(st7789.Config{
        Width: 240, Height: 320,
    })
    
    display.FillScreen(color.RGBA{255, 0, 0, 255})
}
```