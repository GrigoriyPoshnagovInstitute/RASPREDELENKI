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

var errUnhandledDelivery = errors.New("unhandled delivery")

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
	case model.UserCreated{}.Type():
		var e UserCreated
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			return err
		}

		if e.Telegram != nil || e.Email != nil {
			t.logger.Info("User created with contact info, auto-activating", "user_id", e.UserID)
			userID, err := uuid.Parse(e.UserID)
			if err != nil {
				return err
			}
			return t.userService.SetUserStatus(ctx, userID, int(model.Active))
		}
		return nil

	case model.UserUpdated{}.Type():
		var e UserUpdated
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			return err
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
		return t.workflowService.RunUserUpdatedWorkflow(ctx, delivery.CorrelationID, de)

	case model.UserDeleted{}.Type():
		var e UserDeleted
		err := json.Unmarshal(delivery.Body, &e)
		if err != nil {
			return err
		}
		if !e.Hard {
			t.logger.Info("User soft deleted, starting cleanup workflow", "user_id", e.UserID)
			return t.workflowService.RunUserDeletedWorkflow(ctx, delivery.CorrelationID+"_del", e.UserID)
		}
		return nil

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
		})
		if delivery.ContentType != ContentType {
			l.Warning(errors.New("invalid content type"), "skipping")
			return nil
		}
		l = l.WithField("body", json.RawMessage(delivery.Body))

		start := time.Now()
		err := handler(ctx, delivery)
		l.WithField("duration", time.Since(start))

		if err != nil {
			if errors.Is(err, errUnhandledDelivery) {
				l.Info("unhandled delivery, skipping")
				return nil
			}
			l.Error(err, "failed to handle message")
		} else {
			l.Info("successfully handled message")
		}
		return err
	}
}
