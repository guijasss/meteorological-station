package main

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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
