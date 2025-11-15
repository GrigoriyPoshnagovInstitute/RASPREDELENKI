package service

import (
	"context"
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"orderservice/pkg/common/domain"
	appmodel "orderservice/pkg/order/application/model"
	"orderservice/pkg/order/domain/model"
	"orderservice/pkg/order/domain/service"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order appmodel.CreateOrder) (uuid.UUID, error)
	HandlePaymentResult(ctx context.Context, orderID uuid.UUID, success bool) error
}

func NewOrderService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) OrderService {
	return &orderService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type orderService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *orderService) CreateOrder(ctx context.Context, order appmodel.CreateOrder) (uuid.UUID, error) {
	var orderID uuid.UUID

	err := s.uow.Execute(ctx, func(provider RepositoryProvider) error {
		userRepo := provider.LocalUserRepository(ctx)
		productRepo := provider.LocalProductRepository(ctx)

		if _, err := userRepo.Find(order.UserID); err != nil {
			return errors.Wrap(model.ErrUserNotFound, err.Error())
		}

		productIDs := make([]uuid.UUID, len(order.Items))
		for i, item := range order.Items {
			productIDs[i] = item.ProductID
		}

		products, err := productRepo.FindMany(productIDs)
		if err != nil {
			return err
		}
		if len(products) != len(order.Items) {
			return model.ErrProductNotFound
		}

		productMap := make(map[uuid.UUID]model.LocalProduct, len(products))
		for _, p := range products {
			productMap[p.ProductID] = p
		}

		domainItems := make([]model.OrderItem, len(order.Items))
		for i, item := range order.Items {
			product := productMap[item.ProductID]
			domainItems[i] = model.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			}
		}

		domainService := s.domainService(ctx, provider)
		id, err := domainService.CreateOrder(order.UserID, domainItems)
		if err != nil {
			return err
		}
		orderID = id
		return nil
	})

	return orderID, err
}

func (s *orderService) HandlePaymentResult(ctx context.Context, orderID uuid.UUID, success bool) error {
	lockName := orderLock(orderID)
	return s.luow.Execute(ctx, []string{lockName}, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider)
		if success {
			return domainService.MarkAsPaid(orderID)
		}
		return domainService.CancelOrder(orderID, "Payment failed")
	})
}

func (s *orderService) domainService(ctx context.Context, provider RepositoryProvider) service.OrderService {
	return service.NewOrderService(provider.OrderRepository(ctx), s.domainEventDispatcher(ctx))
}

func (s *orderService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseOrderLock = "order_"

func orderLock(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", baseOrderLock, id.String())
}
