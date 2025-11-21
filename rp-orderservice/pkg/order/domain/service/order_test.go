package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"orderservice/pkg/common/domain"
	"orderservice/pkg/order/domain/model"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockOrderRepository) Store(order model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) Find(orderID uuid.UUID) (*model.Order, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestOrderService_CreateOrder(t *testing.T) {
	repo := new(MockOrderRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewOrderService(repo, dispatcher)

	userID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

	items := []model.OrderItem{
		{ProductID: productID, Quantity: 2, Price: 100},
	}

	t.Run("success", func(t *testing.T) {
		repo.On("NextID").Return(orderID, nil).Once()
		repo.On("Store", mock.MatchedBy(func(o model.Order) bool {
			return o.OrderID == orderID && o.UserID == userID && o.TotalPrice == 200
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.OrderCreated) bool {
			return e.OrderID == orderID && e.TotalPrice == 200
		})).Return(nil).Once()

		id, err := service.CreateOrder(userID, items)
		assert.NoError(t, err)
		assert.Equal(t, orderID, id)
		repo.AssertExpectations(t)
		dispatcher.AssertExpectations(t)
	})

	t.Run("empty order", func(t *testing.T) {
		_, err := service.CreateOrder(userID, []model.OrderItem{})
		assert.ErrorIs(t, err, model.ErrEmptyOrder)
	})
}

func TestOrderService_MarkAsPaid(t *testing.T) {
	repo := new(MockOrderRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewOrderService(repo, dispatcher)

	orderID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existingOrder := &model.Order{
			OrderID: orderID,
			Status:  model.StatusCreated,
		}

		repo.On("Find", orderID).Return(existingOrder, nil).Once()
		repo.On("Store", mock.MatchedBy(func(o model.Order) bool {
			return o.OrderID == orderID && o.Status == model.StatusPaid
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.OrderPaid) bool {
			return e.OrderID == orderID
		})).Return(nil).Once()

		err := service.MarkAsPaid(orderID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
		dispatcher.AssertExpectations(t)
	})

	t.Run("already paid", func(t *testing.T) {
		existingOrder := &model.Order{
			OrderID: orderID,
			Status:  model.StatusPaid,
		}
		repo.On("Find", orderID).Return(existingOrder, nil).Once()

		err := service.MarkAsPaid(orderID)
		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Store")
	})
}

func TestOrderService_CancelOrder(t *testing.T) {
	repo := new(MockOrderRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewOrderService(repo, dispatcher)

	orderID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existingOrder := &model.Order{
			OrderID: orderID,
			Status:  model.StatusCreated,
		}

		repo.On("Find", orderID).Return(existingOrder, nil).Once()
		repo.On("Store", mock.MatchedBy(func(o model.Order) bool {
			return o.OrderID == orderID && o.Status == model.StatusCancelled
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.OrderCancelled) bool {
			return e.OrderID == orderID && e.Reason == "test"
		})).Return(nil).Once()

		err := service.CancelOrder(orderID, "test")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("cannot cancel paid order", func(t *testing.T) {
		existingOrder := &model.Order{
			OrderID: orderID,
			Status:  model.StatusPaid,
		}
		repo.On("Find", orderID).Return(existingOrder, nil).Once()

		err := service.CancelOrder(orderID, "reason")
		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Store")
	})
}
