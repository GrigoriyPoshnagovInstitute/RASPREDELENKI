package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1722266009(client mysql.ClientContext) migrator.Migration {
	return &version1722266009{
		client: client,
	}
}

type version1722266009 struct {
	client mysql.ClientContext
}

func (v version1722266009) Version() int64 {
	return 1722266009
}

func (v version1722266009) Description() string {
	return "Create 'notification' table"
}

func (v version1722266009) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE notification
		(
			notification_id VARCHAR(64)  NOT NULL,
			order_id        VARCHAR(64)  NOT NULL,
			user_id         VARCHAR(64)  NOT NULL,
			message         TEXT         NOT NULL,
			created_at      DATETIME     NOT NULL,
			PRIMARY KEY (notification_id),
			INDEX notification_user_id_idx (user_id)
		)
			ENGINE = InnoDB
			CHARACTER SET = utf8mb4
			COLLATE utf8mb4_unicode_ci;
	`)
	return errors.WithStack(err)
}
