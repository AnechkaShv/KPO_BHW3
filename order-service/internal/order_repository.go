package internal

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id string) (*Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *Order) error {
	order.ID = uuid.New().String()
	order.Status = OrderStatusNew
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO orders (id, user_id, amount, description, status) VALUES ($1, $2, $3, $4, $5)",
		order.ID, order.UserID, order.Amount, order.Description, order.Status)
	return err
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	var order Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, amount, description, status FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.UserID, &order.Amount, &order.Description, &order.Status)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]*Order, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, amount, description, status FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Amount, &order.Description, &order.Status); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	return err
}
