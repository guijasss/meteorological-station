package main

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SensorEvent struct {
	station   string
	timestamp string
	sensor    string
	value     float64
}

func sendMQTTEvent(event SensorEvent, broker string) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go-mqtt-publisher")
	opts.SetConnectTimeout(5 * time.Second)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("erro ao conectar MQTT: %v", token.Error())
	}
	defer client.Disconnect(250)

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erro ao serializar evento: %v", err)
	}

	topic := "readings"
	token = client.Publish(topic, 0, false, payload)
	token.Wait() // espera envio terminar

	if token.Error() != nil {
		return fmt.Errorf("erro ao publicar evento: %v", token.Error())
	}

	return nil
}

func main() {
	event := SensorEvent{
		station:   "abc123",
		timestamp: "2025-08-08T14:00:00Z",
		sensor:    "temperature",
		value:     23.5,
	}

	err := sendMQTTEvent(event, "tcp://mqtt-broker:1883")
	if err != nil {
		fmt.Println("Erro:", err)
	} else {
		fmt.Println("Evento enviado com sucesso!")
	}
}
