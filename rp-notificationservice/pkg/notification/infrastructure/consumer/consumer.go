package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appservice "notificationservice/pkg/notification/application/service"
)

const (
	queueName       = "notification_events"
	exchangeName    = "domain_event_exchange"
	orderRoutingKey = "order.*"
)

type EventConsumer struct {
	conn                amqp.Connection
	notificationService appservice.NotificationService
	logger              logging.Logger
	ctx                 context.Context
}

func NewEventConsumer(
	ctx context.Context,
	conn amqp.Connection,
	pool mysql.ConnectionPool,
	logger logging.Logger,
) (*EventConsumer, error) {
	uow := &unitOfWorkForSync{pool: pool}

	return &EventConsumer{
		conn:                conn,
		notificationService: appservice.NewNotificationService(uow),
		logger:              logger,
		ctx:                 ctx,
	}, nil
}

func (c *EventConsumer) Start() error {
	_ = c.conn.Consumer(
		c.ctx,
		c.handle,
		&amqp.QueueConfig{
			Name:    queueName,
			Durable: true,
		},
		&amqp.BindConfig{
			ExchangeName: exchangeName,
			RoutingKeys:  []string{orderRoutingKey},
		},
		nil,
	)
	<-c.ctx.Done()
	return nil
}

func (c *EventConsumer) handle(ctx context.Context, delivery amqp.Delivery) error {
	l := c.logger.WithField("event_type", delivery.Type)
	l.Info("processing event")

	var err error
	var orderID, userID uuid.UUID
	var message string

	switch delivery.Type {
	case "order_created":
		var event struct {
			OrderID string `json:"order_id"`
			UserID  string `json:"user_id"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal order_created")
			break
		}
		orderID, _ = uuid.Parse(event.OrderID)
		userID, _ = uuid.Parse(event.UserID)
		message = fmt.Sprintf("Order #%s has been created.", orderID.String())

	case "order_paid":
		var event struct {
			OrderID string `json:"order_id"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal order_paid")
			break
		}
		orderID, _ = uuid.Parse(event.OrderID)
		// UserID здесь неизвестен, в реальной системе его нужно было бы доставать из заказа
		// Для простоты мы оставим его пустым или вам нужно будет добавить его в событие OrderPaid
		userID = uuid.Nil
		message = fmt.Sprintf("Order #%s has been paid successfully.", orderID.String())

	case "order_cancelled":
		var event struct {
			OrderID string `json:"order_id"`
			Reason  string `json:"reason"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal order_cancelled")
			break
		}
		orderID, _ = uuid.Parse(event.OrderID)
		userID = uuid.Nil // Аналогично order_paid
		message = fmt.Sprintf("Order #%s has been cancelled. Reason: %s", orderID.String(), event.Reason)

	default:
		l.WithField("type", delivery.Type).Info("unhandled event type")
		return nil
	}

	if err != nil {
		l.Error(err, "failed to process event payload")
		return err
	}

	if orderID != uuid.Nil {
		_, createErr := c.notificationService.CreateNotification(ctx, orderID, userID, message)
		if createErr != nil {
			err = createErr // Сохраняем ошибку, если не удалось создать уведомление
			l.Error(err, "failed to create notification")
		}
	}

	return err
}
