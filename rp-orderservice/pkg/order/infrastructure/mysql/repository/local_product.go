package repository

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"orderservice/pkg/order/domain/model"
)

func NewLocalProductRepository(ctx context.Context, client mysql.ClientContext) model.LocalProductRepository {
	return &localProductRepository{
		ctx:    ctx,
		client: client,
	}
}

type localProductRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *localProductRepository) Store(product model.LocalProduct) error {
	_, err := r.client.ExecContext(r.ctx,
		`INSERT INTO local_product (product_id, name, price) VALUES (?, ?, ?) 
		 ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price)`,
		product.ProductID, product.Name, product.Price,
	)
	return errors.WithStack(err)
}

func (r *localProductRepository) Find(productID uuid.UUID) (*model.LocalProduct, error) {
	var product model.LocalProduct
	err := r.client.GetContext(r.ctx, &product, `SELECT product_id, name, price FROM local_product WHERE product_id = ?`, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}
	return &product, nil
}

func (r *localProductRepository) FindMany(productIDs []uuid.UUID) ([]model.LocalProduct, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`SELECT product_id, name, price FROM local_product WHERE product_id IN (?)`, productIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	querier, ok := r.client.(sqlx.ExtContext)
	if !ok {
		return nil, errors.New("client does not implement sqlx.ExtContext")
	}
	query = querier.Rebind(query)

	var products []model.LocalProduct
	err = r.client.SelectContext(r.ctx, &products, query, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return products, nil
}
