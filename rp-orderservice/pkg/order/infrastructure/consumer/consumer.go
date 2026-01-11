package consumer

import (
	"context"
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appservice "orderservice/pkg/order/application/service"
	"orderservice/pkg/order/domain/model"
)

type EventConsumer struct {
	conn            amqp.Connection
	dataSyncService appservice.DataSyncService
	logger          logging.Logger
	ctx             context.Context
	pool            mysql.ConnectionPool
}

func NewEventConsumer(
	ctx context.Context,
	conn amqp.Connection,
	pool mysql.ConnectionPool,
	logger logging.Logger,
) (*EventConsumer, error) {
	uow := &unitOfWorkForSync{pool: pool}

	return &EventConsumer{
		conn:            conn,
		dataSyncService: appservice.NewDataSyncService(uow),
		logger:          logger,
		ctx:             ctx,
		pool:            pool,
	}, nil
}

func (c *EventConsumer) Handler() amqp.Handler {
	return c.handle
}

func (c *EventConsumer) handle(ctx context.Context, delivery amqp.Delivery) error {
	l := c.logger.WithField("event_type", delivery.Type)
	l.Info("processing event")

	switch delivery.Type {
	case "user_created", "user_updated":
		var event struct {
			UserID string `json:"user_id"`
			Login  string `json:"login"`
		}
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			l.Error(err, "failed to unmarshal user event")
			return nil
		}
		userID, parseErr := uuid.Parse(event.UserID)
		if parseErr != nil {
			l.Error(parseErr, "invalid user id in user event")
			return nil
		}

		storeErr := c.dataSyncService.SyncUser(ctx, model.LocalUser{
			UserID: userID,
			Login:  event.Login,
		})
		if storeErr != nil {
			l.Error(storeErr, "failed to sync user")
			return nil
		}
		l.Info("user synced successfully")
		return errors.New("user processed")

	case "product_created", "product_updated":
		var event struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			Price     int64  `json:"price"`
		}
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			l.Error(err, "failed to unmarshal product event")
			return nil
		}

		productID, parseErr := uuid.Parse(event.ProductID)
		if parseErr != nil {
			l.Error(parseErr, "invalid product id in product event")
			return nil
		}

		storeErr := c.dataSyncService.SyncProduct(ctx, model.LocalProduct{
			ProductID: productID,
			Name:      event.Name,
			Price:     event.Price,
		})
		if storeErr != nil {
			l.Error(storeErr, "failed to sync product")
			return nil
		}
		l.Info("product synced successfully")
		return errors.New("product processed")

	default:
		l.WithField("type", delivery.Type).Info("unhandled event type")
		return nil
	}
}
