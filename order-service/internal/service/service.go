package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"html/template"
	"log"
	"order-service/internal/model"
	"order-service/internal/repository"
	pb "order-service/protobuf"
	"strconv"
	"time"
)

var (
	ErrEmptyCart       = errors.New("cart is empty")
	ErrGetCart         = errors.New("failed to get cart")
	ErrMissingUserID   = errors.New("missing user id in context")
	ErrMissingMetadata = errors.New("missing metadata in context")
	ErrSendingEvent    = errors.New("error sending event")
	ErrUserNotFound    = errors.New("user not found")
)

type OrderCreatedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Data      OrderData `json:"data"`
}

type OrderData struct {
	CustomerFirstName string           `json:"customer_first_name"`
	CustomerLastName  string           `json:"customer_last_name"`
	CustomerEmail     string           `json:"customer_email"`
	OrderID           int64            `json:"order_id"`
	OrderDate         time.Time        `json:"order_date"`
	PaymentMethod     string           `json:"payment_method"`
	PaymentIntentID   *string          `json:"payment_intent_id"`
	UserID            int64            `json:"user_id"`
	Items             []*OrderItemData `json:"items"`
	Amount            float64          `json:"amount"`
	ShippingAddress   string           `json:"shipping_address"`
	EstimatedDelivery time.Time        `json:"estimated_delivery"`
}

type OrderItemData struct {
	Quantity       int32        `json:"quantity"`
	Price          float64      `json:"price"`
	Sku            string       `json:"sku"`
	Name           string       `json:"name"`
	ImageURL       template.URL `json:"image_url"`
	ItemTotalPrice float64      `json:"item_total_price"`
}

type Service struct {
	repo                 *repository.Repository
	cartClient           pb.ShoppingCartServiceClient
	userClient           pb.UserServiceClient
	orderCreatedWriter   *kafka.Writer
	orderConfirmedWriter *kafka.Writer
}

func New(repo *repository.Repository, cartClient pb.ShoppingCartServiceClient, userClient pb.UserServiceClient, orderCreatedWriter, orderConfirmedWriter *kafka.Writer) *Service {
	return &Service{
		repo:                 repo,
		cartClient:           cartClient,
		userClient:           userClient,
		orderCreatedWriter:   orderCreatedWriter,
		orderConfirmedWriter: orderConfirmedWriter,
	}
}

func (s *Service) CreateOrderSaga(ctx context.Context, userID int64, shippingAddress, paymentMethod string, paymentIntentID *string) (int64, string, error) {
	mt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", ErrMissingMetadata
	}

	token := mt.Get("Authorization")[0]
	md := metadata.Pairs("Authorization", token)
	outgoingCtx := metadata.NewOutgoingContext(ctx, md)

	getCartResp, err := s.cartClient.GetCart(outgoingCtx, &pb.GetCartRequest{})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return 0, "", ErrMissingUserID
		}
		return 0, "", ErrGetCart
	}

	itemsLen := len(getCartResp.GetItems())

	if itemsLen == 0 {
		return 0, "", ErrEmptyCart
	}

	items := make([]*model.OrderItem, itemsLen)
	for i, item := range getCartResp.GetItems() {
		items[i] = &model.OrderItem{
			Quantity: int64(item.GetQuantity()),
			Price:    item.GetPrice(),
			Sku:      item.GetSku(),
		}
	}

	log.Println("Creating order..")

	amount := getCartResp.GetTotalPrice()

	order := &model.Order{
		UserID:          userID,
		Status:          model.Pending,
		Items:           items,
		TotalPrice:      amount,
		ShippingAddress: shippingAddress,
		CreatedAt:       time.Now(),
	}

	orderID, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return 0, "", err
	}
	order.ID = orderID

	_, err = s.cartClient.ClearCart(outgoingCtx, &pb.ClearCartRequest{})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return 0, "", ErrMissingUserID
		}
		return 0, "", err
	}

	var dataItems []*OrderItemData
	for _, item := range getCartResp.GetItems() {
		dataItems = append(dataItems, &OrderItemData{
			Quantity:       item.Quantity,
			Price:          item.Price,
			Sku:            item.Sku,
			Name:           item.Name,
			ImageURL:       template.URL(item.ImageUrl),
			ItemTotalPrice: float64(item.Quantity) * item.Price,
		})
	}

	userResp, err := s.userClient.GetProfile(outgoingCtx, &pb.GetUserRequest{})
	if err != nil {
		switch {
		case status.Code(err) == codes.FailedPrecondition:
			return 0, "", ErrMissingUserID
		case status.Code(err) == codes.NotFound:
			return 0, "", ErrUserNotFound
		}
		return 0, "", err
	}

	orderData := OrderData{
		CustomerFirstName: userResp.GetFirstName(),
		CustomerLastName:  userResp.GetLastName(),
		CustomerEmail:     userResp.GetEmail(),
		OrderID:           order.ID,
		OrderDate:         order.CreatedAt,
		PaymentMethod:     paymentMethod,
		PaymentIntentID:   paymentIntentID,
		UserID:            userID,
		Items:             dataItems,
		Amount:            amount,
		ShippingAddress:   shippingAddress,
		EstimatedDelivery: time.Now().Add(72 * time.Hour),
	}

	if err := s.sendOrderCreatedEvent(ctx, orderData); err != nil {
		return 0, "", ErrSendingEvent
	}

	return order.ID, string(order.Status), nil
}

func (s *Service) GetOrder(ctx context.Context, userID int64, orderID int64) (*model.Order, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found for user")
	}

	return order, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int64) ([]*model.Order, error) {
	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Service) ConfirmOrder(ctx context.Context, eventData OrderData) error {
	var st = model.Paid
	if eventData.PaymentMethod == "ON_DELIVERY" {
		st = model.Confirmed
	}

	err := s.repo.UpdateOrderStatus(ctx, eventData.OrderID, st)
	if err != nil {
		return err
	}

	if err = s.sendOrderConfirmedEvent(ctx, eventData); err != nil {
		return ErrSendingEvent
	}

	return nil
}

func (s *Service) CancelOrder(ctx context.Context, eventData OrderData) error {
	err := s.repo.UpdateOrderStatus(ctx, eventData.OrderID, model.Cancelled)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendOrderConfirmedEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "orders.confirmed",
		Timestamp: time.Now(),
		Version:   "1.0",
		Data:      eventData,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(strconv.Itoa(int(eventData.OrderID))),
		Value: eventBytes,
	}

	err = s.orderConfirmedWriter.WriteMessages(ctx, msg)

	return err
}

func (s *Service) sendOrderCreatedEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "orders.created",
		Timestamp: time.Now(),
		Version:   "1.0",
		Data:      eventData,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(strconv.Itoa(int(eventData.OrderID))),
		Value: eventBytes,
	}

	err = s.orderCreatedWriter.WriteMessages(ctx, msg)

	return err
}
