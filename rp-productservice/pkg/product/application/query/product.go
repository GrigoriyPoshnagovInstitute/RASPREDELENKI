package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "productservice/pkg/product/application/model"
)

type ProductQueryService interface {
	FindProduct(ctx context.Context, productID uuid.UUID) (*appmodel.Product, error)
}
