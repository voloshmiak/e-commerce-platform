package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"log"
	"product-catalog-service/internal/model"
	"product-catalog-service/internal/repository"
	pb "product-catalog-service/protobuf"
	"strconv"
	"time"
)

var (
	ErrNotFound          = errors.New("err not found")
	ErrSendingEvent      = errors.New("error sending event")
	ErrInsufficientStock = errors.New("insufficient stock")
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
	mongoRepository     *repository.MongoRepository
	elasticRepository   *repository.ElasticRepository
	stockReservedWriter *kafka.Writer
	stockFailedWriter   *kafka.Writer
}

func New(mongoRepository *repository.MongoRepository, elasticRepository *repository.ElasticRepository, stockReservedWriter, stockFailedWriter *kafka.Writer) *Service {
	return &Service{
		mongoRepository:     mongoRepository,
		elasticRepository:   elasticRepository,
		stockReservedWriter: stockReservedWriter,
		stockFailedWriter:   stockFailedWriter,
	}
}

func (s *Service) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	product, err := s.mongoRepository.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return product, nil
}

func (s *Service) ListProducts(ctx context.Context, searchTerm string, page, pageSize int32) ([]*model.Product, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	from := (page - 1) * pageSize

	products, err := s.elasticRepository.GetProducts(ctx, searchTerm, from, pageSize)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Service) CreateProduct(ctx context.Context, r *pb.CreateProductRequest) (string, error) {
	isActive := r.GetStockQuantity() > 0

	product := &model.Product{
		ID:            primitive.NewObjectID(),
		Sku:           r.GetSku(),
		Name:          r.GetName(),
		Description:   r.GetDescription(),
		Price:         r.GetPrice(),
		Currency:      model.Currency(r.GetCurrency()),
		StockQuantity: r.GetStockQuantity(),
		Category:      r.GetCategory(),
		ImageURL:      r.GetImageUrl(),
		Attributes:    r.GetAttributes(),
		IsActive:      isActive,
		CreatedAt:     time.Now(),
	}

	id, err := s.mongoRepository.CreateProduct(ctx, product)
	if err != nil {
		return "", err
	}

	go func() {
		err = s.elasticRepository.CreateOrUpdateProduct(context.Background(), product)
		if err != nil {
			log.Printf("Error indexing product in Elasticsearch: %s", err)
		}
	}()

	return id, nil
}

func (s *Service) UpdateProduct(ctx context.Context, r *pb.UpdateProductRequest) error {
	id, err := primitive.ObjectIDFromHex(r.GetId())
	if err != nil {
		return err
	}

	isActive := r.GetStockQuantity() > 0

	product := &model.Product{
		ID:            id,
		Sku:           r.GetSku(),
		Name:          r.GetName(),
		Description:   r.GetDescription(),
		Price:         r.GetPrice(),
		Currency:      model.Currency(r.GetCurrency()),
		StockQuantity: r.GetStockQuantity(),
		Category:      r.GetCategory(),
		ImageURL:      r.GetImageUrl(),
		Attributes:    r.GetAttributes(),
		IsActive:      isActive,
	}

	result, err := s.mongoRepository.UpdateProduct(ctx, product)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	go func() {
		err = s.elasticRepository.CreateOrUpdateProduct(context.Background(), product)
		if err != nil {
			log.Printf("Error indexing product in Elasticsearch: %s", err)
		}
	}()

	return nil
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	err := s.mongoRepository.DeleteProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	go func() {
		err = s.elasticRepository.DeleteProduct(context.Background(), id)
		if err != nil {
			log.Printf("Error deleting product from Elasticsearch: %s", err)
		}
	}()

	return nil
}

func (s *Service) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	product, err := s.mongoRepository.GetProductBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return product, nil
}

func (s *Service) CheckAndReserveStock(ctx context.Context, eventData OrderData) error {
	skus := make([]string, len(eventData.Items))
	for i, item := range eventData.Items {
		skus[i] = item.Sku
	}
	products, err := s.mongoRepository.BulkGetBySKUs(ctx, skus)
	if err != nil {
		return err
	}

	productBySku := make(map[string]*model.Product, len(products))
	for _, product := range products {
		productBySku[product.Sku] = product
	}

	for _, item := range eventData.Items {
		product, exists := productBySku[item.Sku]
		if !exists {
			if err := s.sendStockFailedEvent(ctx, eventData); err != nil {
				return ErrSendingEvent
			}
			return ErrNotFound
		}

		fmt.Println(product.StockQuantity, "Product stock quantity")
		fmt.Println(item.Quantity, "Item quantity")

		if product.StockQuantity < item.Quantity {
			if err := s.sendStockFailedEvent(ctx, eventData); err != nil {
				return ErrSendingEvent
			}
			return ErrInsufficientStock
		}

		updatedProduct := &model.Product{
			ID:            product.ID,
			Sku:           product.Sku,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			Currency:      product.Currency,
			StockQuantity: product.StockQuantity - item.Quantity,
			Category:      product.Category,
			ImageURL:      product.ImageURL,
			Attributes:    product.Attributes,
			IsActive:      product.StockQuantity-item.Quantity > 0,
			CreatedAt:     product.CreatedAt,
		}

		result, err := s.mongoRepository.UpdateProduct(ctx, updatedProduct)
		if err != nil {
			return err
		}

		if result.MatchedCount == 0 {
			return ErrNotFound
		}

		go func(p *model.Product) {
			err := s.elasticRepository.CreateOrUpdateProduct(context.Background(), p)
			if err != nil {
				log.Printf("Error indexing product in Elasticsearch: %s", err)
			}
		}(updatedProduct)
	}

	if err = s.sendStockReservedEvent(ctx, eventData); err != nil {
		return ErrSendingEvent
	}

	return nil
}

func (s *Service) CompensateStock(ctx context.Context, eventData OrderData) error {
	skus := make([]string, len(eventData.Items))
	for i, item := range eventData.Items {
		skus[i] = item.Sku
	}
	products, err := s.mongoRepository.BulkGetBySKUs(ctx, skus)
	if err != nil {
		return err
	}

	productBySku := make(map[string]*model.Product, len(products))
	for _, product := range products {
		productBySku[product.Sku] = product
	}

	var successfullyCompensated = true
	var missingProducts []string

	for _, item := range eventData.Items {
		product, exists := productBySku[item.Sku]
		if !exists {
			missingProducts = append(missingProducts, item.Sku)
			successfullyCompensated = false
			continue
		}

		updatedProduct := &model.Product{
			ID:            product.ID,
			Sku:           product.Sku,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			Currency:      product.Currency,
			StockQuantity: product.StockQuantity + item.Quantity,
			Category:      product.Category,
			ImageURL:      product.ImageURL,
			Attributes:    product.Attributes,
			IsActive:      product.StockQuantity+item.Quantity > 0,
			CreatedAt:     product.CreatedAt,
		}

		result, err := s.mongoRepository.UpdateProduct(ctx, updatedProduct)
		if err != nil {
			log.Printf("Failed to compensate stock for product %s: %v", product.Sku, err)
			successfullyCompensated = false
			continue
		}

		if result.MatchedCount == 0 {
			log.Printf("Product %s not found during compensation", product.Sku)
			successfullyCompensated = false
			continue
		}

		go func(p *model.Product) {
			err := s.elasticRepository.CreateOrUpdateProduct(context.Background(), p)
			if err != nil {
				log.Printf("Error indexing product in Elasticsearch: %s", err)
			}
		}(updatedProduct)
	}

	if !successfullyCompensated {
		log.Printf("Warning: Could not compensate stock for products with SKUs: %v", missingProducts)
	}

	if err = s.sendStockFailedEvent(ctx, eventData); err != nil {
		return ErrSendingEvent
	}

	return nil
}

func (s *Service) sendStockReservedEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "stock.reserved",
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

	err = s.stockReservedWriter.WriteMessages(ctx, msg)

	return err
}

func (s *Service) sendStockFailedEvent(ctx context.Context, eventData OrderData) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "stock.reservation.failed",
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

	err = s.stockFailedWriter.WriteMessages(ctx, msg)

	return err
}
