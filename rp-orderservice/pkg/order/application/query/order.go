package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "orderservice/pkg/order/application/model"
)

type OrderQueryService interface {
	FindOrder(ctx context.Context, orderID uuid.UUID) (*appmodel.Order, error)
}
