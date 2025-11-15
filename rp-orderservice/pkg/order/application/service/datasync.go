package service

import (
	"context"

	"orderservice/pkg/order/domain/model"
)

type DataSyncService interface {
	SyncUser(ctx context.Context, user model.LocalUser) error
	SyncProduct(ctx context.Context, product model.LocalProduct) error
}

func NewDataSyncService(uow UnitOfWork) DataSyncService {
	return &dataSyncService{uow: uow}
}

type dataSyncService struct {
	uow UnitOfWork
}

func (s *dataSyncService) SyncUser(ctx context.Context, user model.LocalUser) error {
	return s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		return provider.LocalUserRepository(ctx).Store(user)
	})
}

func (s *dataSyncService) SyncProduct(ctx context.Context, product model.LocalProduct) error {
	return s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		return provider.LocalProductRepository(ctx).Store(product)
	})
}
