package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PaymentService interface {
	CreateAccount(ctx context.Context, userID string) (*Account, error)
	GetAccount(ctx context.Context, userID string) (*Account, error)
	Deposit(ctx context.Context, userID string, amount float64) error
	ProcessOrderPayment(ctx context.Context, orderID, userID string, amount float64) (*PaymentResult, error)
	CreateOutboxMessage(ctx context.Context, orderID, payload string) error
}

type paymentService struct {
	db          *sql.DB // Добавляем прямое подключение к БД
	accountRepo AccountRepository
	inboxRepo   InboxRepository
	outboxRepo  OutboxRepository
}

func NewPaymentService(
	db *sql.DB,
	accountRepo AccountRepository,
	inboxRepo InboxRepository,
	outboxRepo OutboxRepository,
) PaymentService {
	return &paymentService{
		db:          db,
		accountRepo: accountRepo,
		inboxRepo:   inboxRepo,
		outboxRepo:  outboxRepo,
	}
}

func (s *paymentService) CreateAccount(ctx context.Context, userID string) (*Account, error) {
	// Check if account already exists
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

	// Check if account exists
	account, err := s.accountRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check account existence: %w", err)
	}
	if account == nil {
		return errors.New("account not found")
	}

	return s.accountRepo.Deposit(ctx, userID, amount)
}

func (s *paymentService) ProcessPaymentTask(ctx context.Context, orderID, userID string, amount float64) (bool, error) {
	// Check if account exists
	account, err := s.accountRepo.GetAccountByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}
	if account == nil {
		// Account doesn't exist - payment failed
		return false, nil
	}

	// Try to withdraw funds
	err = s.accountRepo.Withdraw(ctx, userID, amount)
	if err != nil {
		if err.Error() == "insufficient funds" {
			return false, nil
		}
		return false, fmt.Errorf("failed to withdraw funds: %w", err)
	}

	// Payment succeeded
	return true, nil
}

func (s *paymentService) ProcessOrderPayment(ctx context.Context, orderID, userID string, amount float64) (*PaymentResult, error) {
	// Валидация суммы
	if amount <= 0 {
		return &PaymentResult{
			Success: false,
			Message: "amount must be positive",
			OrderID: orderID,
		}, nil
	}

	// Логика обработки платежа
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	defer tx.Rollback()

	// Проверка счета и баланса
	var balance float64
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM accounts WHERE user_id = $1 FOR UPDATE",
		userID).Scan(&balance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &PaymentResult{
				Success: false,
				Message: "account not found",
				OrderID: orderID,
			}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance < amount {
		return &PaymentResult{
			Success: false,
			Message: "insufficient funds",
			OrderID: orderID,
		}, nil
	}

	// Списание средств
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE user_id = $2",
		amount, userID)
	if err != nil {
		return nil, fmt.Errorf("withdrawal failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &PaymentResult{
		Success: true,
		Message: "payment processed successfully",
		OrderID: orderID,
		Amount:  amount,
	}, nil
}

func (s *paymentService) CreateOutboxMessage(ctx context.Context, orderID, payload string) error {
	return s.outboxRepo.CreateOutboxMessage(ctx, orderID, payload)
}
