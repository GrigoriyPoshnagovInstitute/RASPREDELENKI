package service

import (
	"context"

	"paymentservice/pkg/payment/domain/model"
)

type RepositoryProvider interface {
	AccountRepository(ctx context.Context) model.AccountRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
