package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1722266004(client mysql.ClientContext) migrator.Migration {
	return &version1722266004{
		client: client,
	}
}

type version1722266004 struct {
	client mysql.ClientContext
}

func (v version1722266004) Version() int64 {
	return 1722266004
}

func (v version1722266004) Description() string {
	return "Create 'product' table"
}

func (v version1722266004) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE product
		(
			product_id    VARCHAR(64)  NOT NULL,
			name          VARCHAR(255) NOT NULL,
			description   TEXT,
			price         BIGINT       NOT NULL,
			created_at    DATETIME     NOT NULL,
			updated_at    DATETIME     NOT NULL,
			PRIMARY KEY (product_id),
			UNIQUE INDEX product_name_uidx (name)
		)
			ENGINE = InnoDB
			CHARACTER SET = utf8mb4
			COLLATE utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
