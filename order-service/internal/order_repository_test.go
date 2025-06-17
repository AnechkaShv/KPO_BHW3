package internal

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepository_CreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrderRepository(db)

	t.Run("success", func(t *testing.T) {
		testOrder := &Order{
			UserID:      "user1",
			Amount:      100.50,
			Description: "test order",
		}

		mock.ExpectExec("INSERT INTO orders").
			WithArgs(sqlmock.AnyArg(), testOrder.UserID, testOrder.Amount, testOrder.Description, OrderStatusNew).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateOrder(context.Background(), testOrder)
		assert.NoError(t, err)
		assert.NotEmpty(t, testOrder.ID)
		assert.Equal(t, OrderStatusNew, testOrder.Status)
	})

	t.Run("database error", func(t *testing.T) {
		testOrder := &Order{UserID: "user1"}
		mock.ExpectExec("INSERT INTO orders").WillReturnError(sql.ErrConnDone)

		err := repo.CreateOrder(context.Background(), testOrder)
		assert.Error(t, err)
	})
}

func TestOrderRepository_GetOrderByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrderRepository(db)
	testID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "amount", "description", "status"}).
			AddRow(testID, "user1", 100.50, "test order", "NEW")

		mock.ExpectQuery("SELECT id, user_id, amount, description, status FROM orders WHERE id = ?").
			WithArgs(testID).
			WillReturnRows(rows)

		order, err := repo.GetOrderByID(context.Background(), testID)
		assert.NoError(t, err)
		assert.Equal(t, testID, order.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT .*").WithArgs(testID).WillReturnError(sql.ErrNoRows)

		order, err := repo.GetOrderByID(context.Background(), testID)
		assert.Nil(t, order)
		assert.NoError(t, err)
	})
}
