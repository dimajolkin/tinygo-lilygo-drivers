module tdeck-sdcard

go 1.25

replace github.com/dimajolkin/tinygo-lilygo-drivers => ../..

require (
	github.com/dimajolkin/tinygo-lilygo-drivers v0.0.0-00010101000000-000000000000
	tinygo.org/x/drivers v0.34.0
	tinygo.org/x/tinyfs v0.3.0
)
