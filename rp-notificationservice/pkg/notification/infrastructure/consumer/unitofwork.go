package consumer

import (
	"context"

	"notificationservice/pkg/notification/application/service"
	inframysql "notificationservice/pkg/notification/infrastructure/mysql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
)

type unitOfWorkForSync struct {
	pool mysql.ConnectionPool
}

func (u *unitOfWorkForSync) Execute(ctx context.Context, f func(provider service.RepositoryProvider) error) error {
	uow := mysql.NewUnitOfWork(u.pool, inframysql.NewRepositoryProvider)
	return uow.ExecuteWithRepositoryProvider(ctx, f)
}
