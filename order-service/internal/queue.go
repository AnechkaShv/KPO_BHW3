package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		url:     url,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

type RabbitMQPaymentQueue struct {
	rabbitMQ     *RabbitMQ
	exchangeName string
	routingKey   string
	queueName    string
}

func NewRabbitMQPaymentQueue(rabbitMQ *RabbitMQ, exchangeName, routingKey, queueName string) *RabbitMQPaymentQueue {
	return &RabbitMQPaymentQueue{
		rabbitMQ:     rabbitMQ,
		exchangeName: exchangeName,
		routingKey:   routingKey,
		queueName:    queueName,
	}
}

func (q *RabbitMQPaymentQueue) PublishPaymentRequest(ctx context.Context, message []byte) error {
	err := q.rabbitMQ.channel.PublishWithContext(
		ctx,
		q.exchangeName,
		q.routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish payment request: %w", err)
	}
	return nil
}

func (q *RabbitMQPaymentQueue) SubscribeToPaymentUpdates(ctx context.Context, callback func(orderID string, success bool)) error {
	_, err := q.rabbitMQ.channel.QueueDeclare(
		q.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := q.rabbitMQ.channel.Consume(
		q.queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping payment updates subscriber")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				var result struct {
					OrderID string `json:"order_id"`
					Success bool   `json:"success"`
				}
				if err := json.Unmarshal(msg.Body, &result); err != nil {
					log.Printf("Failed to unmarshal payment result: %v", err)
					continue
				}

				callback(result.OrderID, result.Success)
			}
		}
	}()

	return nil
}
