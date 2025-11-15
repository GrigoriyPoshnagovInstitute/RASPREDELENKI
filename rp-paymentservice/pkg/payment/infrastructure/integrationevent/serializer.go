package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"paymentservice/pkg/payment/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case *model.AccountCreated:
		b, err := json.Marshal(AccountCreated{
			UserID:    e.UserID.String(),
			Balance:   e.Balance,
			CreatedAt: e.CreatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	case *model.AccountBalanceUpdated:
		b, err := json.Marshal(AccountBalanceUpdated{
			UserID:    e.UserID.String(),
			Balance:   e.Balance,
			UpdatedAt: e.UpdatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type AccountCreated struct {
	UserID    string `json:"user_id"`
	Balance   int64  `json:"balance"`
	CreatedAt int64  `json:"created_at"`
}

type AccountBalanceUpdated struct {
	UserID    string `json:"user_id"`
	Balance   int64  `json:"balance"`
	UpdatedAt int64  `json:"updated_at"`
}
