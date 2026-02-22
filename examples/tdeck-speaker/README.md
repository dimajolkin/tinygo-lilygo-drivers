# T-Deck: вывод звука по I2S

В примере **программный I2S** (bit-bang) на пинах 5, 7, 6 по образцу оф. библиотеки; при старте — серия тонов разной частоты (PWM, меандр, I2S).

## Пины (utilities.h)

| Сигнал   | GPIO |
|----------|------|
| I2S_WS   | 5    |
| I2S_BCK  | 7    |
| I2S_DOUT | 6    |

ES7210 (47, 14, 21, 48) — микрофон, для спикера не используется.

## Как устроено

- Пины и формат как в оф. либе [ESP32-audioI2S](https://github.com/Xinyuan-LilyGO/T-Deck/tree/master/lib/ESP32-audioI2S): `setPinout(BCLK=7, LRC=5, DOUT=6)`, I2S STAND (MSB), 16 bit, 16 kHz.
- Слот 16 бит на канал (без доп. нулей), как в `playSample()` — один uint32_t (L+R).
- Реализация — программный I2S (bit-bang); для стабильного звука в прошивке используют аппаратный I2S (ESP-IDF).

## Сборка

```bash
tinygo build -target=esp32-s3-devkitc -o firmware.uf2 .
```
