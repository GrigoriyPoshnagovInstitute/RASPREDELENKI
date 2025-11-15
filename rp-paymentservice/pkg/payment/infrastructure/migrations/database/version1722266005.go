package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1722266005(client mysql.ClientContext) migrator.Migration {
	return &version1722266005{
		client: client,
	}
}

type version1722266005 struct {
	client mysql.ClientContext
}

func (v version1722266005) Version() int64 {
	return 1722266005
}

func (v version1722266005) Description() string {
	return "Create 'account' table"
}

func (v version1722266005) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE account
		(
			user_id       VARCHAR(64)  NOT NULL,
			balance       BIGINT       NOT NULL,
			created_at    DATETIME     NOT NULL,
			updated_at    DATETIME     NOT NULL,
			PRIMARY KEY (user_id)
		)
			ENGINE = InnoDB
			CHARACTER SET = utf8mb4
			COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
