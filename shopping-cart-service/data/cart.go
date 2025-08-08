package data

type Cart struct {
	UserID int64       `json:"user_id"`
	Items  []*CartItem `json:"items"`
}

type CartItem struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

var Carts = []*Cart{
	{
		UserID: 1,
		Items: []*CartItem{
			{
				ID:        1,
				ProductID: 1,
				Quantity:  2,
				Price:     29.99,
			},
			{
				ID:        2,
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
				ID:        3,
				ProductID: 3,
				Quantity:  1,
				Price:     49.99,
			},
			{
				ID:        4,
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

func AddItem(userID int64, productID int64, quantity int32, price float64) int64 {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for _, item := range cart.Items {
				if item.ProductID == productID {
					item.Quantity += quantity
					return item.ID
				}
			}
			newItem := &CartItem{
				ID:        int64(len(cart.Items) + 1),
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			}
			cart.Items = append(cart.Items, newItem)
			return newItem.ID
		}
	}
	newCart := &Cart{
		UserID: userID,
		Items: []*CartItem{
			{
				ID:        1,
				ProductID: productID,
				Quantity:  quantity,
				Price:     price,
			},
		},
	}
	Carts = append(Carts, newCart)

	return newCart.Items[0].ID
}

func UpdateItemQuantity(userID int64, itemID int64, quantity int32) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for _, item := range cart.Items {
				if item.ID == itemID {
					item.Quantity = quantity
					return
				}
			}
			return
		}
	}
}

func RemoveItem(userID int64, itemID int64) {
	for _, cart := range Carts {
		if cart.UserID == userID {
			for i, item := range cart.Items {
				if item.ID == itemID {
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
