package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"orderservice/pkg/order/application/service"
	"orderservice/pkg/order/domain/model"
	"orderservice/pkg/order/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) OrderRepository(ctx context.Context) model.OrderRepository {
	return repository.NewOrderRepository(ctx, r.client)
}

func (r *repositoryProvider) LocalUserRepository(ctx context.Context) model.LocalUserRepository {
	return repository.NewLocalUserRepository(ctx, r.client)
}

func (r *repositoryProvider) LocalProductRepository(ctx context.Context) model.LocalProductRepository {
	return repository.NewLocalProductRepository(ctx, r.client)
}
