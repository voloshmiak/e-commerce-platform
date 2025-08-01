package data

type Cart struct {
	UserID int64       `json:"user_id"`
	Items  []*CartItem `json:"items"`
}

type CartItem struct {
	ProductID int64   `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

var Carts = []*Cart{
	{
		UserID: 1,
		Items: []*CartItem{
			{
				ProductID: 1,
				Quantity:  2,
				Price:     29.99,
			},
			{
				ProductID: 2,
				Quantity:  1,
				Price:     19.99,
			},
		},
	},
	{
		UserID: 2,
		Items: []*CartItem{
			{
				ProductID: 3,
				Quantity:  1,
				Price:     49.99,
			},
			{
				ProductID: 4,
				Quantity:  3,
				Price:     15.99,
			},
		},
	},
}

func GetCart(userID int64) *Cart {
	for _, cart := range Carts {
		if cart.UserID == userID {
			return cart
		}
	}
	return nil
}

func AddItem(userID int64, productID int64, quantity int32, price float64) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for _, item := range cart.Items {
				if item.ProductID == productID {
					item.Quantity += quantity
					return
				}
			}

			cart.Items = append(cart.Items, &CartItem{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			})
			return
		}
	}

	Carts = append(Carts, &Cart{
		UserID: userID,
		Items: []*CartItem{
			{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			},
		},
	})
}

func UpdateItem(userID int64, productID int64, quantity int32, price float64) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for _, item := range cart.Items {
				if item.ProductID == productID {
					item.Quantity = quantity
					item.Price = price
					return
				}
			}

			cart.Items = append(cart.Items, &CartItem{
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			})
			return
		}
	}
}

func RemoveItem(userID int64, productID int64) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for i, item := range cart.Items {
				if item.ProductID == productID {
					cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
					return
				}
			}
			return
		}
	}
}

func ClearCart(userID int64) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			cart.Items = []*CartItem{}
			return
		}
	}
}
