package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "paymentservice/pkg/payment/application/model"
)

type AccountQueryService interface {
	FindUserBalance(ctx context.Context, userID uuid.UUID) (*appmodel.UserBalance, error)
}
