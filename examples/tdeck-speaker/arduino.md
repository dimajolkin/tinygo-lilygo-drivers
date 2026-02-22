#include <Arduino.h>
#include "driver/i2s.h"

static const int PIN_POWERON = 10; // BOARD_POWERON
static const int PIN_I2S_BCK = 7;  // I2S BCK
static const int PIN_I2S_DOUT = 6; // I2S DOUT (to amp)
static const int PIN_I2S_WS = 5;   // I2S WS/LRCLK

static const i2s_port_t I2S_PORT = I2S_NUM_0;

void setup() {
  pinMode(PIN_POWERON, OUTPUT);
  digitalWrite(PIN_POWERON, HIGH); // включаем питание периферии
  delay(50);

  i2s_config_t i2s_config = {
    .mode = (i2s_mode_t)(I2S_MODE_MASTER | I2S_MODE_TX),
    .sample_rate = 16000,
    .bits_per_sample = I2S_BITS_PER_SAMPLE_16BIT,
    .channel_format = I2S_CHANNEL_FMT_ONLY_LEFT, // моно
    .communication_format = I2S_COMM_FORMAT_I2S,
    .intr_alloc_flags = ESP_INTR_FLAG_LEVEL1,
    .dma_buf_count = 6,
    .dma_buf_len = 256,
    .use_apll = false,
    .tx_desc_auto_clear = true,
    .fixed_mclk = 0
  };

  i2s_pin_config_t pin_config = {
    .bck_io_num = PIN_I2S_BCK,
    .ws_io_num = PIN_I2S_WS,
    .data_out_num = PIN_I2S_DOUT,
    .data_in_num = I2S_PIN_NO_CHANGE
  };

  i2s_driver_install(I2S_PORT, &i2s_config, 0, NULL);
  i2s_set_pin(I2S_PORT, &pin_config);
  i2s_zero_dma_buffer(I2S_PORT);
}

void loop() {
  // Квадратная волна 1 кГц при 16 кГц семплинге: период 16 семплов, полпериода 8
  static int16_t buf[256];
  const int halfPeriod = 8;
  const int16_t amp = 1000; // громкость (0..~30000)

  for (int i = 0; i < 256; i++) {
    buf[i] = ((i / halfPeriod) % 2) ? amp : -amp;
  }

  size_t bytes_written = 0;
  i2s_write(I2S_PORT, buf, sizeof(buf), &bytes_written, portMAX_DELAY);
}