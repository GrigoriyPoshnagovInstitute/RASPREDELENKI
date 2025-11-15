package model

import (
	"time"

	"github.com/google/uuid"
)

type AccountCreated struct {
	UserID    uuid.UUID
	Balance   int64
	CreatedAt time.Time
}

func (a AccountCreated) Type() string {
	return "account_created"
}

type AccountBalanceUpdated struct {
	UserID    uuid.UUID
	Balance   int64
	UpdatedAt time.Time
}

func (a AccountBalanceUpdated) Type() string {
	return "account_balance_updated"
}
