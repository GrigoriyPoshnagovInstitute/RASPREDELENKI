package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1722266007(client mysql.ClientContext) migrator.Migration {
	return &version1722266007{
		client: client,
	}
}

type version1722266007 struct {
	client mysql.ClientContext
}

func (v version1722266007) Version() int64 {
	return 1722266007
}

func (v version1722266007) Description() string {
	return "Create 'local_user' table"
}

func (v version1722266007) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE local_user
		(
			user_id       VARCHAR(64)  NOT NULL,
			login         VARCHAR(32)  NOT NULL,
			PRIMARY KEY (user_id)
		)
			ENGINE = InnoDB
			CHARACTER SET = utf8mb4
			COLLATE utf8mb4_unicode_ci;
	`)
	return errors.WithStack(err)
}
