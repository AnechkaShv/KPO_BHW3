package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, userID string) (*Account, error)
	GetAccountByUserID(ctx context.Context, userID string) (*Account, error)
	Deposit(ctx context.Context, userID string, amount float64) error
	Withdraw(ctx context.Context, userID string, amount float64) error
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) CreateAccount(ctx context.Context, userID string) (*Account, error) {
	accountID := uuid.New().String()
	var account Account
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO accounts (id, user_id, balance) VALUES ($1, $2, $3) RETURNING id, user_id, balance",
		accountID, userID, 0).Scan(&account.ID, &account.UserID, &account.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return &account, nil
}

func (r *accountRepository) GetAccountByUserID(ctx context.Context, userID string) (*Account, error) {
	var account Account
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, balance FROM accounts WHERE user_id = $1", userID).
		Scan(&account.ID, &account.UserID, &account.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &account, nil
}

func (r *accountRepository) Deposit(ctx context.Context, userID string, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE user_id = $2", amount, userID)
	return err
}

func (r *accountRepository) Withdraw(ctx context.Context, userID string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRowContext(ctx,
		"SELECT balance FROM accounts WHERE user_id = $1 FOR UPDATE", userID).
		Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if currentBalance < amount {
		return errors.New("insufficient funds")
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return tx.Commit()
}
