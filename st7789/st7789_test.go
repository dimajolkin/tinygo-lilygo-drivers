package st7789

import (
	"image/color"
	"testing"

	drivers "github.com/dimajolkin/tinygo-lilygo-drivers"
)

// MockSPI реализует интерфейс drivers.SPI для тестирования
type MockSPI struct {
	transfers []uint8
}

func (m *MockSPI) Transfer(b byte) (byte, error) {
	m.transfers = append(m.transfers, b)
	return 0xFF, nil // Возвращаем фиктивное значение
}

func (m *MockSPI) Tx(w, r []byte) error {
	for _, b := range w {
		m.transfers = append(m.transfers, b)
	}
	return nil
}

// Проверка, что MockSPI реализует интерфейс drivers.SPI
var _ drivers.SPI = (*MockSPI)(nil)

// MockPin реализует интерфейс Pin для тестирования
type MockPin struct {
	state bool
	mode  uint8
}

func (m *MockPin) Configure(config interface{}) {}
func (m *MockPin) Set(high bool)                { m.state = high }
func (m *MockPin) High()                        { m.state = true }
func (m *MockPin) Low()                         { m.state = false }
func (m *MockPin) Get() bool                    { return m.state }

func TestNew(t *testing.T) {
	mockSPI := &MockSPI{}
	resetPin := &MockPin{}
	dcPin := &MockPin{}
	csPin := &MockPin{}
	blPin := &MockPin{}

	device := New(mockSPI, resetPin, dcPin, csPin, blPin)

	if device == nil {
		t.Fatal("New() returned nil")
	}

	if device.width != 240 {
		t.Errorf("Expected default width 240, got %d", device.width)
	}

	if device.height != 320 {
		t.Errorf("Expected default height 320, got %d", device.height)
	}

	if device.rotation != Rotation0 {
		t.Errorf("Expected default rotation 0, got %d", device.rotation)
	}

	// Проверяем, что буферы инициализированы
	if len(device.largeBuffer) != 4096 {
		t.Errorf("Expected large buffer size 4096, got %d", len(device.largeBuffer))
	}

	if device.colorCache == nil {
		t.Error("Color cache not initialized")
	}
}

func TestColorToRGB565(t *testing.T) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})

	tests := []struct {
		name     string
		color    color.RGBA
		expected uint16
	}{
		{
			name:     "Red",
			color:    color.RGBA{255, 0, 0, 255},
			expected: 0xF800, // 11111 000000 00000
		},
		{
			name:     "Green",
			color:    color.RGBA{0, 255, 0, 255},
			expected: 0x07E0, // 00000 111111 00000
		},
		{
			name:     "Blue",
			color:    color.RGBA{0, 0, 255, 255},
			expected: 0x001F, // 00000 000000 11111
		},
		{
			name:     "White",
			color:    color.RGBA{255, 255, 255, 255},
			expected: 0xFFFF, // 11111 111111 11111
		},
		{
			name:     "Black",
			color:    color.RGBA{0, 0, 0, 255},
			expected: 0x0000, // 00000 000000 00000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := device.colorToRGB565(tt.color)
			if result != tt.expected {
				t.Errorf("colorToRGB565(%+v) = 0x%04X, expected 0x%04X", tt.color, result, tt.expected)
			}
		})
	}
}

func TestColorCaching(t *testing.T) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})

	redColor := color.RGBA{255, 0, 0, 255}

	// Первый вызов должен вычислить и закешировать
	result1 := device.colorToRGB565(redColor)

	// Второй вызов должен взять из кеша
	result2 := device.colorToRGB565(redColor)

	if result1 != result2 {
		t.Errorf("Color caching failed: first call = 0x%04X, second call = 0x%04X", result1, result2)
	}

	// Проверяем, что цвет действительно в кеше
	key := (uint32(redColor.R) << 24) | (uint32(redColor.G) << 16) | (uint32(redColor.B) << 8) | uint32(redColor.A)
	if cached, exists := device.colorCache[key]; !exists || cached != result1 {
		t.Error("Color not properly cached")
	}
}

func TestSize(t *testing.T) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})

	tests := []struct {
		rotation    Rotation
		expectedW   int16
		expectedH   int16
		description string
	}{
		{Rotation0, 240, 320, "Portrait"},
		{Rotation90, 320, 240, "Landscape 90"},
		{Rotation180, 240, 320, "Portrait 180"},
		{Rotation270, 320, 240, "Landscape 270"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			device.rotation = tt.rotation
			w, h := device.Size()
			if w != tt.expectedW || h != tt.expectedH {
				t.Errorf("Size() with %s = (%d, %d), expected (%d, %d)",
					tt.description, w, h, tt.expectedW, tt.expectedH)
			}
		})
	}
}

func TestRotation(t *testing.T) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})

	// Тестируем getter
	if device.Rotation() != Rotation0 {
		t.Errorf("Initial rotation should be Rotation0, got %d", device.Rotation())
	}

	// Тестируем setter
	device.SetRotation(Rotation90)
	if device.Rotation() != Rotation90 {
		t.Errorf("After SetRotation(Rotation90), got %d", device.Rotation())
	}
}

func TestConfigureDefaultValues(t *testing.T) {
	mockSPI := &MockSPI{}
	resetPin := &MockPin{}
	dcPin := &MockPin{}
	csPin := &MockPin{}
	blPin := &MockPin{}

	device := New(mockSPI, resetPin, dcPin, csPin, blPin)

	// Тест с пустой конфигурацией (должны использоваться значения по умолчанию)
	err := device.Configure(Config{})
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	if device.width != 240 {
		t.Errorf("Expected default width 240, got %d", device.width)
	}

	if device.height != 320 {
		t.Errorf("Expected default height 320, got %d", device.height)
	}

	// Проверяем, что была выполнена последовательность инициализации
	if len(mockSPI.transfers) == 0 {
		t.Error("No SPI transfers during configuration")
	}

	// Проверяем, что подсветка включена
	if !blPin.state {
		t.Error("Backlight should be enabled after configuration")
	}
}

func TestConfigureCustomValues(t *testing.T) {
	mockSPI := &MockSPI{}
	resetPin := &MockPin{}
	dcPin := &MockPin{}
	csPin := &MockPin{}
	blPin := &MockPin{}

	device := New(mockSPI, resetPin, dcPin, csPin, blPin)

	// Тест с кастомными значениями
	config := Config{
		Width:    480,
		Height:   640,
		Rotation: Rotation90,
	}

	err := device.Configure(config)
	if err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	if device.width != 480 {
		t.Errorf("Expected width 480, got %d", device.width)
	}

	if device.height != 640 {
		t.Errorf("Expected height 640, got %d", device.height)
	}

	if device.rotation != Rotation90 {
		t.Errorf("Expected rotation Rotation90, got %d", device.rotation)
	}
}

func TestEnableBacklight(t *testing.T) {
	mockSPI := &MockSPI{}
	blPin := &MockPin{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, blPin)

	// Тест включения подсветки
	device.EnableBacklight(true)
	if !blPin.state {
		t.Error("Backlight should be enabled")
	}

	// Тест выключения подсветки
	device.EnableBacklight(false)
	if blPin.state {
		t.Error("Backlight should be disabled")
	}
}

// Benchmark тесты
func BenchmarkColorToRGB565(b *testing.B) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})
	testColor := color.RGBA{128, 128, 128, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device.colorToRGB565(testColor)
	}
}

func BenchmarkColorToRGB565Cached(b *testing.B) {
	mockSPI := &MockSPI{}
	device := New(mockSPI, &MockPin{}, &MockPin{}, &MockPin{}, &MockPin{})
	testColor := color.RGBA{128, 128, 128, 255}

	// Предварительно кешируем цвет
	device.colorToRGB565(testColor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device.colorToRGB565(testColor)
	}
}
