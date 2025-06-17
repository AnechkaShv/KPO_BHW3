package internal

import (
	"context"
	"encoding/json"
	"fmt"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, amount float64, description string) (*Order, error)
	GetOrder(ctx context.Context, id string) (*Order, error)
	ListOrders(ctx context.Context, userID string) ([]*Order, error)
	ProcessPaymentEvent(ctx context.Context, orderID string, success bool) error
}

type orderService struct {
	orderRepo    OrderRepository
	outboxRepo   OutboxRepository
	paymentQueue *RabbitMQPaymentQueue
}

func NewOrderService(
	orderRepo OrderRepository,
	outboxRepo OutboxRepository,
	paymentQueue *RabbitMQPaymentQueue,
) OrderService {
	return &orderService{
		orderRepo:    orderRepo,
		outboxRepo:   outboxRepo,
		paymentQueue: paymentQueue,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, amount float64, description string) (*Order, error) {
	order := &Order{
		UserID:      userID,
		Amount:      amount,
		Description: description,
		Status:      OrderStatusNew,
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	paymentTask := map[string]interface{}{
		"order_id":    order.ID,
		"user_id":     userID,
		"amount":      amount,
		"description": description,
	}

	payload, err := json.Marshal(paymentTask)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment task: %w", err)
	}

	if err := s.outboxRepo.CreateOutboxMessage(ctx, order.ID, string(payload)); err != nil {
		return nil, fmt.Errorf("failed to create outbox message: %w", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*Order, error) {
	return s.orderRepo.GetOrderByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, userID string) ([]*Order, error) {
	return s.orderRepo.GetOrdersByUserID(ctx, userID)
}

func (s *orderService) ProcessPaymentEvent(ctx context.Context, orderID string, success bool) error {
	var status OrderStatus
	if success {
		status = OrderStatusPaid
	} else {
		status = OrderStatusCancelled
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}
