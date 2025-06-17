package internal

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type InboxRepository interface {
	CreateInboxMessage(ctx context.Context, orderID, payload string) error
	GetUnprocessedMessages(ctx context.Context) ([]*InboxMessage, error)
	MarkMessageAsProcessed(ctx context.Context, id string) error
}

type inboxRepository struct {
	db *sql.DB
}

func NewInboxRepository(db *sql.DB) InboxRepository {
	return &inboxRepository{db: db}
}

func (r *inboxRepository) CreateInboxMessage(ctx context.Context, orderID, payload string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO inbox_messages (id, order_id, payload, processed) VALUES ($1, $2, $3, $4)",
		uuid.New().String(), orderID, payload, false)
	return err
}

func (r *inboxRepository) GetUnprocessedMessages(ctx context.Context) ([]*InboxMessage, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, payload, processed FROM inbox_messages WHERE processed = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*InboxMessage
	for rows.Next() {
		var msg InboxMessage
		if err := rows.Scan(&msg.ID, &msg.OrderID, &msg.Payload, &msg.Processed); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (r *inboxRepository) MarkMessageAsProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE inbox_messages SET processed = true WHERE id = $1", id)
	return err
}
