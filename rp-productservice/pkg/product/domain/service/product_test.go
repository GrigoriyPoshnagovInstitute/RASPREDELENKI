package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"productservice/pkg/common/domain"
	"productservice/pkg/product/domain/model"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockProductRepository) Store(product model.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Find(spec model.FindSpec) (*model.Product, error) {
	args := m.Called(spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(productID uuid.UUID) error {
	args := m.Called(productID)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestProductService_CreateProduct(t *testing.T) {
	repo := new(MockProductRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewProductService(repo, dispatcher)

	name := "Test Product"
	price := int64(1000)
	productID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo.On("Find", model.FindSpec{Name: &name}).Return(nil, model.ErrProductNotFound).Once()
		repo.On("NextID").Return(productID, nil).Once()
		repo.On("Store", mock.MatchedBy(func(p model.Product) bool {
			return p.Name == name && p.Price == price && p.ProductID == productID
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.ProductCreated) bool {
			return e.ProductID == productID
		})).Return(nil).Once()

		id, err := service.CreateProduct(name, price, nil)
		assert.NoError(t, err)
		assert.Equal(t, productID, id)
		repo.AssertExpectations(t)
	})

	t.Run("name_conflict", func(t *testing.T) {
		repo.On("Find", model.FindSpec{Name: &name}).Return(&model.Product{}, nil).Once()

		_, err := service.CreateProduct(name, price, nil)
		assert.ErrorIs(t, err, model.ErrProductNameAlreadyUsed)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	repo := new(MockProductRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewProductService(repo, dispatcher)

	productID := uuid.New()
	oldName := "Old Name"
	newName := "New Name"

	t.Run("success", func(t *testing.T) {
		existing := &model.Product{ProductID: productID, Name: oldName, Price: 100}
		repo.On("Find", model.FindSpec{ProductID: &productID}).Return(existing, nil).Once()
		repo.On("Find", model.FindSpec{Name: &newName}).Return(nil, model.ErrProductNotFound).Once()
		repo.On("Store", mock.MatchedBy(func(p model.Product) bool {
			return p.Name == newName && p.Price == 200
		})).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.ProductUpdated) bool {
			return e.ProductID == productID && *e.UpdatedFields.Name == newName
		})).Return(nil).Once()

		err := service.UpdateProduct(productID, newName, 200, nil)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		existing := &model.Product{ProductID: productID, Name: oldName}
		otherProduct := &model.Product{ProductID: uuid.New(), Name: newName}

		repo.On("Find", model.FindSpec{ProductID: &productID}).Return(existing, nil).Once()
		repo.On("Find", model.FindSpec{Name: &newName}).Return(otherProduct, nil).Once()

		err := service.UpdateProduct(productID, newName, 200, nil)
		assert.ErrorIs(t, err, model.ErrProductNameAlreadyUsed)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	repo := new(MockProductRepository)
	dispatcher := new(MockEventDispatcher)
	service := NewProductService(repo, dispatcher)

	productID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo.On("Find", model.FindSpec{ProductID: &productID}).Return(&model.Product{}, nil).Once()
		repo.On("Delete", productID).Return(nil).Once()
		dispatcher.On("Dispatch", mock.MatchedBy(func(e *model.ProductDeleted) bool {
			return e.ProductID == productID
		})).Return(nil).Once()

		err := service.DeleteProduct(productID)
		assert.NoError(t, err)
	})
}
