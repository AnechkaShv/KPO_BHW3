package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type PaymentService interface {
	CreateAccount(ctx context.Context, userID string) (*Account, error)
	GetAccount(ctx context.Context, userID string) (*Account, error)
	Deposit(ctx context.Context, userID string, amount float64) error
	ProcessOrderPayment(ctx context.Context, orderID, userID string, amount float64) (*PaymentResult, error)
}

type paymentService struct {
	db          *sql.DB
	accountRepo AccountRepository
	inboxRepo   InboxRepository
	queue       *RabbitMQPaymentQueue
}

func NewPaymentService(
	db *sql.DB,
	accountRepo AccountRepository,
	inboxRepo InboxRepository,
	queue *RabbitMQPaymentQueue,
) PaymentService {
	return &paymentService{
		db:          db,
		accountRepo: accountRepo,
		inboxRepo:   inboxRepo,
		queue:       queue,
	}
}

func (s *paymentService) CreateAccount(ctx context.Context, userID string) (*Account, error) {
	account, err := s.accountRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check account existence: %w", err)
	}
	if account != nil {
		return nil, errors.New("account already exists")
	}

	return s.accountRepo.CreateAccount(ctx, userID)
}

func (s *paymentService) GetAccount(ctx context.Context, userID string) (*Account, error) {
	account, err := s.accountRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	return account, nil
}

func (s *paymentService) Deposit(ctx context.Context, userID string, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	account, err := s.accountRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check account existence: %w", err)
	}
	if account == nil {
		return errors.New("account not found")
	}

	return s.accountRepo.Deposit(ctx, userID, amount)
}

func (s *paymentService) ProcessOrderPayment(ctx context.Context, orderID, userID string, amount float64) (*PaymentResult, error) {
	log.Printf("Processing payment: OrderID=%s, UserID=%s, Amount=%.2f", orderID, userID, amount)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var accountID string
	var balance float64
	err = tx.QueryRowContext(ctx,
		"SELECT id, balance FROM accounts WHERE user_id = $1 FOR UPDATE",
		userID).Scan(&accountID, &balance)

	if err != nil {
		if err == sql.ErrNoRows {
			result := &PaymentResult{
				OrderID: orderID,
				Success: false,
				Message: "account not found",
			}
			if err := s.sendPaymentResponse(ctx, result); err != nil {
				log.Printf("Failed to send payment response: %v", err)
			}
			return result, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	log.Printf("Current balance: %.2f, Payment amount: %.2f", balance, amount)

	if balance < amount {
		result := &PaymentResult{
			OrderID: orderID,
			Success: false,
			Message: "insufficient funds",
		}
		if err := s.sendPaymentResponse(ctx, result); err != nil {
			log.Printf("Failed to send payment response: %v", err)
		}
		return result, nil
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Payment successful. New balance: %.2f", balance-amount)

	result := &PaymentResult{
		OrderID: orderID,
		Success: true,
		Message: "payment processed successfully",
	}
	if err := s.sendPaymentResponse(ctx, result); err != nil {
		log.Printf("Failed to send payment response: %v", err)
	}

	return result, nil
}

func (s *paymentService) sendPaymentResponse(ctx context.Context, result *PaymentResult) error {
	response := map[string]interface{}{
		"order_id": result.OrderID,
		"success":  result.Success,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	return s.queue.PublishPaymentRequest(ctx, responseBytes)
}
