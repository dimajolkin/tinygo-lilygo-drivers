# SCD40 CO2 sensor (I2C)

Пример чтения CO2, температуры и влажности с Sensirion SCD40 по I2C. Вывод в serial (println).

Пины по умолчанию для ESP32-S3: SCL=GPIO17, SDA=GPIO18. При необходимости измени в `main.go`.

Сборка:
```bash
tinygo build -target=esp32-s3 -o firmware.uf2 .
```
