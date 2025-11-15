package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderCreated struct {
	OrderID    uuid.UUID
	UserID     uuid.UUID
	TotalPrice int64
	Items      []OrderItem
	CreatedAt  time.Time
}

func (e OrderCreated) Type() string {
	return "order_created"
}

type OrderPaid struct {
	OrderID uuid.UUID
	PaidAt  time.Time
}

func (e OrderPaid) Type() string {
	return "order_paid"
}

type OrderCancelled struct {
	OrderID     uuid.UUID
	Reason      string
	CancelledAt time.Time
}

func (e OrderCancelled) Type() string {
	return "order_cancelled"
}
