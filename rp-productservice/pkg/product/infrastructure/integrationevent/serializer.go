package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"productservice/pkg/product/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case *model.ProductCreated:
		b, err := json.Marshal(ProductCreated{
			ProductID:   e.ProductID.String(),
			Name:        e.Name,
			Description: e.Description,
			Price:       e.Price,
			CreatedAt:   e.CreatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	case *model.ProductUpdated:
		ie := ProductUpdated{
			ProductID: e.ProductID.String(),
			UpdatedAt: e.UpdatedAt.Unix(),
		}
		if e.UpdatedFields.Name != nil {
			ie.UpdatedFields.Name = e.UpdatedFields.Name
		}
		if e.UpdatedFields.Description != nil {
			ie.UpdatedFields.Description = e.UpdatedFields.Description
		}
		if e.UpdatedFields.Price != nil {
			ie.UpdatedFields.Price = e.UpdatedFields.Price
		}
		b, err := json.Marshal(ie)
		return string(b), errors.WithStack(err)

	case *model.ProductDeleted:
		b, err := json.Marshal(ProductDeleted{
			ProductID: e.ProductID.String(),
			DeletedAt: e.DeletedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type ProductCreated struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Price       int64   `json:"price"`
	CreatedAt   int64   `json:"created_at"`
}

type ProductUpdated struct {
	ProductID     string `json:"product_id"`
	UpdatedFields struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Price       *int64  `json:"price,omitempty"`
	} `json:"updated_fields,omitempty"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

type ProductDeleted struct {
	ProductID string `json:"product_id"`
	DeletedAt int64  `json:"deleted_at"`
}
