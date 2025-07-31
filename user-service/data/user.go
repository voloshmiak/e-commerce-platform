package data

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreatedAt    time.Time `json:"created_at"`
}

var Users = []*User{
	{
		ID:           1,
		Email:        "example@email.com",
		PasswordHash: "$2a$10$EIX/3z5Z1",
		FirstName:    "John",
		LastName:     "Doe",
		CreatedAt:    time.Now(),
	},
	{
		ID:           2,
		Email:        "admin@email.com",
		PasswordHash: "$2a$10$EIX/3z5Z1",
		FirstName:    "Admin",
		LastName:     "User",
		CreatedAt:    time.Now(),
	},
}

func AddUser(email string, password string, firstName string, lastName string) int64 {
	user := &User{
		ID:           int64(len(Users) + 1),
		Email:        email,
		PasswordHash: password,
		FirstName:    firstName,
		LastName:     lastName,
		CreatedAt:    time.Now(),
	}
	Users = append(Users, user)
	return user.ID
}

func GetUserByEmail(email string) *User {
	for _, user := range Users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

func GetUserByID(id int64) *User {
	for _, user := range Users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

func UpdateUser(id int64, email string, firstName string, lastName string) {
	for i, user := range Users {
		if user.ID == id {
			Users[i].Email = email
			Users[i].FirstName = firstName
			Users[i].LastName = lastName
		}
	}
}
