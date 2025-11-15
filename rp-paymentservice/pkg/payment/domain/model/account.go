package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Account struct {
	UserID    uuid.UUID
	Balance   int64 // Баланс в копейках
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FindSpec struct {
	UserID *uuid.UUID
}

type AccountRepository interface {
	NextID(userID uuid.UUID) uuid.UUID
	Store(account Account) error
	Find(spec FindSpec) (*Account, error)
}
