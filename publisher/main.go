package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func publishEvents(events []SensorEvent, client mqtt.Client) {
	for _, event := range events {
		data, _ := json.Marshal(event)
		token := client.Publish("readings", 0, false, data)
		token.Wait()
		fmt.Printf("📡 enviado: %s\n", data)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	opts := mqtt.NewClientOptions().
		AddBroker("tcp://mqtt-broker:1883").
		SetClientID("weather-station")

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)

	ws := NewWeatherStation()

	for {
		events := ws.ReadAll()
		publishEvents(events, client)
		time.Sleep(1 * time.Second)
	}
}
