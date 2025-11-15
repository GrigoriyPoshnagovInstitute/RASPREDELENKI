package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type Notification struct {
	NotificationID uuid.UUID
	OrderID        uuid.UUID
	UserID         uuid.UUID
	Message        string
	CreatedAt      time.Time
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification Notification) error
	FindForUser(userID uuid.UUID) ([]Notification, error)
}
