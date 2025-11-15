package model

import "github.com/google/uuid"

type UserBalance struct {
	UserID  uuid.UUID
	Balance int64
}
