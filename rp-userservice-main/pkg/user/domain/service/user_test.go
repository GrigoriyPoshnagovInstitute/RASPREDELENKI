package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"userservice/pkg/common/domain"
	"userservice/pkg/user/domain/model"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Store(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Find(spec model.FindSpec) (*model.User, error) {
	args := m.Called(spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) HardDelete(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewUserService(repo, dispatcher)

	login := "testuser"
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo.On("Find", model.FindSpec{Login: &login}).Return(nil, model.ErrUserNotFound).Once()
		repo.On("NextID").Return(userID, nil).Once()
		repo.On("Store", mock.MatchedBy(func(u model.User) bool {
			return u.Login == login && u.Status == model.Active
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.UserCreated) bool {
			return e.UserID == userID
		})).Return(nil).Once()

		id, err := service.CreateUser(model.Active, login)
		assert.NoError(t, err)
		assert.Equal(t, userID, id)
		repo.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		repo.On("Find", model.FindSpec{Login: &login}).Return(&model.User{}, nil).Once()

		_, err := service.CreateUser(model.Active, login)
		assert.ErrorIs(t, err, model.ErrUserLoginAlreadyUsed)
	})
}

func TestUserService_UpdateUserEmail(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewUserService(repo, dispatcher)

	userID := uuid.New()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		existing := &model.User{UserID: userID, Login: "test"}
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(existing, nil).Once()
		repo.On("Find", model.FindSpec{Email: &email}).Return(nil, model.ErrUserNotFound).Once()
		repo.On("Store", mock.MatchedBy(func(u model.User) bool {
			return u.Email != nil && *u.Email == email
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.UserUpdated) bool {
			return e.UserID == userID && e.UpdatedFields.Email != nil && *e.UpdatedFields.Email == email
		})).Return(nil).Once()

		err := service.UpdateUserEmail(userID, &email)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("remove_email", func(t *testing.T) {
		existing := &model.User{UserID: userID, Login: "test", Email: &email}
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(existing, nil).Once()
		repo.On("Store", mock.MatchedBy(func(u model.User) bool {
			return u.Email == nil
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.UserUpdated) bool {
			return e.UserID == userID && e.RemovedFields.Email != nil && *e.RemovedFields.Email == true
		})).Return(nil).Once()

		err := service.UpdateUserEmail(userID, nil)
		assert.NoError(t, err)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewUserService(repo, dispatcher)

	userID := uuid.New()

	t.Run("soft_delete", func(t *testing.T) {
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(&model.User{UserID: userID}, nil).Once()
		repo.On("Store", mock.MatchedBy(func(u model.User) bool {
			return u.Status == model.Deleted && u.DeletedAt != nil
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.UserDeleted) bool {
			return e.UserID == userID && !e.Hard
		})).Return(nil).Once()

		err := service.DeleteUser(userID, false)
		assert.NoError(t, err)
	})

	t.Run("hard_delete", func(t *testing.T) {
		repo.On("Find", model.FindSpec{UserID: &userID}).Return(&model.User{UserID: userID}, nil).Once()
		repo.On("HardDelete", userID).Return(nil).Once()
		repo.On("Store", mock.MatchedBy(func(u model.User) bool {
			return u.Status == model.Deleted
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.UserDeleted) bool {
			return e.UserID == userID && e.Hard
		})).Return(nil).Once()

		err := service.DeleteUser(userID, true)
		assert.NoError(t, err)
	})
}
