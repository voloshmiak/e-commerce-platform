package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/refund"
	"google.golang.org/grpc/metadata"
	"html/template"
	"log"
	"math"
	"payment-service/internal/model"
	"payment-service/internal/repository"
	pb "payment-service/protobuf"
	"strconv"
	"time"
)

var (
	ErrEmptyCart              = errors.New("cart is empty")
	ErrProcessingPayment      = errors.New("error processing payment")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrSendingEvent           = errors.New("error sending event")
	ErrPaymentNotSucceeded    = errors.New("payment not succeeded")
	ErrMissingPaymentIntentID = errors.New("missing payment intent ID")
	ErrMissingMetadata        = errors.New("missing metadata in context")
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
	paymentFailedWriter  *kafka.Writer
	paymentSucceedWriter *kafka.Writer
}

func New(repo *repository.Repository, cartClient pb.ShoppingCartServiceClient, paymentFailedWriter, paymentSucceedWriter *kafka.Writer) *Service {
	return &Service{
		repo:                 repo,
		cartClient:           cartClient,
		paymentFailedWriter:  paymentFailedWriter,
		paymentSucceedWriter: paymentSucceedWriter,
	}
}

func (s *Service) ProcessPayment(ctx context.Context) (*stripe.PaymentIntent, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMissingMetadata
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := s.cartClient.GetCart(ctx, &pb.GetCartRequest{})
	if err != nil {
		return nil, err
	}

	items := resp.GetItems()

	if len(items) == 0 {
		return nil, ErrEmptyCart
	}

	var amount float64
	for _, item := range items {
		amount += item.GetPrice() * float64(item.GetQuantity())
	}

	amountInCents := int64(math.Round(amount*100) / 100)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountInCents),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, ErrProcessingPayment
	}

	transaction := model.Transaction{
		OrderID:              nil,
		Amount:               amount,
		Currency:             model.USD,
		Status:               model.Pending,
		GatewayTransactionID: &pi.ID,
		PaymentMethod:        model.Card,
	}

	_, err = s.repo.CreateTransaction(ctx, &transaction)
	if err != nil {
		return nil, err
	}

	return pi, nil
}

func (s *Service) ConfirmOrderPayment(ctx context.Context, eventData OrderData) error {
	if eventData.PaymentMethod == "ON_DELIVERY" {
		if err := s.sendPaymentSucceededEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}
		return nil
	}

	if eventData.PaymentIntentID == nil {
		if err := s.sendPaymentFailedEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}
		return ErrMissingPaymentIntentID
	}

	pi, err := paymentintent.Get(*eventData.PaymentIntentID, nil)
	if err != nil {
		if err = s.sendPaymentFailedEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}
		return err
	}

	orderAmountInCents := int64(math.Round(eventData.Amount*100) / 100)
	if pi.Amount != orderAmountInCents {
		if err = s.sendPaymentFailedEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}
		return ErrInvalidAmount
	}

	status := model.Failed
	if pi.Status == "succeeded" {
		status = model.Completed
	}

	if err = s.repo.UpdateTransactionStatus(ctx, *eventData.PaymentIntentID, status); err != nil {
		return err
	}

	if err = s.repo.UpdateTransactionOrderID(ctx, *eventData.PaymentIntentID, eventData.OrderID); err != nil {
		return err
	}

	if status == model.Failed {
		if err = s.sendPaymentFailedEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}
		return ErrPaymentNotSucceeded
	}

	if err = s.sendPaymentSucceededEvent(ctx, eventData); err != nil {
		return ErrSendingEvent
	}

	return nil
}

func (s *Service) CompensatePayment(ctx context.Context, eventData OrderData) error {
	if eventData.PaymentMethod == "ON_DELIVERY" {
		if err := s.sendPaymentFailedEvent(ctx, eventData); err != nil {
			return ErrSendingEvent
		}

		return nil
	}

	if eventData.PaymentIntentID == nil {
		return ErrMissingPaymentIntentID
	}

	pi, err := paymentintent.Get(*eventData.PaymentIntentID, nil)
	if err != nil {
		return ErrProcessingPayment
	}

	if pi.Status != "succeeded" {
		return nil
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(*eventData.PaymentIntentID),
	}

	r, err := refund.New(params)
	if err != nil {
		return ErrProcessingPayment
	}

	log.Printf("Successfully created a full refund with ID: %s\n", r.ID)
	log.Printf("Refund Status: %s\n", r.Status)

	err = s.repo.UpdateTransactionStatus(ctx, *eventData.PaymentIntentID, model.Refunded)
	if err != nil {
		return err
	}

	if err = s.sendPaymentFailedEvent(ctx, eventData); err != nil {
		return ErrSendingEvent
	}

	return nil
}

func (s *Service) sendPaymentFailedEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "payment.failed",
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

	err = s.paymentFailedWriter.WriteMessages(ctx, msg)

	return err
}

func (s *Service) sendPaymentSucceededEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "payment.succeeded",
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

	err = s.paymentSucceedWriter.WriteMessages(ctx, msg)

	return err
}
