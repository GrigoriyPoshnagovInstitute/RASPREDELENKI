package transport

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notificationservice/api/server/notificationinternal"
	"notificationservice/pkg/notification/application/query"
)

func NewNotificationInternalAPI(
	queryService query.NotificationQueryService,
) notificationinternal.NotificationInternalServiceServer {
	return &notificationInternalAPI{
		queryService: queryService,
	}
}

type notificationInternalAPI struct {
	queryService query.NotificationQueryService
	notificationinternal.UnimplementedNotificationInternalServiceServer
}

func (a *notificationInternalAPI) FindNotificationsForUser(ctx context.Context, request *notificationinternal.FindNotificationsForUserRequest) (*notificationinternal.FindNotificationsForUserResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user id")
	}

	notifications, err := a.queryService.FindForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responseNotifications := make([]*notificationinternal.Notification, len(notifications))
	for i, n := range notifications {
		responseNotifications[i] = &notificationinternal.Notification{
			NotificationID: n.NotificationID.String(),
			UserID:         n.UserID.String(),
			OrderID:        n.OrderID.String(),
			Message:        n.Message,
			CreatedAt:      n.CreatedAt,
		}
	}

	return &notificationinternal.FindNotificationsForUserResponse{
		Notifications: responseNotifications,
	}, nil
}
