package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/google/uuid"
)

type SensorEvent struct {
	station   string
	timestamp string
	sensor    string
	value     float64
}

type Sensor interface {
	Name() string
	Read() float64
	SetValue(float64)
}

type BaseSensor struct {
	name      string
	baseValue float64
	noise     float64
}

func (s *BaseSensor) Name() string {
	return s.name
}

func (s *BaseSensor) fluctuate() float64 {
	delta := (rand.Float64()*0.002 - 0.001) // entre -0.001 e 0.001
	s.baseValue = math.Max(0.0, s.baseValue+delta)
	noise := (rand.Float64()*0.01 - 0.005) // entre -0.005 e 0.005
	return math.Max(0.0, s.baseValue+noise)
}

func (s *BaseSensor) SetValue(newValue float64) {
	if newValue < 0.0 {
		s.baseValue = 0.0
	} else {
		s.baseValue = newValue
	}
}

func (s *BaseSensor) Read() float64 {
	return math.Round(s.fluctuate()*100) / 100
}

type TemperatureSensor struct {
	BaseSensor
}

func NewTemperatureSensor() *TemperatureSensor {
	return &TemperatureSensor{
		BaseSensor{
			name:      "temperature",
			baseValue: 20.0,
			noise:     0.3,
		},
	}
}

type AirHumiditySensor struct {
	BaseSensor
}

func NewAirHumiditySensor() *AirHumiditySensor {
	return &AirHumiditySensor{
		BaseSensor{
			name:      "air_humidity",
			baseValue: 60.0,
			noise:     1.0,
		},
	}
}

type WindDirectionSensor struct {
	BaseSensor
}

func NewWindDirectionSensor() *WindDirectionSensor {
	return &WindDirectionSensor{
		BaseSensor{
			name:      "wind_direction",
			baseValue: 90.0,
			noise:     5.0,
		},
	}
}

func (s *WindDirectionSensor) Read() float64 {
	// flutuação circular
	noise := (rand.Float64()*2*s.noise - s.noise)
	val := math.Mod(s.baseValue+noise+360.0, 360.0)
	return math.Round(val*100) / 100
}

type WeatherStation struct {
	StationID string
	Sensors   map[string]Sensor
}

func NewWeatherStation() *WeatherStation {
	return &WeatherStation{
		StationID: uuid.New().String(),
		Sensors: map[string]Sensor{
			"temperature":    NewTemperatureSensor(),
			"air_humidity":   NewAirHumiditySensor(),
			"wind_direction": NewWindDirectionSensor(),
			// Adicione outros sensores aqui...
		},
	}
}

func (ws *WeatherStation) ReadSensor(sensorName string) (float64, error) {
	sensor, ok := ws.Sensors[sensorName]
	if !ok {
		return 0, fmt.Errorf("sensor %q not found", sensorName)
	}
	return sensor.Read(), nil
}

func (ws *WeatherStation) SetSensorValue(sensorName string, newValue float64) error {
	sensor, ok := ws.Sensors[sensorName]
	if !ok {
		return fmt.Errorf("sensor %q not found", sensorName)
	}
	sensor.SetValue(newValue)
	return nil
}
