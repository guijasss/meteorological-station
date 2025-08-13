package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SensorEvent struct {
	Station   string  `json:"station"`
	Timestamp int64   `json:"timestamp"`
	Sensor    string  `json:"sensor"`
	Value     float64 `json:"value"`
}

type Sensor interface {
	Name() string
	Read() float64
	Station() string
	SetValue(float64)
}

type BaseSensor struct {
	name      string
	baseValue float64
	noise     float64
	station   string
}

func (s *BaseSensor) Name() string {
	return s.name
}

func (s *BaseSensor) Station() string {
	return s.station
}

func (s *BaseSensor) fluctuate() float64 {
	delta := (rand.Float64()*0.002 - 0.001)
	s.baseValue = math.Max(0.0, s.baseValue+delta)
	noise := (rand.Float64()*0.01 - 0.005)
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

type PressureSensor struct {
	BaseSensor
}

func NewPressureSensor() *PressureSensor {
	return &PressureSensor{
		BaseSensor{
			name:      "pressure",
			baseValue: 1013.0,
			noise:     0.5,
		},
	}
}

type WindSpeedSensor struct {
	BaseSensor
}

func NewWindSpeedSensor() *WindSpeedSensor {
	return &WindSpeedSensor{
		BaseSensor{
			name:      "wind_speed",
			baseValue: 90.0,
			noise:     0.2,
		},
	}
}

type SolarRadiationSensor struct {
	BaseSensor
}

func NewSolarRadiatioSensor() *SolarRadiationSensor {
	return &SolarRadiationSensor{
		BaseSensor{
			name:      "solar_radiation",
			baseValue: 200.0,
			noise:     1,
		},
	}
}

type UVSensor struct {
	BaseSensor
}

func NewUVSensor() *UVSensor {
	return &UVSensor{
		BaseSensor{
			name:      "uv",
			baseValue: 5.0,
			noise:     0.5,
		},
	}
}

type PrecipitationSensor struct {
	BaseSensor
}

func NewPrecipitationSensor() *PrecipitationSensor {
	return &PrecipitationSensor{
		BaseSensor{
			name:      "precipitation_rate",
			baseValue: 0.0,
			noise:     0.01,
		},
	}
}

type SoilHumiditySensor struct {
	BaseSensor
}

func NewSoilHumiditySensor() *SoilHumiditySensor {
	return &SoilHumiditySensor{
		BaseSensor{
			name:      "soil_humidity",
			baseValue: 30.0,
			noise:     1.0,
		},
	}
}

type WeatherStation struct {
	StationID string
	Sensors   map[string]Sensor
}

func NewWeatherStation() *WeatherStation {
	return &WeatherStation{
		StationID: uuid.New().String(),
		Sensors: map[string]Sensor{
			"air_humidity":       NewAirHumiditySensor(),
			"precipitation_rate": NewPrecipitationSensor(),
			"pressure_rate":      NewPressureSensor(),
			"soil_humidity":      NewSoilHumiditySensor(),
			"solar_radiation":    NewSolarRadiatioSensor(),
			"temperature":        NewTemperatureSensor(),
			"uv":                 NewUVSensor(),
			"wind_direction":     NewWindDirectionSensor(),
			"wind_speed":         NewWindSpeedSensor(),
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

func (ws *WeatherStation) ReadAll() []SensorEvent {
	events := make([]SensorEvent, 0)
	for _, sensor := range ws.Sensors {
		events = append(events, SensorEvent{
			Station:   sensor.Station(),
			Timestamp: time.Now().Unix(),
			Sensor:    sensor.Name(),
			Value:     sensor.Read(),
		})
	}
	return events
}
