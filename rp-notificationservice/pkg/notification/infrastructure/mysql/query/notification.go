package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "notificationservice/pkg/notification/application/model"
	"notificationservice/pkg/notification/application/query"
	"notificationservice/pkg/notification/infrastructure/metrics"
)

func NewNotificationQueryService(client mysql.ClientContext) query.NotificationQueryService {
	return &notificationQueryService{
		client: client,
	}
}

type notificationQueryService struct {
	client mysql.ClientContext
}

func (s *notificationQueryService) FindForUser(ctx context.Context, userID uuid.UUID) (_ []appmodel.Notification, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			status = "error"
		}
		metrics.DatabaseDuration.WithLabelValues("find_query", "notification", status).Observe(time.Since(start).Seconds())
	}()

	var notificationsData []struct {
		NotificationID uuid.UUID `db:"notification_id"`
		UserID         uuid.UUID `db:"user_id"`
		OrderID        uuid.UUID `db:"order_id"`
		Message        string    `db:"message"`
		CreatedAt      time.Time `db:"created_at"`
	}

	err = s.client.SelectContext(ctx, &notificationsData, "SELECT * FROM notification WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	notifications := make([]appmodel.Notification, len(notificationsData))
	for i, data := range notificationsData {
		notifications[i] = appmodel.Notification{
			NotificationID: data.NotificationID,
			UserID:         data.UserID,
			OrderID:        data.OrderID,
			Message:        data.Message,
			CreatedAt:      data.CreatedAt.Unix(),
		}
	}

	return notifications, nil
}
