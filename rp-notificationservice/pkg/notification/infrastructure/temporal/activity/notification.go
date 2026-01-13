package activity

import (
	"context"
	"fmt"
)

type NotificationActivities struct{}

func (a *NotificationActivities) SendOrderCreatedNotification(_ context.Context, userID, orderID string) error {
	fmt.Printf("Sending notification for Order %s to User %s\n", orderID, userID)
	return nil
}
