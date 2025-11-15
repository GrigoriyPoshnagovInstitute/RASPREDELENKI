package service

import (
	"context"

	"orderservice/pkg/order/domain/model"
)

type RepositoryProvider interface {
	OrderRepository(ctx context.Context) model.OrderRepository
	LocalUserRepository(ctx context.Context) model.LocalUserRepository
	LocalProductRepository(ctx context.Context) model.LocalProductRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}

type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
