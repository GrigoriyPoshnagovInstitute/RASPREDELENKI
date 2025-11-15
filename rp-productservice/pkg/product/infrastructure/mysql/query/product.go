package query

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "productservice/pkg/product/application/model"
	"productservice/pkg/product/application/query"
	"productservice/pkg/product/domain/model"
)

func NewProductQueryService(client mysql.ClientContext) query.ProductQueryService {
	return &productQueryService{
		client: client,
	}
}

type productQueryService struct {
	client mysql.ClientContext
}

func (p *productQueryService) FindProduct(ctx context.Context, productID uuid.UUID) (*appmodel.Product, error) {
	product := struct {
		ProductID   uuid.UUID        `db:"product_id"`
		Name        string           `db:"name"`
		Description sql.Null[string] `db:"description"`
		Price       int64            `db:"price"`
	}{}

	err := p.client.GetContext(
		ctx,
		&product,
		`SELECT product_id, name, description, price FROM product.go WHERE product_id = ?`,
		productID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &appmodel.Product{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: fromSQLNull(product.Description),
		Price:       product.Price,
	}, nil
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}
