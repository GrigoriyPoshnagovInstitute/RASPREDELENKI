package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductCreated struct {
	ProductID   uuid.UUID
	Name        string
	Description *string
	Price       int64
	CreatedAt   time.Time
}

func (p ProductCreated) Type() string {
	return "product_created"
}

type ProductUpdated struct {
	ProductID     uuid.UUID
	UpdatedFields struct {
		Name        *string
		Description *string
		Price       *int64
	}
	UpdatedAt time.Time
}

func (p ProductUpdated) Type() string {
	return "product_updated"
}

type ProductDeleted struct {
	ProductID uuid.UUID
	DeletedAt time.Time
}

func (p ProductDeleted) Type() string {
	return "product_deleted"
}
