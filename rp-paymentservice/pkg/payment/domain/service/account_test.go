package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"paymentservice/pkg/common/domain"
	"paymentservice/pkg/payment/domain/model"
)

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) NextID(userID uuid.UUID) uuid.UUID {
	args := m.Called(userID)
	return args.Get(0).(uuid.UUID)
}

func (m *MockAccountRepository) Store(account model.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAccountRepository) Find(spec model.FindSpec) (*model.Account, error) {
	args := m.Called(spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Account), args.Error(1)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestAccountService_CreateAccount(t *testing.T) {
	repo := new(MockAccountRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewAccountService(repo, dispatcher)

	userID := uuid.New()
	initialBalance := int64(1000)

	t.Run("success_new_account", func(t *testing.T) {
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(nil, model.ErrAccountNotFound).Once()
		repo.On("Store", mock.MatchedBy(func(a model.Account) bool {
			return a.UserID == userID && a.Balance == initialBalance
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.AccountCreated) bool {
			return e.UserID == userID && e.Balance == initialBalance
		})).Return(nil).Once()

		err := service.CreateAccount(userID, initialBalance)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("already_exists", func(t *testing.T) {
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(&model.Account{UserID: userID}, nil).Once()

		err := service.CreateAccount(userID, initialBalance)
		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Store")
	})
}

func TestAccountService_UpdateBalance(t *testing.T) {
	repo := new(MockAccountRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewAccountService(repo, dispatcher)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existing := &model.Account{UserID: userID, Balance: 100}
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(existing, nil).Once()
		repo.On("Store", mock.MatchedBy(func(a model.Account) bool {
			return a.UserID == userID && a.Balance == 200
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.AccountBalanceUpdated) bool {
			return e.UserID == userID && e.Balance == 200
		})).Return(nil).Once()

		err := service.UpdateBalance(userID, 200)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("no_change", func(t *testing.T) {
		existing := &model.Account{UserID: userID, Balance: 100}
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(existing, nil).Once()

		err := service.UpdateBalance(userID, 100)
		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Store")
	})
}
