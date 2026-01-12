package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"paymentservice/pkg/payment/domain/model"
	"paymentservice/pkg/payment/infrastructure/metrics"
)

const (
	statusSuccess = "success"
	statusError   = "error"
)

func NewAccountRepository(ctx context.Context, client mysql.ClientContext) model.AccountRepository {
	return &accountRepository{
		ctx:    ctx,
		client: client,
	}
}

type accountRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (p *accountRepository) NextID(userID uuid.UUID) uuid.UUID {
	return userID
}

func (p *accountRepository) Store(account model.Account) (err error) {
	start := time.Now()
	defer func() {
		status := statusSuccess
		if err != nil {
			status = statusError
		}
		metrics.DatabaseDuration.WithLabelValues("store", "account", status).Observe(time.Since(start).Seconds())
	}()

	_, err = p.client.ExecContext(p.ctx,
		`
	INSERT INTO account (user_id, balance, created_at, updated_at) VALUES (?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		balance=VALUES(balance),
	    updated_at=VALUES(updated_at)
	`,
		account.UserID,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt,
	)
	return errors.WithStack(err)
}

func (p *accountRepository) Find(spec model.FindSpec) (_ *model.Account, err error) {
	start := time.Now()
	defer func() {
		status := statusSuccess
		if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, model.ErrAccountNotFound) {
			status = statusError
		}
		metrics.DatabaseDuration.WithLabelValues("find", "account", status).Observe(time.Since(start).Seconds())
	}()

	account := struct {
		UserID    uuid.UUID `db:"user_id"`
		Balance   int64     `db:"balance"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}
	query, args := p.buildSpecArgs(spec)

	err = p.client.GetContext(
		p.ctx,
		&account,
		`SELECT user_id, balance, created_at, updated_at FROM account WHERE `+query,
		args...,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrAccountNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Account{
		UserID:    account.UserID,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}, nil
}

func (p *accountRepository) buildSpecArgs(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.UserID != nil {
		parts = append(parts, "user_id = ?")
		args = append(args, *spec.UserID)
	}
	return strings.Join(parts, " AND "), args
}
