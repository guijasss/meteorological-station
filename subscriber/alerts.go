package main

type HighTemperatureAlert struct{}

func (a HighTemperatureAlert) Check(event SensorEvent) *AlertMessage {
	if event.Sensor == "temperature" && event.Value > 1 {
		return &AlertMessage{
			Station:   event.Station,
			Timestamp: event.Timestamp,
			Metric:    "temperature",
			Values:    map[string]float64{"temperature": event.Value},
			Severity:  2,
		}
	}
	return nil
}

type LowHumidityAlert struct{}

func (a LowHumidityAlert) Check(event SensorEvent) *AlertMessage {
	if event.Sensor == "humidity" && event.Value < 35 {
		return &AlertMessage{
			Station:   event.Station,
			Timestamp: event.Timestamp,
			Metric:    "humidity",
			Values:    map[string]float64{"humidity": event.Value},
			Severity:  1,
		}
	}
	return nil
}
