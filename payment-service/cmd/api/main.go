package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"payment-service/consumer"
	pb "payment-service/protobuf"
	"payment-service/server"
	"payment-service/service"
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

	svc := &service.PaymentService{}
	cons := &consumer.PaymentConsumer{Service: svc}

	// Kafka consumer
	go cons.ListenForStockReserved()

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server.PaymentServer{})

	// gRPC server
	go func() {
		log.Println("Starting gRPC payment service server on port :8080")

		if err = s.Serve(listener); err != nil {
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
