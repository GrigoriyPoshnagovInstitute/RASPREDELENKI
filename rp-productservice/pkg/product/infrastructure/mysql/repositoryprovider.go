package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"productservice/pkg/product/application/service"
	"productservice/pkg/product/domain/model"
	"productservice/pkg/product/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) ProductRepository(ctx context.Context) model.ProductRepository {
	return repository.NewProductRepository(ctx, r.client)
}
