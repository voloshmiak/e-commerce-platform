package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pb "product-catalog-service/protobuf"
	"strconv"
	"time"
)

func ListenForOrderCreated() error {
	time.Sleep(30 * time.Second)
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewProductCatalogServiceClient(conn)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "orders.created",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for  order creation messages..")

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

			resp, err := client.CheckStock(context.Background(), &pb.CheckStockRequest{
				Id:       productID,
				Quantity: int32(quantity),
			})
			if err != nil {
				log.Printf("Failed to check stock for product %d: %v", productID, err)
				continue
			}

			if !resp.GetInStock() {
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
					return fmt.Errorf("failed to write message to Kafka: %v", err)
				}

				break
			}
		}

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
			return fmt.Errorf("failed to write message to Kafka: %v", err)
		}
	}

	return nil
}
