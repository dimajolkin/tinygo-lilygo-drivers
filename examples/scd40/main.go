package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/scd4x"
)

const (
	boardI2CSCL = machine.GPIO17
	boardI2CSDA = machine.GPIO18
)

func main() {
	i2c := machine.I2C1
	err := i2c.Configure(machine.I2CConfig{
		SCL: boardI2CSCL,
		SDA: boardI2CSDA,
	})
	if err != nil {
		println("I2C configure:", err)
		return
	}

	time.Sleep(1 * time.Second)
	sensor := scd4x.New(i2c)
	if err := sensor.Configure(); err != nil {
		println("sensor configure:", err)
		return
	}
	time.Sleep(500 * time.Millisecond)

	println("Starting periodic measurement...")
	if err := sensor.StartPeriodicMeasurement(); err != nil {
		println("start measurement:", err)
		return
	}
	time.Sleep(5 * time.Second)

	println("SCD40 CO2, temp, humidity (I2C)")
	for {
		ready, err := sensor.DataReady()
		if err != nil {
			println("DataReady:", err)
			time.Sleep(time.Second)
			continue
		}
		if !ready {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := sensor.ReadData(); err != nil {
			println("ReadData:", err)
			time.Sleep(time.Second)
			continue
		}
		co2 := sensor.CO2()
		tempC := float32(sensor.Temperature()) / 1000
		humPct := float32(sensor.Humidity()) / 100
		println("CO2:", co2, "ppm  temp:", tempC, "°C  humidity:", humPct, "%")
		time.Sleep(5 * time.Second)
	}
}
