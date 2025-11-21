package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	appmodel "userservice/pkg/user/application/model"
	domainmodel "userservice/pkg/user/domain/model"
)

type MockRepositoryProvider struct {
	mock.Mock
}

func (m *MockRepositoryProvider) UserRepository(ctx context.Context) domainmodel.UserRepository {
	args := m.Called(ctx)
	return args.Get(0).(domainmodel.UserRepository)
}

type MockLockableUnitOfWork struct {
	mock.Mock
}

func (m *MockLockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx, lockNames)
	provider := args.Get(0).(RepositoryProvider)
	return f(provider)
}

type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Execute(ctx context.Context, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx)
	provider := args.Get(0).(RepositoryProvider)
	return f(provider)
}

type StubUserRepository struct {
	mock.Mock
}

func (m *StubUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StubUserRepository) Store(u domainmodel.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *StubUserRepository) Find(spec domainmodel.FindSpec) (*domainmodel.User, error) {
	args := m.Called(spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodel.User), args.Error(1)
}

func (m *StubUserRepository) HardDelete(id uuid.UUID) error {
	return nil
}

type DummyDispatcher struct{}

func (d *DummyDispatcher) Dispatch(ctx context.Context, event outbox.Event) error {
	return nil
}

func TestUserService_StoreUser_Create(t *testing.T) {
	provider := new(MockRepositoryProvider)
	luow := new(MockLockableUnitOfWork)
	uow := new(MockUnitOfWork)

	service := NewUserService(uow, luow, &DummyDispatcher{})

	repo := new(StubUserRepository)

	ctx := context.Background()
	userID := uuid.New()
	login := "new_user"
	email := "test@email.com"

	cmd := appmodel.User{
		Login:  login,
		Email:  &email,
		Status: int(domainmodel.Active),
	}

	luow.On("Execute", ctx, mock.AnythingOfType("[]string")).Return(provider)

	provider.On("UserRepository", ctx).Return(repo)

	repo.On("Find", domainmodel.FindSpec{Login: &login}).Return(nil, domainmodel.ErrUserNotFound)
	repo.On("NextID").Return(userID, nil)
	repo.On("Store", mock.MatchedBy(func(u domainmodel.User) bool {
		return u.Login == login
	})).Return(nil)

	repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(&domainmodel.User{UserID: userID, Login: login}, nil)
	repo.On("Find", domainmodel.FindSpec{Email: &email}).Return(nil, domainmodel.ErrUserNotFound)
	repo.On("Store", mock.MatchedBy(func(u domainmodel.User) bool {
		return u.Email != nil && *u.Email == email
	})).Return(nil)

	repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(&domainmodel.User{UserID: userID, Login: login, Email: &email}, nil)

	id, err := service.StoreUser(ctx, cmd)
	assert.NoError(t, err)
	assert.Equal(t, userID, id)
}

func TestUserService_FindUser(t *testing.T) {
	provider := new(MockRepositoryProvider)
	luow := new(MockLockableUnitOfWork)
	service := NewUserService(nil, luow, &DummyDispatcher{})
	repo := new(StubUserRepository)

	ctx := context.Background()
	userID := uuid.New()
	login := "found_user"

	luow.On("Execute", ctx, mock.Anything).Return(provider)
	provider.On("UserRepository", ctx).Return(repo)

	repo.On("Find", domainmodel.FindSpec{UserID: &userID}).Return(&domainmodel.User{
		UserID: userID,
		Login:  login,
		Status: domainmodel.Active,
	}, nil)

	u, err := service.FindUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, login, u.Login)
	assert.Equal(t, int(domainmodel.Active), u.Status)
}
