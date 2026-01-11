package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	appmodel "paymentservice/pkg/payment/application/model"
	domainmodel "paymentservice/pkg/payment/domain/model"
)

type MockRepositoryProvider struct {
	mock.Mock
}

func (m *MockRepositoryProvider) AccountRepository(ctx context.Context) domainmodel.AccountRepository {
	return m.Called(ctx).Get(0).(domainmodel.AccountRepository)
}

type MockLockableUnitOfWork struct {
	mock.Mock
}

func (m *MockLockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx, lockNames)
	provider := args.Get(0).(RepositoryProvider)
	return f(provider)
}

type StubAccountRepo struct {
	mock.Mock
}

func (m *StubAccountRepo) NextID(userID uuid.UUID) uuid.UUID {
	return userID
}

func (m *StubAccountRepo) Store(a domainmodel.Account) error {
	return m.Called(a).Error(0)
}

func (m *StubAccountRepo) Find(s domainmodel.FindSpec) (*domainmodel.Account, error) {
	args := m.Called(s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodel.Account), args.Error(1)
}

type DummyDispatcher struct{}

func (d *DummyDispatcher) Dispatch(_ context.Context, _ outbox.Event) error {
	return nil
}

func TestAccountService_StoreUserBalance(t *testing.T) {
	provider := new(MockRepositoryProvider)
	luow := new(MockLockableUnitOfWork)
	repo := new(StubAccountRepo)

	service := NewAccountService(nil, luow, &DummyDispatcher{})

	ctx := context.Background()
	userID := uuid.New()
	balance := int64(5000)

	cmd := appmodel.UserBalance{
		UserID:  userID,
		Balance: balance,
	}

	t.Run("create_if_not_exists", func(t *testing.T) {
		luow.On("Execute", ctx, mock.Anything).Return(provider)
		provider.On("AccountRepository", ctx).Return(repo)

		repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(nil, domainmodel.ErrAccountNotFound).Once()

		repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(nil, domainmodel.ErrAccountNotFound).Once()
		repo.On("Store", mock.MatchedBy(func(a domainmodel.Account) bool {
			return a.UserID == userID && a.Balance == balance
		})).Return(nil)

		err := service.StoreUserBalance(ctx, cmd)
		assert.NoError(t, err)
	})

	t.Run("update_if_exists", func(t *testing.T) {
		luow.On("Execute", ctx, mock.Anything).Return(provider)
		provider.On("AccountRepository", ctx).Return(repo)

		existing := &domainmodel.Account{UserID: userID, Balance: 100}

		repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(existing, nil).Once()

		repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(existing, nil).Once()
		repo.On("Store", mock.MatchedBy(func(a domainmodel.Account) bool {
			return a.Balance == balance
		})).Return(nil)

		err := service.StoreUserBalance(ctx, cmd)
		assert.NoError(t, err)
	})
}
