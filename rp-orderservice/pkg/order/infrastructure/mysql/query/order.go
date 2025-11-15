package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "orderservice/pkg/order/application/model"
	"orderservice/pkg/order/application/query"
	"orderservice/pkg/order/domain/model"
)

func NewOrderQueryService(client mysql.ClientContext) query.OrderQueryService {
	return &orderQueryService{
		client: client,
	}
}

type orderQueryService struct {
	client mysql.ClientContext
}

func (s *orderQueryService) FindOrder(ctx context.Context, orderID uuid.UUID) (*appmodel.Order, error) {
	orderData := struct {
		OrderID    uuid.UUID `db:"order_id"`
		UserID     uuid.UUID `db:"user_id"`
		TotalPrice int64     `db:"total_price"`
		Status     int       `db:"status"`
		CreatedAt  time.Time `db:"created_at"`
	}{}

	err := s.client.GetContext(ctx, &orderData, "SELECT order_id, user_id, total_price, status, created_at FROM `order` WHERE order_id = ?", orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrOrderNotFound)
		}
		return nil, errors.WithStack(err)
	}

	var itemsData []struct {
		ProductID uuid.UUID `db:"product_id"`
		Quantity  int       `db:"quantity"`
	}
	err = s.client.SelectContext(ctx, &itemsData, `SELECT product_id, quantity FROM order_item WHERE order_id = ?`, orderID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]appmodel.OrderItem, len(itemsData))
	for i, itemData := range itemsData {
		items[i] = appmodel.OrderItem{
			ProductID: itemData.ProductID,
			Quantity:  itemData.Quantity,
		}
	}

	return &appmodel.Order{
		OrderID:    orderData.OrderID,
		UserID:     orderData.UserID,
		Items:      items,
		TotalPrice: orderData.TotalPrice,
		Status:     orderData.Status,
		CreatedAt:  orderData.CreatedAt.Unix(),
	}, nil
}
