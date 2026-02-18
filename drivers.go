// Package drivers provides TinyGo drivers for LilyGo devices.
//
// This package serves as the main entry point for the tinygo-lilygo-drivers library.
// Individual drivers are available as subpackages:
//
//   - st7789: TFT display driver optimized for LilyGo devices
//   - tdeck:  T-Deck keyboard (I2C), backlight, key codes
//
// Example usage:
//
//	import "github.com/dimajolkin/tinygo-lilygo-drivers/st7789"
//
//	// Configure SPI and create display
//	display := st7789.New(&spi, resetPin, dcPin, csPin, blPin)
//	display.Configure(st7789.Config{Width: 240, Height: 320})
package drivers

// Version represents the current version of the drivers package
const Version = "v0.1.0"
