package transport

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"orderservice/api/server/orderinternal"
	appmodel "orderservice/pkg/order/application/model"
	"orderservice/pkg/order/application/query"
	"orderservice/pkg/order/application/service"
)

func NewOrderInternalAPI(
	orderQueryService query.OrderQueryService,
	orderService service.OrderService,
) orderinternal.OrderInternalServiceServer {
	return &orderInternalAPI{
		orderQueryService: orderQueryService,
		orderService:      orderService,
	}
}

type orderInternalAPI struct {
	orderQueryService query.OrderQueryService
	orderService      service.OrderService
	orderinternal.UnimplementedOrderInternalServiceServer
}

func (a *orderInternalAPI) CreateOrder(ctx context.Context, request *orderinternal.CreateOrderRequest) (*orderinternal.CreateOrderResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user id")
	}

	items := make([]appmodel.OrderItem, len(request.Items))
	for i, item := range request.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid product id: %s", item.ProductID)
		}
		items[i] = appmodel.OrderItem{
			ProductID: productID,
			Quantity:  int(item.Quantity),
		}
	}

	orderID, err := a.orderService.CreateOrder(ctx, appmodel.CreateOrder{
		UserID: userID,
		Items:  items,
	})
	if err != nil {
		return nil, err
	}

	return &orderinternal.CreateOrderResponse{OrderID: orderID.String()}, nil
}

func (a *orderInternalAPI) FindOrder(ctx context.Context, request *orderinternal.FindOrderRequest) (*orderinternal.FindOrderResponse, error) {
	orderID, err := uuid.Parse(request.OrderID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid order id")
	}

	order, err := a.orderQueryService.FindOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return &orderinternal.FindOrderResponse{}, nil
	}

	items := make([]*orderinternal.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderinternal.OrderItem{
			ProductID: item.ProductID.String(),
			Quantity:  int32(item.Quantity), // nolint:gosec
		}
	}

	return &orderinternal.FindOrderResponse{
		Order: &orderinternal.Order{
			OrderID:    order.OrderID.String(),
			UserID:     order.UserID.String(),
			Items:      items,
			TotalPrice: order.TotalPrice,
			Status:     orderinternal.OrderStatus(order.Status), // nolint:gosec
			CreatedAt:  order.CreatedAt,
		},
	}, nil
}
