package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func sanitizeTag(tag string) string {
	original := tag
	tag = strings.ReplaceAll(tag, " ", "_")
	tag = strings.ReplaceAll(tag, ",", "_")
	tag = strings.ReplaceAll(tag, "=", "_")
	tag = strings.ReplaceAll(tag, "\"", "_")
	tag = strings.ReplaceAll(tag, "'", "_")

	if original != tag {
		fmt.Printf("🔍 Tag sanitizada: '%s' -> '%s'\n", original, tag)
	}

	return tag
}

func sendAlert(client mqtt.Client, alert *AlertMessage) {
	payload, _ := json.Marshal(alert)
	token := client.Publish("alerts", 1, true, payload)
	token.Wait()
	fmt.Println("🚨 Alerta enviado:", string(payload))
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt-broker:1883")
	opts.SetClientID("go-subscriber")
	opts.CleanSession = false
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.SetMessageChannelDepth(100)

	alerts := []Alert{
		HighTemperatureAlert{},
		LowHumidityAlert{},
	}

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		var event SensorEvent
		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			fmt.Println("Erro ao decodificar evento:", err)
			return
		}

		sendToQuestDB(event)

		for _, alert := range alerts {
			if alertMsg := alert.Check(event); alertMsg != nil {
				sendAlert(client, alertMsg)
			}
		}
	}

	opts.SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("✅ Assinante conectado ao broker MQTT")

	topic := "readings"
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Printf("📡 Assinando tópico: %s\n", topic)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc

	client.Disconnect(250)
	fmt.Println("🔌 Assinante desconectado")
}
