package model

import "github.com/google/uuid"

type Notification struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID
	OrderID        uuid.UUID
	Message        string
	CreatedAt      int64
}
