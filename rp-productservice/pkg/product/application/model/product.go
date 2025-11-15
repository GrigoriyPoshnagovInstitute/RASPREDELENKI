package model

import "github.com/google/uuid"

type Product struct {
	ProductID   uuid.UUID
	Name        string
	Price       int64
	Description *string
}
