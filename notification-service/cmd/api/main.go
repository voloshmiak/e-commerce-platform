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
		Topic:    "orders.created",
		MaxBytes: 10e6,
	})
	defer r.Close()

	ns := &service.NotificationService{}

	log.Println("Listening got order creation messages...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
		}

		ns.SendNotification(string(m.Value))
	}
}
