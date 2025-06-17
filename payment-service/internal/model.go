package internal

type Account struct {
	ID      string  `json:"id" db:"id"`
	UserID  string  `json:"user_id" db:"user_id"`
	Balance float64 `json:"balance" db:"balance"`
}

type PaymentResult struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount,omitempty"`
}

type PaymentEvent struct {
	OrderID string  `json:"order_id" db:"order_id"`
	UserID  string  `json:"user_id" db:"user_id"`
	Amount  float64 `json:"amount" db:"amount"`
	Success bool    `json:"success" db:"success"`
}

type OutboxMessage struct {
	ID        string `json:"id" db:"id"`
	OrderID   string `json:"order_id" db:"order_id"`
	Payload   string `json:"payload" db:"payload"`
	Processed bool   `json:"processed" db:"processed"`
}

type InboxMessage struct {
	ID        string `json:"id" db:"id"`
	OrderID   string `json:"order_id" db:"order_id"`
	Payload   string `json:"payload" db:"payload"`
	Processed bool   `json:"processed" db:"processed"`
}
