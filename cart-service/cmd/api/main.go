package main

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"os"
	"os/signal"
	pb "shopping-cart-service/protobuf"
	"shopping-cart-service/repository"
	"shopping-cart-service/server"
	"shopping-cart-service/service"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	time.Sleep(30 * time.Second)
	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	repo := repository.NewCartStorage(rdb, 3600)
	svc := &service.ShoppingCartService{
		Repository: repo,
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(FetchUserIDInterceptor))
	pb.RegisterShoppingCartServiceServer(s, &server.ShoppingCartServer{
		Service: svc,
	})

	// gRPC server
	go func() {
		log.Println("Starting gRPC shopping cart service server on port :8080")

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

func FetchUserIDInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	mt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata in context")
	}

	fmt.Println(1)

	fmt.Println(ctx)

	id := mt.Get("user-id")
	if len(id) > 0 {
		return handler(ctx, req)
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
