package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type NotificationProducer struct {
	writer *kafka.Writer
}

func NewNotificationProducer(brokers []string) *NotificationProducer {
	return &NotificationProducer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  "notifications",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
			RequiredAcks:           kafka.RequireOne, // Ждать подтверждения записи
			Async:                  false,            // Отправлять синхронно для надежности
			BatchSize:              1,                // Отправлять сразу, не копить пачку
		},
	}
}

func (p *NotificationProducer) PublishMessageEvent(ctx context.Context, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})

	if err != nil {
		log.Printf("KAFKA PRODUCER ERROR: %v", err) // Если не отправится — увидим в логах app
	} else {
		log.Printf("KAFKA PRODUCER: Message sent to Kafka")
	}

	return err
}
