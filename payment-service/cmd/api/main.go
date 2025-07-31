package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "payment-service/protobuf"
	"payment-service/service"
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

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &service.PaymentService{})

	log.Println("Starting gRPC payment service server on port :8080")

	if err = s.Serve(listener); err != nil {
		return err
	}

	return nil
}
