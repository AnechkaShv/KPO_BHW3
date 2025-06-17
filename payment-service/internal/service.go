package internal

import (
	"context"
	"database/sql"
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
}

func NewPaymentService(
	db *sql.DB,
	accountRepo AccountRepository,
	inboxRepo InboxRepository,
) PaymentService {
	return &paymentService{
		db:          db,
		accountRepo: accountRepo,
		inboxRepo:   inboxRepo,
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

	// 1. Проверяем существование аккаунта
	var accountID string
	var balance float64
	err = tx.QueryRowContext(ctx,
		"SELECT id, balance FROM accounts WHERE user_id = $1 FOR UPDATE",
		userID).Scan(&accountID, &balance)

	if err != nil {
		if err == sql.ErrNoRows {
			return &PaymentResult{
				Success: false,
				Message: "account not found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	log.Printf("Current balance: %.2f, Payment amount: %.2f", balance, amount)

	// 2. Проверяем достаточность средств
	if balance < amount {
		return &PaymentResult{
			Success: false,
			Message: "insufficient funds",
		}, nil
	}

	// 3. Списание средств
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// 4. Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Payment successful. New balance: %.2f", balance-amount)

	return &PaymentResult{
		Success: true,
		Message: "payment processed successfully",
	}, nil
}
