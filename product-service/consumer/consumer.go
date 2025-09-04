package consumer

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"product-catalog-service/service"
	"strconv"
	"time"
)

type ProductConsumer struct {
	Service *service.ProductCatalogService
}

func (c *ProductConsumer) ListenForOrderCreated() {
	time.Sleep(30 * time.Second)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "orders.created",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for orders.created messages..")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println(string(m.Value))

		var order map[string]interface{}
		err = json.Unmarshal(m.Value, &order)
		if err != nil {
			log.Println(err)
			break
		}

		for _, item := range order["items"].([]interface{}) {
			itemMap := item.(map[string]interface{})
			productID := int64(itemMap["product_id"].(float64))
			quantity := int64(itemMap["quantity"].(float64))

			inStock := c.Service.CheckStock(productID, int32(quantity))

			if !inStock {
				w := &kafka.Writer{
					Addr:                   kafka.TCP("kafka"),
					Topic:                  "stock.reservation.failed",
					Balancer:               &kafka.LeastBytes{},
					AllowAutoTopicCreation: true,
				}
				defer w.Close()

				err = w.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
					Value: m.Value,
				})

				if err != nil {
					log.Printf("failed to write message to Kafka: %v", err)
				}

				break
			}
		}

		log.Println("All items in stock, reserving stock...")

		w := &kafka.Writer{
			Addr:                   kafka.TCP("kafka"),
			Topic:                  "stock.reserved",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		}
		defer w.Close()

		err = w.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
			Value: m.Value,
		})

		if err != nil {
			log.Printf("failed to write message to Kafka: %v", err)
		}
	}
}

func (c *ProductConsumer) ListenForFailedPayment() {
	time.Sleep(30 * time.Second)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "payment.failed",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for payment.failed messages...")

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

		w := &kafka.Writer{
			Addr:                   kafka.TCP("kafka"),
			Topic:                  "stock.reservation.failed",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		}
		defer w.Close()

		err = w.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
			Value: m.Value,
		})

		if err != nil {
			log.Printf("failed to write message to Kafka: %v\n", err)
		}
	}
}
