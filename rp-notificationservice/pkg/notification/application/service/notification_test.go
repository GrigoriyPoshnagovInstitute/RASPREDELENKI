package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainmodel "notificationservice/pkg/notification/domain/model"
)

type MockRepositoryProvider struct {
	mock.Mock
}

func (m *MockRepositoryProvider) NotificationRepository(ctx context.Context) domainmodel.NotificationRepository {
	args := m.Called(ctx)
	return args.Get(0).(domainmodel.NotificationRepository)
}

type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Execute(ctx context.Context, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx)
	provider := args.Get(0).(RepositoryProvider)
	return f(provider)
}

type StubNotifRepo struct {
	mock.Mock
}

func (m *StubNotifRepo) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StubNotifRepo) Store(n domainmodel.Notification) error {
	args := m.Called(n)
	return args.Error(0)
}

func (m *StubNotifRepo) FindForUser(_ uuid.UUID) ([]domainmodel.Notification, error) {
	return nil, nil
}

func TestNotificationService_CreateNotification(t *testing.T) {
	provider := new(MockRepositoryProvider)
	uow := new(MockUnitOfWork)
	repo := new(StubNotifRepo)

	service := NewNotificationService(uow)

	ctx := context.Background()
	orderID := uuid.New()
	userID := uuid.New()
	notifID := uuid.New()
	message := "Hello"

	// Setup Mocks
	uow.On("Execute", ctx).Return(provider)
	provider.On("NotificationRepository", ctx).Return(repo)

	repo.On("NextID").Return(notifID, nil)
	repo.On("Store", mock.MatchedBy(func(n domainmodel.Notification) bool {
		return n.NotificationID == notifID && n.Message == message && n.OrderID == orderID && n.UserID == userID
	})).Return(nil)

	id, err := service.CreateNotification(ctx, orderID, userID, message)
	assert.NoError(t, err)
	assert.Equal(t, notifID, id)
}
