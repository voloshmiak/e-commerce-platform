package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"notification-service/service"
	"time"
)

func main() {
	time.Sleep(30 * time.Second)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "payment.succeeded",
		MaxBytes: 10e6,
	})
	defer r.Close()

	ns := &service.NotificationService{}

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
		}

		ns.SendNotification(string(m.Value))
	}
}
