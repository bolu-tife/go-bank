package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Email    string `json:"email"`
	Token string `json:"token"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}
type AccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}
type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Email             string    `json:"email"`
	EncryptedPassword string    `json:"-"`
	Number            int64     `json:"number"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"createdAt"`
	Deleted           bool      `json:"deleted"`
}

func (a *Account) ValidatePassword(password string) bool{
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(password)) == nil
}

func NewAccount(firstName, lastName, email, password string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName:         firstName,
		LastName:          lastName,
		Email:             email,
		EncryptedPassword: string(hash),
		Number:            int64(rand.Intn(1000000)),
		CreatedAt:         time.Now().UTC(),
	}, nil
}
