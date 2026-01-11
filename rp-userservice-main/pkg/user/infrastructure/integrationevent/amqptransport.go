package integrationevent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"github.com/google/uuid"

	"userservice/pkg/user/application/service"
	"userservice/pkg/user/domain/model"
	"userservice/pkg/user/infrastructure/temporal"
)

var (
	errUnhandledDelivery = errors.New("unhandled delivery")
	errProcessed         = errors.New("processed")
)

func NewAMQPTransport(logger logging.Logger, workflowService temporal.WorkflowService, userService service.UserService) AMQPTransport {
	return &amqpTransport{
		logger:          logger,
		workflowService: workflowService,
		userService:     userService,
	}
}

type AMQPTransport interface {
	Handler() amqp.Handler
}

type amqpTransport struct {
	logger          logging.Logger
	workflowService temporal.WorkflowService
	userService     service.UserService
}

func (t *amqpTransport) Handler() amqp.Handler {
	return t.withLog(t.handle)
}

func (t *amqpTransport) handle(ctx context.Context, delivery amqp.Delivery) error {
	switch delivery.Type {
	case model.UserUpdated{}.Type():
		var e UserUpdated
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			t.logger.Error(err, "failed to unmarshal UserUpdated")
			return nil
		}
		de := model.UserUpdated{
			UserID:    uuid.MustParse(e.UserID),
			UpdatedAt: time.Unix(e.UpdatedAt, 0),
		}
		if e.UpdatedFields != nil {
			de.UpdatedFields = &struct {
				Status   *model.UserStatus
				Email    *string
				Telegram *string
			}{
				Status:   (*model.UserStatus)(e.UpdatedFields.Status),
				Email:    e.UpdatedFields.Email,
				Telegram: e.UpdatedFields.Telegram,
			}
		}
		if e.RemovedFields != nil {
			de.RemovedFields = &struct {
				Email    *bool
				Telegram *bool
			}{
				Email:    e.RemovedFields.Email,
				Telegram: e.RemovedFields.Telegram,
			}
		}
		err = t.workflowService.RunUserUpdatedWorkflow(ctx, delivery.CorrelationID, de)
		if err != nil {
			return nil
		}
		return errProcessed

	case model.UserDeleted{}.Type():
		var e UserDeleted
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			t.logger.Error(err, "failed to unmarshal UserDeleted")
			return nil
		}
		if !e.Hard {
			t.logger.Info("User soft deleted, starting cleanup workflow", "user_id", e.UserID)
			err = t.workflowService.RunUserDeletedWorkflow(ctx, delivery.CorrelationID+"_del", e.UserID)
			if err != nil {
				return nil
			}
		}
		return errProcessed

	default:
		return errUnhandledDelivery
	}
}

func (t *amqpTransport) withLog(handler amqp.Handler) amqp.Handler {
	return func(ctx context.Context, delivery amqp.Delivery) error {
		l := t.logger.WithFields(logging.Fields{
			"routing_key":    delivery.RoutingKey,
			"correlation_id": delivery.CorrelationID,
			"content_type":   delivery.ContentType,
			"event_type":     delivery.Type,
		})

		if delivery.ContentType != ContentType {
			l.Warning(errors.New("invalid content type"), "skipping")
			return errProcessed
		}

		l = l.WithField("body", json.RawMessage(delivery.Body))
		start := time.Now()

		err := handler(ctx, delivery)

		l = l.WithField("duration", time.Since(start))

		if err != nil {
			if errors.Is(err, errUnhandledDelivery) {
				l.Info("unhandled delivery, skipping")
				return errProcessed
			}

			if errors.Is(err, errProcessed) {
				l.Info("successfully handled message")
				return err
			}

			l.Error(err, "failed to handle message")
			return err
		}

		l.Warning(errors.New("handler returned nil"), "triggering nack/requeue")
		return nil
	}
}
