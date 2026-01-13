package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	ProductTaskQueue      = "productservice_task_queue"
	PaymentTaskQueue      = "paymentservice_task_queue"
	NotificationTaskQueue = "notificationservice_task_queue"
)

type CreateOrderParams struct {
	OrderID    string
	UserID     string
	Items      []OrderItem
	TotalPrice int64
}

type OrderItem struct {
	ProductID string
	Quantity  int
}

func CreateOrderWorkflow(ctx workflow.Context, params CreateOrderParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting CreateOrderWorkflow", "OrderID", params.OrderID)

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	// 1. Reserve Products
	ctxProduct := workflow.WithActivityOptions(ctx, options)
	ctxProduct = workflow.WithTaskQueue(ctxProduct, ProductTaskQueue)

	var reserved bool
	err := workflow.ExecuteActivity(ctxProduct, "ReserveProducts", params.Items).Get(ctxProduct, &reserved)
	if err != nil {
		logger.Error("Failed to reserve products", "Error", err)
		return err
	}

	// 2. Process Payment
	ctxPayment := workflow.WithActivityOptions(ctx, options)
	ctxPayment = workflow.WithTaskQueue(ctxPayment, PaymentTaskQueue)

	var paid bool
	err = workflow.ExecuteActivity(ctxPayment, "ProcessPayment", params.UserID, params.TotalPrice).Get(ctxPayment, &paid)
	if err != nil {
		logger.Error("Payment failed, compensating...", "Error", err)

		// Compensation: Release Products
		_ = workflow.ExecuteActivity(ctxProduct, "ReleaseProducts", params.Items).Get(ctxProduct, nil)
		return err
	}

	// 3. Send Notification
	ctxNotify := workflow.WithActivityOptions(ctx, options)
	ctxNotify = workflow.WithTaskQueue(ctxNotify, NotificationTaskQueue)

	err = workflow.ExecuteActivity(ctxNotify, "SendOrderCreatedNotification", params.UserID, params.OrderID).Get(ctxNotify, nil)
	if err != nil {
		logger.Error("Failed to send notification", "Error", err)
		// Notification failure shouldn't rollback order, just log
	}

	logger.Info("Order created successfully", "OrderID", params.OrderID)
	return nil
}
