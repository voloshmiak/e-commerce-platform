package consumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"order-service/data"
	"order-service/service"
	"time"
)

type OrderConsumer struct {
	Service *service.OrderService
}

func (oc *OrderConsumer) ListenForSucceededPayment() {
	time.Sleep(30 * time.Second)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "payment.succeeded",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for payment.succeeded messages...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var order map[string]interface{}
		err = json.Unmarshal(m.Value, &order)
		if err != nil {
			log.Println(err)
			continue
		}

		oc.Service.UpdateOrderStatus(int64(order["id"].(float64)), string(data.Paid))
	}
}

func (oc *OrderConsumer) ListenForStockFailed() {
	time.Sleep(30 * time.Second)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "stock.reservation.failed",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for stock.reservation.failed messages...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var order map[string]interface{}
		err = json.Unmarshal(m.Value, &order)
		if err != nil {
			log.Println(err)
			continue
		}

		oc.Service.UpdateOrderStatus(int64(order["id"].(float64)), string(data.Cancelled))

		log.Println("Stock reservation failed for order ID:", order["id"])
	}
}
