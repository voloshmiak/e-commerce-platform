package main

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"order-service/consumer"
	pb "order-service/protobuf"
	"order-service/server"
	"order-service/service"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	svc := &service.OrderService{}
	cons := &consumer.OrderConsumer{}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, server.NewOrderServer(svc))

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Kafka consumer
	go func() {
		log.Println("Starting Kafka consumer for order created events...")

		err := cons.HandleOrderCreated()
		if err != nil {
			log.Println(err)
		}
	}()

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
