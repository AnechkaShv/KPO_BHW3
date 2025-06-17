package internal

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          string      `json:"id" db:"id"`
	UserID      string      `json:"user_id" db:"user_id"`
	Amount      float64     `json:"amount" db:"amount"`
	Description string      `json:"description" db:"description"`
	Status      OrderStatus `json:"status" db:"status"`
}

type OutboxMessage struct {
	ID        string `json:"id" db:"id"`
	OrderID   string `json:"order_id" db:"order_id"`
	Payload   string `json:"payload" db:"payload"`
	Processed bool   `json:"processed" db:"processed"`
}
