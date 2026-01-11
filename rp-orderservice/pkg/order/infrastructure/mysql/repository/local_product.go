package repository

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
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
	var product sqlxProduct
	err := r.client.GetContext(r.ctx, &product, `SELECT product_id, name, price FROM local_product WHERE product_id = ?`, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}
	return &model.LocalProduct{
		ProductID: product.ProductID,
		Name:      product.Name,
		Price:     product.Price,
	}, nil
}

func (r *localProductRepository) FindMany(productIDs []uuid.UUID) ([]model.LocalProduct, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	products := make([]model.LocalProduct, 0, len(productIDs))

	for _, productID := range productIDs {
		var product sqlxProduct
		err := r.client.GetContext(r.ctx, &product,
			`SELECT product_id, name, price FROM local_product WHERE product_id = ?`,
			productID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		products = append(products, model.LocalProduct{
			ProductID: product.ProductID,
			Name:      product.Name,
			Price:     product.Price,
		})
	}

	return products, nil
}

type sqlxProduct struct {
	ProductID uuid.UUID `db:"product_id"`
	Name      string    `db:"name"`
	Price     int64     `db:"price"`
}
