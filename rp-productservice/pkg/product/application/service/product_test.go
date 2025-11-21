package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"

	appmodel "productservice/pkg/product/application/model"
	domainmodel "productservice/pkg/product/domain/model"
)

type MockRepositoryProvider struct {
	mock.Mock
}

func (m *MockRepositoryProvider) ProductRepository(ctx context.Context) domainmodel.ProductRepository {
	return m.Called(ctx).Get(0).(domainmodel.ProductRepository)
}

type MockLockableUnitOfWork struct {
	mock.Mock
}

func (m *MockLockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error {
	args := m.Called(ctx, lockNames)
	provider := args.Get(0).(RepositoryProvider)
	return f(provider)
}

type StubProductRepo struct {
	mock.Mock
}

func (m *StubProductRepo) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *StubProductRepo) Store(p domainmodel.Product) error {
	return m.Called(p).Error(0)
}

func (m *StubProductRepo) Find(s domainmodel.FindSpec) (*domainmodel.Product, error) {
	args := m.Called(s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodel.Product), args.Error(1)
}

func (m *StubProductRepo) Delete(id uuid.UUID) error {
	return nil
}

type DummyDispatcher struct{}

func (d *DummyDispatcher) Dispatch(ctx context.Context, event outbox.Event) error {
	return nil
}

func TestProductService_StoreProduct_Create(t *testing.T) {
	provider := new(MockRepositoryProvider)
	luow := new(MockLockableUnitOfWork)
	repo := new(StubProductRepo)

	service := NewProductService(nil, luow, &DummyDispatcher{})

	ctx := context.Background()
	productID := uuid.New()
	name := "New Product"
	price := int64(500)

	cmd := appmodel.Product{
		Name:  name,
		Price: price,
	}

	luow.On("Execute", ctx, mock.Anything).Return(provider)
	provider.On("ProductRepository", ctx).Return(repo)

	repo.On("Find", domainmodel.FindSpec{Name: &name}).Return(nil, domainmodel.ErrProductNotFound)
	repo.On("NextID").Return(productID, nil)
	repo.On("Store", mock.Anything).Return(nil)

	id, err := service.StoreProduct(ctx, cmd)
	assert.NoError(t, err)
	assert.Equal(t, productID, id)
}

func TestProductService_StoreProduct_Update(t *testing.T) {
	provider := new(MockRepositoryProvider)
	luow := new(MockLockableUnitOfWork)
	repo := new(StubProductRepo)

	service := NewProductService(nil, luow, &DummyDispatcher{})

	ctx := context.Background()
	productID := uuid.New()
	name := "Updated Name"
	price := int64(600)

	cmd := appmodel.Product{
		ProductID: productID,
		Name:      name,
		Price:     price,
	}

	luow.On("Execute", ctx, mock.Anything).Return(provider)
	provider.On("ProductRepository", ctx).Return(repo)

	existing := &domainmodel.Product{ProductID: productID, Name: "Old Name", Price: 100}

	repo.On("Find", domainmodel.FindSpec{ProductID: &productID}).Return(existing, nil)
	repo.On("Find", domainmodel.FindSpec{Name: &name}).Return(nil, domainmodel.ErrProductNotFound)
	repo.On("Store", mock.MatchedBy(func(p domainmodel.Product) bool {
		return p.Name == name && p.Price == price
	})).Return(nil)

	id, err := service.StoreProduct(ctx, cmd)
	assert.NoError(t, err)
	assert.Equal(t, productID, id)
}
