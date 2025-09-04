package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"payment-service/data"
	"payment-service/service"
	"strconv"
	"time"
)

type PaymentConsumer struct {
	Service *service.PaymentService
}

func (pc *PaymentConsumer) ListenForStockReserved() {
	time.Sleep(30 * time.Second)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "stock.reserved",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for stock.reserved messages..")

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
			break
		}

		log.Println("Processing payment for order:", order)

		_, err = pc.Service.ProcessPayment(int64(order["id"].(float64)), order["total_price"].(float64), string(data.USD), string(data.CreditCard))
		if err != nil {
			w := &kafka.Writer{
				Addr:                   kafka.TCP("kafka"),
				Topic:                  "payment.failed",
				Balancer:               &kafka.LeastBytes{},
				AllowAutoTopicCreation: true,
			}
			defer w.Close()

			err = w.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
				Value: []byte(fmt.Sprintf("Payment failed for order %d: %v", int(order["id"].(float64)), err)),
			})

			if err != nil {
				log.Printf("failed to write message to Kafka: %v", err)
			}
		}

		w := &kafka.Writer{
			Addr:                   kafka.TCP("kafka"),
			Topic:                  "payment.succeeded",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		}
		defer w.Close()

		encodedOrder, err := json.Marshal(order)
		if err != nil {
			log.Printf("failed to marshal order: %v", err)
		}

		err = w.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
			Value: encodedOrder,
		})

		if err != nil {
			log.Printf("failed to write message to Kafka: %v", err)
		}
	}
}
