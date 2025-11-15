package service

import (
	"time"

	"github.com/google/uuid"

	"notificationservice/pkg/notification/domain/model"
)

type NotificationService interface {
	CreateNotification(orderID, userID uuid.UUID, message string) (uuid.UUID, error)
}

func NewNotificationService(repo model.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepository: repo,
	}
}

type notificationService struct {
	notificationRepository model.NotificationRepository
}

func (s *notificationService) CreateNotification(orderID, userID uuid.UUID, message string) (uuid.UUID, error) {
	notificationID, err := s.notificationRepository.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	notification := model.Notification{
		NotificationID: notificationID,
		OrderID:        orderID,
		UserID:         userID,
		Message:        message,
		CreatedAt:      time.Now(),
	}

	if err := s.notificationRepository.Store(notification); err != nil {
		return uuid.Nil, err
	}

	return notificationID, nil
}
