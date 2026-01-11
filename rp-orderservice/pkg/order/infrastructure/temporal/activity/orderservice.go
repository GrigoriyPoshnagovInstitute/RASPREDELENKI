package activity

import (
	"orderservice/pkg/order/application/service"
)

func NewOrderServiceActivities(orderService service.OrderService) *OrderServiceActivities {
	return &OrderServiceActivities{
		orderService: orderService,
	}
}

type OrderServiceActivities struct {
	orderService service.OrderService
}
