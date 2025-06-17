package internal

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type OutboxRepository interface {
	CreateOutboxMessage(ctx context.Context, orderID, payload string) error
	GetUnprocessedMessages(ctx context.Context) ([]*OutboxMessage, error)
	MarkMessageAsProcessed(ctx context.Context, id string) error
}

type outboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) CreateOutboxMessage(ctx context.Context, orderID, payload string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO outbox_messages (id, order_id, payload, processed) VALUES ($1, $2, $3, $4)",
		uuid.New().String(), orderID, payload, false)
	return err
}

func (r *outboxRepository) GetUnprocessedMessages(ctx context.Context) ([]*OutboxMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, payload, processed FROM outbox_messages WHERE processed = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*OutboxMessage
	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.OrderID, &msg.Payload, &msg.Processed); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *outboxRepository) MarkMessageAsProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE outbox_messages SET processed = true WHERE id = $1", id)
	return err
}
