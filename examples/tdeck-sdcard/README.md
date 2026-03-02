# T-Deck: список файлов с SD-карты

Пример выводит в UART список файлов и папок в корне FAT-раздела на SD-карте.

## Пины (utilities.h)

| Сигнал   | GPIO |
|----------|------|
| SD_CS    | 39   |
| SPI_SCK  | 40   |
| SPI_MOSI | 41   |
| SPI_MISO | 38   |

Питание периферии: GPIO10 (BOARD_POWERON) — в High перед работой с SD.

## Зависимости

- `tinygo.org/x/drivers/sdcard` — драйвер SD по SPI
- `tinygo.org/x/tinyfs/fatfs` — FAT для доступа к файлам (CGo)

## Сборка

```bash
tinygo build -target=esp32-s3-devkitc -o firmware.uf2 .
```

Подключи USB-UART для вывода списка файлов.
