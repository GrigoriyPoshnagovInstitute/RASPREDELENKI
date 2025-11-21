package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notificationservice/pkg/notification/domain/model"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockNotificationRepository) Store(notification model.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) FindForUser(userID uuid.UUID) ([]model.Notification, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Notification), args.Error(1)
}

func TestNotificationService_CreateNotification(t *testing.T) {
	repo := new(MockNotificationRepository)
	service := NewNotificationService(repo)

	orderID := uuid.New()
	userID := uuid.New()
	notifID := uuid.New()
	msg := "test message"

	t.Run("success", func(t *testing.T) {
		repo.On("NextID").Return(notifID, nil).Once()
		repo.On("Store", mock.MatchedBy(func(n model.Notification) bool {
			return n.NotificationID == notifID && n.OrderID == orderID && n.UserID == userID && n.Message == msg
		})).Return(nil).Once()

		id, err := service.CreateNotification(orderID, userID, msg)
		assert.NoError(t, err)
		assert.Equal(t, notifID, id)
		repo.AssertExpectations(t)
	})
}
