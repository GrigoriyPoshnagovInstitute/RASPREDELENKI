package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"paymentservice/pkg/common/domain"
	"paymentservice/pkg/payment/domain/model"
)

type AccountService interface {
	CreateAccount(userID uuid.UUID, initialBalance int64) error
	UpdateBalance(userID uuid.UUID, newBalance int64) error
}

func NewAccountService(
	accountRepository model.AccountRepository,
	eventDispatcher domain.EventDispatcher,
) AccountService {
	return &accountService{
		accountRepository: accountRepository,
		eventDispatcher:   eventDispatcher,
	}
}

type accountService struct {
	accountRepository model.AccountRepository
	eventDispatcher   domain.EventDispatcher
}

func (s *accountService) CreateAccount(userID uuid.UUID, initialBalance int64) error {
	_, err := s.accountRepository.Find(model.FindSpec{UserID: &userID})
	if err == nil {
		// Аккаунт уже существует, ничего не делаем
		return nil
	}
	if !errors.Is(err, model.ErrAccountNotFound) {
		return err
	}

	currentTime := time.Now()
	account := model.Account{
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	err = s.accountRepository.Store(account)
	if err != nil {
		return err
	}

	return s.eventDispatcher.Dispatch(&model.AccountCreated{
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: currentTime,
	})
}

func (s *accountService) UpdateBalance(userID uuid.UUID, newBalance int64) error {
	account, err := s.accountRepository.Find(model.FindSpec{UserID: &userID})
	if err != nil {
		return err
	}

	if account.Balance == newBalance {
		return nil
	}

	currentTime := time.Now()
	account.Balance = newBalance
	account.UpdatedAt = currentTime

	err = s.accountRepository.Store(*account)
	if err != nil {
		return err
	}

	return s.eventDispatcher.Dispatch(&model.AccountBalanceUpdated{
		UserID:    userID,
		Balance:   newBalance,
		UpdatedAt: currentTime,
	})
}
