package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"notification-service/service"
)

func main() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:     "topic-A",
		Partition: 0,
		MaxBytes:  10e6,
	})

	ns := &service.NotificationService{}

	log.Println("Starting notification service server")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		ns.SendNotification(string(m.Value))
	}
}
