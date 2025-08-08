package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt-broker:1883")
	opts.SetClientID("go-subscriber")

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("📥 Recebido no tópico %s: %s\n", msg.Topic(), string(msg.Payload()))
	}

	opts.SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("✅ Assinante conectado ao broker MQTT")

	topic := "sensors/temperature"
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Printf("📡 Assinando tópico: %s\n", topic)

	// Espera Ctrl+C para encerrar
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc

	client.Disconnect(250)
	fmt.Println("🔌 Assinante desconectado")
}
