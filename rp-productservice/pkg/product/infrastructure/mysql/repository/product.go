package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"productservice/pkg/product/domain/model"
)

func NewProductRepository(ctx context.Context, client mysql.ClientContext) model.ProductRepository {
	return &productRepository{
		ctx:    ctx,
		client: client,
	}
}

type productRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (p *productRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (p *productRepository) Store(product model.Product) error {
	_, err := p.client.ExecContext(p.ctx,
		`
	INSERT INTO product (product_id, name, description, price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		name=VALUES(name),
	    description=VALUES(description),
	    price=VALUES(price),
	    updated_at=VALUES(updated_at)
	`,
		product.ProductID,
		product.Name,
		toSQLNull(product.Description),
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
	)
	return errors.WithStack(err)
}

func (p *productRepository) Find(spec model.FindSpec) (*model.Product, error) {
	product := struct {
		ProductID   uuid.UUID        `db:"product_id"`
		Name        string           `db:"name"`
		Description sql.Null[string] `db:"description"`
		Price       int64            `db:"price"`
		CreatedAt   time.Time        `db:"created_at"`
		UpdatedAt   time.Time        `db:"updated_at"`
	}{}
	query, args := p.buildSpecArgs(spec)

	err := p.client.GetContext(
		p.ctx,
		&product,
		`SELECT product_id, name, description, price, created_at, updated_at FROM product WHERE `+query,
		args...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Product{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: fromSQLNull(product.Description),
		Price:       product.Price,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}, nil
}

func (p *productRepository) Delete(productID uuid.UUID) error {
	_, err := p.client.ExecContext(p.ctx, `DELETE FROM product WHERE product_id = ?`, productID)
	return errors.WithStack(err)
}

func (p *productRepository) buildSpecArgs(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.ProductID != nil {
		parts = append(parts, "product_id = ?")
		args = append(args, *spec.ProductID)
	}
	if spec.Name != nil {
		parts = append(parts, "name = ?")
		args = append(args, *spec.Name)
	}
	return strings.Join(parts, " AND "), args
}

func fromSQLNull[T any](v sql.Null[T]) *T {
	if v.Valid {
		return &v.V
	}
	return nil
}

func toSQLNull[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{
		V:     *v,
		Valid: true,
	}
}
