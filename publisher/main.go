package main

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt-broker:1883") // ajuste se estiver em container
	opts.SetClientID("go-publisher")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("✅ Publicador conectado ao broker MQTT")

	topic := "readings"

	for i := 0; i < 10; i++ {
		payload := fmt.Sprintf("Temperatura %d: %d°C", i+1, 20+i)
		token := client.Publish(topic, 0, false, payload)
		token.Wait()
		fmt.Printf("📤 Enviado: %s\n", payload)
		time.Sleep(1 * time.Second)
	}

	client.Disconnect(250)
	fmt.Println("🔌 Publicador desconectado")
	os.Exit(0)
}
