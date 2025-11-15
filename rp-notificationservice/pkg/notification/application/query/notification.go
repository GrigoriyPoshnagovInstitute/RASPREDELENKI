package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "notificationservice/pkg/notification/application/model"
)

type NotificationQueryService interface {
	FindForUser(ctx context.Context, userID uuid.UUID) ([]appmodel.Notification, error)
}
