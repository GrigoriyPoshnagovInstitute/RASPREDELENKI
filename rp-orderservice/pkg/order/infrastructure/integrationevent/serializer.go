package integrationevent

import (
	"encoding/json"

	"github.com/pkg/errors"

	"orderservice/pkg/order/domain/model"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case *model.OrderCreated:
		items := make([]OrderItem, len(e.Items))
		for i, item := range e.Items {
			items[i] = OrderItem{
				ProductID: item.ProductID.String(),
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
		}
		b, err := json.Marshal(OrderCreated{
			OrderID:    e.OrderID.String(),
			UserID:     e.UserID.String(),
			TotalPrice: e.TotalPrice,
			Items:      items,
			CreatedAt:  e.CreatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)

	case *model.OrderPaid:
		b, err := json.Marshal(OrderPaid{
			OrderID: e.OrderID.String(),
			PaidAt:  e.PaidAt.Unix(),
		})
		return string(b), errors.WithStack(err)

	case *model.OrderCancelled:
		b, err := json.Marshal(OrderCancelled{
			OrderID:     e.OrderID.String(),
			Reason:      e.Reason,
			CancelledAt: e.CancelledAt.Unix(),
		})
		return string(b), errors.WithStack(err)

	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     int64  `json:"price"`
}

type OrderCreated struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	TotalPrice int64       `json:"total_price"`
	Items      []OrderItem `json:"items"`
	CreatedAt  int64       `json:"created_at"`
}

type OrderPaid struct {
	OrderID string `json:"order_id"`
	PaidAt  int64  `json:"paid_at"`
}

type OrderCancelled struct {
	OrderID     string `json:"order_id"`
	Reason      string `json:"reason"`
	CancelledAt int64  `json:"cancelled_at"`
}
