package main

import (
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"order-service/consumer"
	pb "order-service/protobuf"
	"order-service/server"
	"order-service/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	time.Sleep(50 * time.Second)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	conn, err := grpc.NewClient("shopping-cart-service:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	cartClient := pb.NewShoppingCartServiceClient(conn)
	svc := &service.OrderService{}
	cons := &consumer.OrderConsumer{Service: svc}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, server.NewOrderServer(svc, cartClient))

	// Kafka consumer
	go cons.ListenForSucceededPayment()
	go cons.ListenForStockFailed()

	// gRPC server
	go func() {
		log.Println("Starting gRPC order svc server on port :8080")

		if err = s.Serve(listener); err != nil || errors.Is(err, grpc.ErrServerStopped) {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Received shutdown signal, stopping server...")

	s.GracefulStop()

	return nil
}
