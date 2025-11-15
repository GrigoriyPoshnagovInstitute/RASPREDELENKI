package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notificationservice/pkg/notification/application/service"
)

func NewUnitOfWork(uow mysql.UnitOfWorkWithRepositoryProvider[service.RepositoryProvider]) service.UnitOfWork {
	return &unitOfWork{
		uow: uow,
	}
}

type unitOfWork struct {
	uow mysql.UnitOfWorkWithRepositoryProvider[service.RepositoryProvider]
}

func (u *unitOfWork) Execute(ctx context.Context, f func(provider service.RepositoryProvider) error) error {
	return u.uow.ExecuteWithRepositoryProvider(ctx, f)
}
