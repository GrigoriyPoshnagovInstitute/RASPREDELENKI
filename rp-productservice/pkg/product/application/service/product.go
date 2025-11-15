package service

import (
	"context"
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"productservice/pkg/common/domain"
	appmodel "productservice/pkg/product/application/model"
	"productservice/pkg/product/domain/model"
	"productservice/pkg/product/domain/service"
)

type ProductService interface {
	StoreProduct(ctx context.Context, product appmodel.Product) (uuid.UUID, error)
}

func NewProductService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) ProductService {
	return &productService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type productService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *productService) StoreProduct(ctx context.Context, product appmodel.Product) (uuid.UUID, error) {
	var lockNames []string
	if product.ProductID != uuid.Nil {
		lockNames = append(lockNames, productLock(product.ProductID))
	} else {
		lockNames = append(lockNames, productNameLock(product.Name))
	}

	productID := product.ProductID
	err := s.luow.Execute(ctx, lockNames, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.ProductRepository(ctx))
		if product.ProductID == uuid.Nil {
			pID, err := domainService.CreateProduct(product.Name, product.Price, product.Description)
			if err != nil {
				return err
			}
			productID = pID
		} else {
			err := domainService.UpdateProduct(productID, product.Name, product.Price, product.Description)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return productID, err
}

func (s *productService) domainService(ctx context.Context, repository model.ProductRepository) service.ProductService {
	return service.NewProductService(repository, s.domainEventDispatcher(ctx))
}

func (s *productService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseProductLock = "product_"

func productLock(id uuid.UUID) string {
	return baseProductLock + id.String()
}

func productNameLock(name string) string {
	return fmt.Sprintf("%sname_%s", baseProductLock, name)
}
