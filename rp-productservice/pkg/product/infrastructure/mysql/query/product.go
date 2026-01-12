package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "productservice/pkg/product/application/model"
	"productservice/pkg/product/application/query"
	"productservice/pkg/product/domain/model"
	"productservice/pkg/product/infrastructure/metrics"
)

func NewProductQueryService(client mysql.ClientContext) query.ProductQueryService {
	return &productQueryService{
		client: client,
	}
}

type productQueryService struct {
	client mysql.ClientContext
}

func (p *productQueryService) FindProduct(ctx context.Context, productID uuid.UUID) (_ *appmodel.Product, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, model.ErrProductNotFound) {
			status = "error"
		}
		metrics.DatabaseDuration.WithLabelValues("find_query", "product", status).Observe(time.Since(start).Seconds())
	}()

	product := struct {
		ProductID   uuid.UUID        `db:"product_id"`
		Name        string           `db:"name"`
		Description sql.Null[string] `db:"description"`
		Price       int64            `db:"price"`
	}{}

	err = p.client.GetContext(
		ctx,
		&product,
		`SELECT product_id, name, description, price FROM product WHERE product_id = ?`,
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
