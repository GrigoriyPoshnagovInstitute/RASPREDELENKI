package service

import (
	"time"

	"orderservice/pkg/common/domain"
	"orderservice/pkg/order/domain/model"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(userID uuid.UUID, items []model.OrderItem) (uuid.UUID, error)
	MarkAsPaid(orderID uuid.UUID) error
	CancelOrder(orderID uuid.UUID, reason string) error
}

func NewOrderService(
	orderRepo model.OrderRepository,
	eventDispatcher domain.EventDispatcher,
) OrderService {
	return &orderService{
		orderRepository: orderRepo,
		eventDispatcher: eventDispatcher,
	}
}

type orderService struct {
	orderRepository model.OrderRepository
	eventDispatcher domain.EventDispatcher
}

func (s *orderService) CreateOrder(userID uuid.UUID, items []model.OrderItem) (uuid.UUID, error) {
	if len(items) == 0 {
		return uuid.Nil, model.ErrEmptyOrder
	}

	orderID, err := s.orderRepository.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	var totalPrice int64
	for _, item := range items {
		totalPrice += item.Price * int64(item.Quantity)
	}

	currentTime := time.Now()
	order := model.Order{
		OrderID:    orderID,
		UserID:     userID,
		Items:      items,
		TotalPrice: totalPrice,
		Status:     model.StatusCreated,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	}

	if err := s.orderRepository.Store(order); err != nil {
		return uuid.Nil, err
	}

	// орем, что заказ создан

	return orderID, s.eventDispatcher.Dispatch(&model.OrderCreated{
		OrderID:    orderID,
		UserID:     userID,
		TotalPrice: totalPrice,
		Items:      items,
		CreatedAt:  currentTime,
	})
}

func (s *orderService) MarkAsPaid(orderID uuid.UUID) error {
	order, err := s.orderRepository.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status == model.StatusPaid {
		return nil
	}

	order.Status = model.StatusPaid
	order.UpdatedAt = time.Now()

	if err := s.orderRepository.Store(*order); err != nil {
		return err
	}

	return s.eventDispatcher.Dispatch(&model.OrderPaid{
		OrderID: orderID,
		PaidAt:  order.UpdatedAt,
	})
}

func (s *orderService) CancelOrder(orderID uuid.UUID, reason string) error {
	order, err := s.orderRepository.Find(orderID)
	if err != nil {
		return err
	}

	if order.Status == model.StatusCancelled || order.Status == model.StatusPaid {
		return nil
	}

	order.Status = model.StatusCancelled
	order.UpdatedAt = time.Now()

	if err := s.orderRepository.Store(*order); err != nil {
		return err
	}

	// орем, что заказ отменен

	return s.eventDispatcher.Dispatch(&model.OrderCancelled{
		OrderID:     orderID,
		Reason:      reason,
		CancelledAt: order.UpdatedAt,
	})
}
