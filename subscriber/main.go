package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func sendAlert(client mqtt.Client, alert *AlertMessage) {
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		fmt.Printf("❌ Erro ao serializar alerta: %v\n", err)
		return
	}

	token := client.Publish("alerts", 0, false, alertJSON)
	if token.Wait() && token.Error() != nil {
		fmt.Printf("❌ Erro ao enviar alerta: %v\n", token.Error())
	} else {
		fmt.Printf("🚨 Alerta enviado: %s\n", alert.Values)
	}
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt-broker:1883")
	opts.SetClientID("go-subscriber-questdb")
	opts.CleanSession = false
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.SetMessageChannelDepth(100)

	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)

	processor := NewAsyncProcessor(1000, 10) // Buffer 1000, batch 10

	alerts := []Alert{
		HighTemperatureAlert{},
		LowHumidityAlert{},
	}

	var messageCount int64
	var mu sync.Mutex

	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		var event SensorEvent
		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			fmt.Printf("❌ Erro ao decodificar evento: %v\n", err)
			return
		}

		processor.SendEvent(event)

		for _, alert := range alerts {
			if alertMsg := alert.Check(event); alertMsg != nil {
				sendAlert(client, alertMsg)
			}
		}

		mu.Lock()
		messageCount++
		mu.Unlock()
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

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics, bufferSize := processor.GetMetrics()

				mu.Lock()
				msgCount := messageCount
				messageCount = 0 // Reset contador
				mu.Unlock()

				fmt.Printf("📊 Métricas (5s) - MQTT: %d msg/s | QuestDB: %d enviados, %d falhas, %d buffer\n",
					msgCount/5, metrics.Sent, metrics.Failed, bufferSize)
			}
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc

	fmt.Println("🛑 Iniciando shutdown graceful...")

	processor.Stop()

	client.Disconnect(250)
	fmt.Println("🔌 Assinante desconectado")

	metrics, _ := processor.GetMetrics()
	fmt.Printf("📋 Métricas Finais - QuestDB: %d enviados, %d falhas\n",
		metrics.Sent, metrics.Failed)
}
