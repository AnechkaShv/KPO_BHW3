package internal

import (
	"context"
	"errors"
	"fmt"
)

type PaymentService interface {
	CreateAccount(ctx context.Context, userID string) (*Account, error)
	GetAccount(ctx context.Context, userID string) (*Account, error)
	Deposit(ctx context.Context, userID string, amount float64) error
	ProcessPaymentTask(ctx context.Context, orderID, userID string, amount float64) (bool, error)
	CreateOutboxMessage(ctx context.Context, orderID, payload string) error
}

type paymentService struct {
	accountRepo AccountRepository
	inboxRepo   InboxRepository
	outboxRepo  OutboxRepository
}

func NewPaymentService(
	accountRepo AccountRepository,
	inboxRepo InboxRepository,
	outboxRepo OutboxRepository,
) PaymentService {
	return &paymentService{
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

func (s *paymentService) CreateOutboxMessage(ctx context.Context, orderID, payload string) error {
	return s.outboxRepo.CreateOutboxMessage(ctx, orderID, payload)
}
