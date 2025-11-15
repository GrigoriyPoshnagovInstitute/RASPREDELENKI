package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound        = errors.New("product.go not found")
	ErrProductNameAlreadyUsed = errors.New("product.go name already used")
)

type Product struct {
	ProductID   uuid.UUID
	Name        string
	Description *string
	Price       int64 // Цена в копейках
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FindSpec struct {
	ProductID *uuid.UUID
	Name      *string
}

type ProductRepository interface {
	NextID() (uuid.UUID, error)
	Store(product Product) error
	Find(spec FindSpec) (*Product, error)
	Delete(productID uuid.UUID) error
}
