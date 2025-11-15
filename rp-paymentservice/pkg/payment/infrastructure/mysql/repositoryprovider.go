package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"paymentservice/pkg/payment/application/service"
	"paymentservice/pkg/payment/domain/model"
	"paymentservice/pkg/payment/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) AccountRepository(ctx context.Context) model.AccountRepository {
	return repository.NewAccountRepository(ctx, r.client)
}
