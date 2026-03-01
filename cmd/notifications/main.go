package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	brokers := []string{os.Getenv("KAFKA_BROKERS")}
	topic := "notifications"

	log.Println("--- STARTING NOTIFICATION SERVICE ---")

	// Ждем 15 секунд, чтобы Kafka (Confluent тяжелее) успела полностью встать
	log.Println("Waiting for Kafka to warm up...")
	time.Sleep(15 * time.Second)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     "campus-notif-group-v9", // Новый ID группы
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     1 * time.Second,
		Logger:      log.New(os.Stdout, "[KAFKA-READER] ", 0),
	})

	defer reader.Close()

	for {
		log.Println("Reading next message...")
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		log.Printf("!!! RECEIVED: %s", string(m.Value))

		var event map[string]interface{}
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("JSON error: %v", err)
			continue
		}

		log.Printf(">>> NOTIFICATION SENT: %v", event["text"])
	}
}
