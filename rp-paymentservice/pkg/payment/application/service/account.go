package service

import (
	"context"
	"errors"
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"paymentservice/pkg/common/domain"
	appmodel "paymentservice/pkg/payment/application/model"
	"paymentservice/pkg/payment/domain/model"
	"paymentservice/pkg/payment/domain/service"
)

type AccountService interface {
	StoreUserBalance(ctx context.Context, balance appmodel.UserBalance) error
}

func NewAccountService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) AccountService {
	return &accountService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type accountService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *accountService) StoreUserBalance(ctx context.Context, balance appmodel.UserBalance) error {
	lockName := userBalanceLock(balance.UserID)

	return s.luow.Execute(ctx, []string{lockName}, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.AccountRepository(ctx))

		_, err := provider.AccountRepository(ctx).Find(model.FindSpec{UserID: &balance.UserID})
		if errors.Is(err, model.ErrAccountNotFound) {
			return domainService.CreateAccount(balance.UserID, balance.Balance)
		}
		if err != nil {
			return err
		}

		return domainService.UpdateBalance(balance.UserID, balance.Balance)
	})
}

func (s *accountService) domainService(ctx context.Context, repository model.AccountRepository) service.AccountService {
	return service.NewAccountService(repository, s.domainEventDispatcher(ctx))
}

func (s *accountService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseUserBalanceLock = "user_balance_"

func userBalanceLock(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", baseUserBalanceLock, id.String())
}
