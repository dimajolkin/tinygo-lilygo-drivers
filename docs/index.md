# TinyGo LilyGo Drivers

TinyGo drivers for LilyGo devices.

## Installation

```bash
go get github.com/dimajolkin/tinygo-lilygo-drivers
```

## Usage

### Import specific drivers (recommended)

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

### Import main package

You can also import the main package to access version information:

```go
import "github.com/dimajolkin/tinygo-lilygo-drivers"

// Get library version
version := drivers.Version
```
