package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notificationservice/pkg/notification/application/service"
	"notificationservice/pkg/notification/domain/model"
	"notificationservice/pkg/notification/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) NotificationRepository(ctx context.Context) model.NotificationRepository {
	return repository.NewNotificationRepository(ctx, r.client)
}
