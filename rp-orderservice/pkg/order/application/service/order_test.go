package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/client"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	"orderservice/pkg/order/application/model"
	domainmodel "orderservice/pkg/order/domain/model"
)

type MockRepositoryProvider struct {
	mock.Mock
}

func (m *MockRepositoryProvider) OrderRepository(ctx context.Context) domainmodel.OrderRepository {
	args := m.Called(ctx)
	return args.Get(0).(domainmodel.OrderRepository)
}

func (m *MockRepositoryProvider) LocalUserRepository(ctx context.Context) domainmodel.LocalUserRepository {
	args := m.Called(ctx)
	return args.Get(0).(domainmodel.LocalUserRepository)
}

func (m *MockRepositoryProvider) LocalProductRepository(ctx context.Context) domainmodel.LocalProductRepository {
	args := m.Called(ctx)
	return args.Get(0).(domainmodel.LocalProductRepository)
}

type MockLockableUnitOfWork struct {
	mock.Mock
}

func (m *MockLockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx, lockNames, f)
	return args.Error(0)
}

type MockUnitOfWork struct {
	mock.Mock
	provider *MockRepositoryProvider
}

func (m *MockUnitOfWork) Execute(_ context.Context, f func(provider RepositoryProvider) error) error {
	return f(m.provider)
}

type StubLocalUserRepo struct {
	mock.Mock
}

func (m *StubLocalUserRepo) Store(_ domainmodel.LocalUser) error { return nil }
func (m *StubLocalUserRepo) Find(id uuid.UUID) (*domainmodel.LocalUser, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodel.LocalUser), args.Error(1)
}

type StubLocalProductRepo struct {
	mock.Mock
}

func (m *StubLocalProductRepo) Store(_ domainmodel.LocalProduct) error              { return nil }
func (m *StubLocalProductRepo) Find(_ uuid.UUID) (*domainmodel.LocalProduct, error) { return nil, nil }
func (m *StubLocalProductRepo) FindMany(ids []uuid.UUID) ([]domainmodel.LocalProduct, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domainmodel.LocalProduct), args.Error(1)
}

type StubOrderRepo struct {
	mock.Mock
}

func (m *StubOrderRepo) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StubOrderRepo) Store(o domainmodel.Order) error {
	args := m.Called(o)
	return args.Error(0)
}

func (m *StubOrderRepo) Find(_ uuid.UUID) (*domainmodel.Order, error) { return nil, nil }

type MockTemporalClient struct {
	mock.Mock
}

func (m *MockTemporalClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	callArgs := m.Called(ctx, options, workflow, args)
	// FIX: Проверка на nil, чтобы избежать паники при приведении типов
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(client.WorkflowRun), callArgs.Error(1)
}

func TestOrderAppService_CreateOrder(t *testing.T) {
	provider := new(MockRepositoryProvider)
	uow := &MockUnitOfWork{provider: provider}
	luow := new(MockLockableUnitOfWork)
	temporalClient := new(MockTemporalClient)

	userID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

	userRepo := new(StubLocalUserRepo)
	prodRepo := new(StubLocalProductRepo)
	orderRepo := new(StubOrderRepo)

	t.Run("success", func(t *testing.T) {
		provider.On("LocalUserRepository", mock.Anything).Return(userRepo)
		provider.On("LocalProductRepository", mock.Anything).Return(prodRepo)
		provider.On("OrderRepository", mock.Anything).Return(orderRepo)

		userRepo.On("Find", userID).Return(&domainmodel.LocalUser{UserID: userID}, nil)

		prodRepo.On("FindMany", []uuid.UUID{productID}).Return([]domainmodel.LocalProduct{
			{ProductID: productID, Price: 100},
		}, nil)

		orderRepo.On("NextID").Return(orderID, nil)
		orderRepo.On("Store", mock.AnythingOfType("model.Order")).Return(nil)

		temporalClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		dummyDispatcher := &DummyDispatcher{}
		service := NewOrderService(uow, luow, dummyDispatcher, temporalClient)

		createOrderCmd := model.CreateOrder{
			UserID: userID,
			Items:  []model.OrderItem{{ProductID: productID, Quantity: 1}},
		}

		id, err := service.CreateOrder(context.Background(), createOrderCmd)
		assert.NoError(t, err)
		assert.Equal(t, orderID, id)
	})
}

type DummyDispatcher struct{}

func (d *DummyDispatcher) Dispatch(_ context.Context, _ outbox.Event) error { return nil }
