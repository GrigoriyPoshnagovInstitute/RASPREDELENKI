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

const (
	queueName         = "order_events"
	exchangeName      = "domain_event_exchange"
	userRoutingKey    = "user.*"
	productRoutingKey = "product.*"
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
			RoutingKeys:  []string{userRoutingKey, productRoutingKey},
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
	switch delivery.Type {
	case "user_created", "user_updated":
		var event struct {
			UserID string `json:"user_id"`
			Login  string `json:"login"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal user event")
			break
		}
		userID, parseErr := uuid.Parse(event.UserID)
		if parseErr != nil {
			err = errors.Wrap(parseErr, "invalid user id in user event")
			break
		}
		err = c.dataSyncService.SyncUser(ctx, model.LocalUser{
			UserID: userID,
			Login:  event.Login,
		})

	case "product_created", "product_updated":
		var event struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			Price     int64  `json:"price"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal product event")
			break
		}
		productID, parseErr := uuid.Parse(event.ProductID)
		if parseErr != nil {
			err = errors.Wrap(parseErr, "invalid product id in product event")
			break
		}
		err = c.dataSyncService.SyncProduct(ctx, model.LocalProduct{
			ProductID: productID,
			Name:      event.Name,
			Price:     event.Price,
		})

	default:
		l.WithField("type", delivery.Type).Info("unhandled event type")
	}

	if err != nil {
		l.Error(err, "failed to process event")
	}

	return err
}
