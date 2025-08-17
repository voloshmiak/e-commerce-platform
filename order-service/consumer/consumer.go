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

type Payload struct {
	counter int
	id      int64
}

type OrderConsumer struct {
	service *service.OrderService
}

func (oc *OrderConsumer) listenForSucceededPayment(payload *Payload) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "payment.succeeded",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for succeeded payment messages...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("Part completed: payment succeeded message received:", string(m.Value))

		payload.counter++
	}
}

func (oc *OrderConsumer) listenForFailedPayment(payload *Payload) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "payment.failed",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for failed payments...")

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

		oc.service.UpdateOrderStatus(int64(order["id"].(float64)), string(data.Cancelled))

		log.Println("Payment failed for order ID:", order["id"])

		payload.counter = 0
	}
}

func (oc *OrderConsumer) listenForStockReserved(payload *Payload) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "stock.reserved",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for stock reserved...")

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

		payload.id = int64(order["id"].(float64))

		log.Println("Part completed: stock reserved message received:", string(m.Value))

		payload.counter++
	}
}

func (oc *OrderConsumer) listenForStockFailed(payload *Payload) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "stock.reservation.failed",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for stock reservation failed...")

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

		oc.service.UpdateOrderStatus(int64(order["id"].(float64)), string(data.Cancelled))

		log.Println("Stock reservation failed for order ID:", order["id"])

		payload.counter = 0
	}
}

func (oc *OrderConsumer) HandleOrderCreated() error {
	time.Sleep(30 * time.Second)
	var payload Payload
	go func() {
		err := oc.listenForStockReserved(&payload)
		if err != nil {
			log.Println("Error listening for stock reserved:", err)
		}
	}()
	go func() {
		err := oc.listenForSucceededPayment(&payload)
		if err != nil {
			log.Println("Error listening for stock reservation failed:", err)
		}
	}()
	go func() {
		err := oc.listenForFailedPayment(&payload)
		if err != nil {
			log.Println(err)
		}
	}()
	go func() {
		err := oc.listenForStockFailed(&payload)
		if err != nil {
			log.Println(err)
		}
	}()

	log.Println("Waiting for saga to complete...")

	for {
		if payload.counter == 2 {
			oc.service.UpdateOrderStatus(payload.id, string(data.Paid))

			payload.counter = 0
		}
	}
}
