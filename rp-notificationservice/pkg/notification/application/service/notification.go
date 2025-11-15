package service

import (
	"context"

	"github.com/google/uuid"

	"notificationservice/pkg/notification/domain/service"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, orderID, userID uuid.UUID, message string) (uuid.UUID, error)
}

func NewNotificationService(uow UnitOfWork) NotificationService {
	return &notificationService{uow: uow}
}

type notificationService struct {
	uow UnitOfWork
}

func (s *notificationService) CreateNotification(ctx context.Context, orderID, userID uuid.UUID, message string) (uuid.UUID, error) {
	var notificationID uuid.UUID
	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainService := service.NewNotificationService(provider.NotificationRepository(ctx))
		id, err := domainService.CreateNotification(orderID, userID, message)
		if err != nil {
			return err
		}
		notificationID = id
		return nil
	})
	return notificationID, err
}
