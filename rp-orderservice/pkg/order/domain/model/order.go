package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrProductNotFound = errors.New("product for order not found")
	ErrUserNotFound    = errors.New("user for order not found")
	ErrEmptyOrder      = errors.New("order must contain at least one item")
)

type OrderStatus int

const (
	StatusCreated OrderStatus = iota
	StatusPaymentPending
	StatusPaid
	StatusCancelled
)

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     int64 // Цена за единицу в копейках на момент заказа
}

type Order struct {
	OrderID    uuid.UUID
	UserID     uuid.UUID
	Items      []OrderItem
	TotalPrice int64 // Общая цена в копейках
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderRepository interface {
	NextID() (uuid.UUID, error)
	Store(order Order) error
	Find(orderID uuid.UUID) (*Order, error)
}
