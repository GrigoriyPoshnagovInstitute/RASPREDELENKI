package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notificationservice/pkg/notification/domain/model"
	"notificationservice/pkg/notification/infrastructure/metrics"
)

const (
	statusSuccess = "success"
	statusError   = "error"
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

func (r *notificationRepository) Store(notification model.Notification) (err error) {
	start := time.Now()
	defer func() {
		status := statusSuccess
		if err != nil {
			status = statusError
		}
		metrics.DatabaseDuration.WithLabelValues("store", "notification", status).Observe(time.Since(start).Seconds())
	}()

	_, err = r.client.ExecContext(r.ctx,
		`INSERT INTO notification (notification_id, order_id, user_id, message, created_at) VALUES (?, ?, ?, ?, ?)`,
		notification.NotificationID, notification.OrderID, notification.UserID, notification.Message, notification.CreatedAt,
	)
	return errors.WithStack(err)
}

func (r *notificationRepository) FindForUser(userID uuid.UUID) (_ []model.Notification, err error) {
	start := time.Now()
	defer func() {
		status := statusSuccess
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			status = statusError
		}
		metrics.DatabaseDuration.WithLabelValues("find", "notification", status).Observe(time.Since(start).Seconds())
	}()

	var notifications []model.Notification
	err = r.client.SelectContext(r.ctx, &notifications, "SELECT * FROM notification WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return notifications, nil
}
