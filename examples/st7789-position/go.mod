module st7789-position-example

go 1.25

require (
	github.com/dimajolkin/tinygo-lilygo-drivers v0.0.0
	tinygo.org/x/drivers v0.34.0
)

replace github.com/dimajolkin/tinygo-lilygo-drivers => ../..

require github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
