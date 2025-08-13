package main

type SensorEvent struct {
	Station   string  `json:"station"`
	Timestamp int64   `json:"timestamp"`
	Sensor    string  `json:"sensor"`
	Value     float64 `json:"value"`
}

type Alert interface {
	Check(event SensorEvent) *AlertMessage
}

type AlertMessage struct {
	Station   string             `json:"station"`
	Timestamp int64              `json:"timestamp"`
	Metric    string             `json:"metric"`
	Values    map[string]float64 `json:"values"`
	Severity  int                `json:"severity"`
}
