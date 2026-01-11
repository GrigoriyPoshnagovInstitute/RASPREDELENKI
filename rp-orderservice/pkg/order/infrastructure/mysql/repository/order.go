package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"orderservice/pkg/order/domain/model"
	"orderservice/pkg/order/infrastructure/metrics"
)

func NewOrderRepository(ctx context.Context, client mysql.ClientContext) model.OrderRepository {
	return &orderRepository{
		ctx:    ctx,
		client: client,
	}
}

type orderRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *orderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (r *orderRepository) Store(order model.Order) (err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.DatabaseDuration.WithLabelValues("store", "order", status).Observe(time.Since(start).Seconds())
	}()

	_, err = r.client.ExecContext(r.ctx,
		"INSERT INTO `order` (order_id, user_id, total_price, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE total_price=VALUES(total_price), status=VALUES(status), updated_at=VALUES(updated_at)",
		order.OrderID, order.UserID, order.TotalPrice, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = r.client.ExecContext(r.ctx, `DELETE FROM order_item WHERE order_id = ?`, order.OrderID)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, item := range order.Items {
		_, err = r.client.ExecContext(r.ctx,
			`INSERT INTO order_item (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)`,
			order.OrderID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *orderRepository) Find(orderID uuid.UUID) (_ *model.Order, err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil && !errors.Is(err, model.ErrOrderNotFound) {
			status = "error"
		}
		metrics.DatabaseDuration.WithLabelValues("find", "order", status).Observe(time.Since(start).Seconds())
	}()

	orderData := struct {
		OrderID    uuid.UUID `db:"order_id"`
		UserID     uuid.UUID `db:"user_id"`
		TotalPrice int64     `db:"total_price"`
		Status     int       `db:"status"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}{}

	err = r.client.GetContext(r.ctx, &orderData, "SELECT order_id, user_id, total_price, status, created_at, updated_at FROM `order` WHERE order_id = ?", orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrOrderNotFound)
		}
		return nil, errors.WithStack(err)
	}

	var itemsData []struct {
		ProductID uuid.UUID `db:"product_id"`
		Quantity  int       `db:"quantity"`
		Price     int64     `db:"price"`
	}
	err = r.client.SelectContext(r.ctx, &itemsData, `SELECT product_id, quantity, price FROM order_item WHERE order_id = ?`, orderID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	items := make([]model.OrderItem, len(itemsData))
	for i, itemData := range itemsData {
		items[i] = model.OrderItem{
			ProductID: itemData.ProductID,
			Quantity:  itemData.Quantity,
			Price:     itemData.Price,
		}
	}

	return &model.Order{
		OrderID:    orderData.OrderID,
		UserID:     orderData.UserID,
		Items:      items,
		TotalPrice: orderData.TotalPrice,
		Status:     model.OrderStatus(orderData.Status),
		CreatedAt:  orderData.CreatedAt,
		UpdatedAt:  orderData.UpdatedAt,
	}, nil
}
