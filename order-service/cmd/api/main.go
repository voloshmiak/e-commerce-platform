package main

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	pb "order-service/protobuf"
	"order-service/service"
	"strconv"
	"strings"
)

const (
	secret = "my-secret-key"
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

	s := grpc.NewServer(grpc.UnaryInterceptor(FetchUserIDInterceptor))
	pb.RegisterOrderServiceServer(s, &service.OrderService{})

	log.Println("Starting gRPC order service server on port :8080")

	if err = s.Serve(listener); err != nil {
		return err
	}

	return nil
}

func FetchUserIDInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	mt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata in context")
	}

	bearedToken := mt.Get("Authorization")[0]

	tokenString := strings.TrimPrefix(bearedToken, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)

	userID, ok := claims["user-id"]
	if !ok {
		return nil, fmt.Errorf("user-id not found in token claims")
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		return nil, fmt.Errorf("user-id is not a valid float64")
	}

	userIDInt := int(userIDFloat)

	md := metadata.Pairs("user-id", strconv.Itoa(userIDInt))

	ctx = metadata.NewIncomingContext(ctx, md)

	resp, err := handler(ctx, req)
	return resp, err
}
