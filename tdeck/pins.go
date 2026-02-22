package tdeck

// Пины LilyGo T-Deck: https://github.com/Xinyuan-LilyGO/T-Deck (examples/UnitTest/utilities.h), https://www.espboards.dev/esp32/lilygo-t-deck
const (
	PinLeft  = 1  // G04 trackball
	PinUp    = 3  // G01 trackball
	PinRight = 2  // G02 trackball
	PinDown  = 15 // G03 trackball
	PinOK    = 0  // BOARD_BOOT_PIN

	SpeakerPin = 22 // IO22 в пинаутах; реальный звук в utilities.h — I2S ниже

	I2SWS   = 5  // BOARD_I2S_WS
	I2SBCK  = 7  // BOARD_I2S_BCK
	I2SDOUT = 6  // BOARD_I2S_DOUT (данные в DAC/усилитель)
	// ES7210 (audio codec): BOARD_ES7210_* in utilities.h
	ES7210_SCK  = 47 // BOARD_ES7210_SCK
	ES7210_DIN  = 14 // BOARD_ES7210_DIN
	ES7210_LRCK = 21 // BOARD_ES7210_LRCK
	ES7210_MCLK = 48 // BOARD_ES7210_MCLK
)
