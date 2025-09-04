package service

import (
	"log"
	"shopping-cart-service/repository"
	"strconv"
)

type ShoppingCartService struct {
	Repository repository.CartRepository
}

func NewShoppingCartService(repository repository.CartRepository) *ShoppingCartService {
	return &ShoppingCartService{
		Repository: repository,
	}
}

func (s *ShoppingCartService) GetCart(userID int64) (map[string]string, error) {
	items, err := s.Repository.GetCart(strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ShoppingCartService) AddItem(userID, productID int64, quantity int32, price float64) error {
	err := s.Repository.AddItem(strconv.FormatInt(userID, 10),
		strconv.FormatInt(productID, 10),
		strconv.Itoa(int(quantity)),
		strconv.FormatFloat(price, 'f', 2, 64),
	)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *ShoppingCartService) UpdateItem() {

}

func (s *ShoppingCartService) RemoveItem() {

}

func (s *ShoppingCartService) ClearCart() {

}
