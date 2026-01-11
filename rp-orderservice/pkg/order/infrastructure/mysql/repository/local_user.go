package repository

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"orderservice/pkg/order/domain/model"
)

func NewLocalUserRepository(ctx context.Context, client mysql.ClientContext) model.LocalUserRepository {
	return &localUserRepository{
		ctx:    ctx,
		client: client,
	}
}

type localUserRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *localUserRepository) Store(user model.LocalUser) error {
	_, err := r.client.ExecContext(r.ctx,
		`INSERT INTO local_user (user_id, login) VALUES (?, ?) ON DUPLICATE KEY UPDATE login=VALUES(login)`,
		user.UserID, user.Login,
	)
	return errors.WithStack(err)
}

func (r *localUserRepository) Find(userID uuid.UUID) (*model.LocalUser, error) {
	var user sqlxLocalUser
	err := r.client.GetContext(r.ctx, &user, `SELECT user_id, login FROM local_user WHERE user_id = ?`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrUserNotFound)
		}
		return nil, errors.WithStack(err)
	}
	return &model.LocalUser{
		UserID: user.UserID,
		Login:  user.Login,
	}, nil
}

type sqlxLocalUser struct {
	UserID uuid.UUID `db:"user_id"`
	Login  string    `db:"login"`
}
