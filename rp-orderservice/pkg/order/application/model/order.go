package model

import "github.com/google/uuid"

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
}

type CreateOrder struct {
	UserID uuid.UUID
	Items  []OrderItem
}

type Order struct {
	OrderID    uuid.UUID
	UserID     uuid.UUID
	Items      []OrderItem
	TotalPrice int64
	Status     int
	CreatedAt  int64
}
