package query

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "paymentservice/pkg/payment/application/model"
	"paymentservice/pkg/payment/application/query"
	"paymentservice/pkg/payment/domain/model"
)

func NewAccountQueryService(client mysql.ClientContext) query.AccountQueryService {
	return &accountQueryService{
		client: client,
	}
}

type accountQueryService struct {
	client mysql.ClientContext
}

func (p *accountQueryService) FindUserBalance(ctx context.Context, userID uuid.UUID) (*appmodel.UserBalance, error) {
	account := struct {
		UserID  uuid.UUID `db:"user_id"`
		Balance int64     `db:"balance"`
	}{}

	err := p.client.GetContext(
		ctx,
		&account,
		`SELECT user_id, balance FROM account WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrAccountNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &appmodel.UserBalance{
		UserID:  account.UserID,
		Balance: account.Balance,
	}, nil
}
