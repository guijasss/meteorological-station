package main

import (
	"fmt"
	"net"
)

func sendToQuestDB(event SensorEvent) {
	station := sanitizeTag(event.Station)
	sensor := sanitizeTag(event.Sensor)
	timestampNanos := event.Timestamp * 1_000_000_000

	line := fmt.Sprintf(
		"readings,station=%s,sensor=%s value=%f %d\n",
		station,
		sensor,
		event.Value,
		timestampNanos,
	)

	conn, err := net.Dial("tcp", "questdb:9009")
	if err != nil {
		fmt.Printf("❌ Erro TCP: %v\n", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(line))
	if err != nil {
		fmt.Printf("❌ Erro ao enviar via TCP: %v\n", err)
		return
	}

	fmt.Println("✅ Dados enviados via TCP!")
}
