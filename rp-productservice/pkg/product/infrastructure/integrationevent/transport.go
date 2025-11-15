package integrationevent

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
)

const (
	TransportName    = "domain"
	ExchangeName     = "domain_event_exchange"
	ExchangeKind     = "topic"
	RoutingKeyPrefix = "product.go."
	ContentType      = "application/json"
)

func NewTransport(logger logging.Logger, producer amqp.Producer) outbox.Transport {
	return &transport{
		logger:   logger,
		producer: producer,
	}
}

type transport struct {
	logger   logging.Logger
	producer amqp.Producer
}

func (t *transport) HandleEvents(ctx context.Context, correlationID, eventType, payload string) error {
	l := t.logger.WithFields(logging.Fields{
		"correlationID": correlationID,
		"eventType":     eventType,
		"payload":       payload,
	})

	err := t.producer.Publish(ctx, amqp.Delivery{
		RoutingKey:    RoutingKeyPrefix + eventType,
		CorrelationID: correlationID,
		ContentType:   ContentType,
		Type:          eventType,
		Body:          []byte(payload),
	})
	if err != nil {
		l.Error(err, "failed to publish event")
		return err
	}
	l.Info("successfully published event")
	return nil
}
