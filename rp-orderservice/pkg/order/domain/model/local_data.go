package model

import (
	"github.com/google/uuid"
)

type LocalUser struct {
	UserID uuid.UUID
	Login  string
}

type LocalUserRepository interface {
	Store(user LocalUser) error
	Find(userID uuid.UUID) (*LocalUser, error)
}

type LocalProduct struct {
	ProductID uuid.UUID
	Name      string
	Price     int64
}

type LocalProductRepository interface {
	Store(product LocalProduct) error
	Find(productID uuid.UUID) (*LocalProduct, error)
	FindMany(productIDs []uuid.UUID) ([]LocalProduct, error)
}
