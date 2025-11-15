package service

import (
	"context"

	"productservice/pkg/product/domain/model"
)

type RepositoryProvider interface {
	ProductRepository(ctx context.Context) model.ProductRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
