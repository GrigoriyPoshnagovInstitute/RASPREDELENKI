package service

import (
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"

	"productservice/pkg/common/domain"
	"productservice/pkg/product/domain/model"
)

type ProductService interface {
	CreateProduct(name string, price int64, description *string) (uuid.UUID, error)
	UpdateProduct(productID uuid.UUID, name string, price int64, description *string) error
	DeleteProduct(productID uuid.UUID) error
}

func NewProductService(
	productRepository model.ProductRepository,
	eventDispatcher domain.EventDispatcher,
) ProductService {
	return &productService{
		productRepository: productRepository,
		eventDispatcher:   eventDispatcher,
	}
}

type productService struct {
	productRepository model.ProductRepository
	eventDispatcher   domain.EventDispatcher
}

func (s *productService) CreateProduct(name string, price int64, description *string) (uuid.UUID, error) {
	_, err := s.productRepository.Find(model.FindSpec{Name: &name})
	if err != nil && !errors.Is(err, model.ErrProductNotFound) {
		return uuid.Nil, err
	}
	if err == nil {
		return uuid.Nil, model.ErrProductNameAlreadyUsed
	}

	productID, err := s.productRepository.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	product := model.Product{
		ProductID:   productID,
		Name:        name,
		Description: description,
		Price:       price,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}

	err = s.productRepository.Store(product)
	if err != nil {
		return uuid.Nil, err
	}

	// кричим, что продукт создан
	return productID, s.eventDispatcher.Dispatch(&model.ProductCreated{
		ProductID:   productID,
		Name:        name,
		Description: description,
		Price:       price,
		CreatedAt:   currentTime,
	})
}

func (s *productService) UpdateProduct(productID uuid.UUID, name string, price int64, description *string) error {
	product, err := s.productRepository.Find(model.FindSpec{ProductID: &productID})
	if err != nil {
		return err
	}

	if product.Name != name {
		existing, err := s.productRepository.Find(model.FindSpec{Name: &name})
		if err != nil && !errors.Is(err, model.ErrProductNotFound) {
			return err
		}
		if existing != nil && existing.ProductID != productID {
			return model.ErrProductNameAlreadyUsed
		}
	}

	if product.Name == name && product.Price == price && reflect.DeepEqual(product.Description, description) {
		return nil
	}

	currentTime := time.Now()
	product.Name = name
	product.Price = price
	product.Description = description
	product.UpdatedAt = currentTime

	err = s.productRepository.Store(*product)
	if err != nil {
		return err
	}

	// кричим, что продукт обновлен
	return s.eventDispatcher.Dispatch(&model.ProductUpdated{
		ProductID: productID,
		UpdatedFields: struct {
			Name        *string
			Description *string
			Price       *int64
		}{Name: &name, Description: description, Price: &price},
		UpdatedAt: currentTime,
	})
}

func (s *productService) DeleteProduct(productID uuid.UUID) error {
	_, err := s.productRepository.Find(model.FindSpec{ProductID: &productID})
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			return nil
		}
		return err
	}

	err = s.productRepository.Delete(productID)
	if err != nil {
		return err
	}

	return s.eventDispatcher.Dispatch(&model.ProductDeleted{
		ProductID: productID,
		DeletedAt: time.Now(),
	})
}
