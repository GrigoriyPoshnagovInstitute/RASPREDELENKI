package repository

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notificationservice/pkg/notification/domain/model"
)

func NewNotificationRepository(ctx context.Context, client mysql.ClientContext) model.NotificationRepository {
	return &notificationRepository{
		ctx:    ctx,
		client: client,
	}
}

type notificationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *notificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *notificationRepository) Store(notification model.Notification) error {
	_, err := r.client.ExecContext(r.ctx,
		`INSERT INTO notification (notification_id, order_id, user_id, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		notification.NotificationID, notification.OrderID, notification.UserID, notification.Message, notification.CreatedAt,
	)
	return errors.WithStack(err)
}

func (r *notificationRepository) FindForUser(userID uuid.UUID) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.client.SelectContext(r.ctx, &notifications, "SELECT * FROM notification WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return notifications, nil
}
