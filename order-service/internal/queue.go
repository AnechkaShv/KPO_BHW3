package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	var conn *amqp.Connection
	var err error

	retryDelay := 5 * time.Second
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		log.Printf("Attempt %d/%d: Failed to connect to RabbitMQ at %s: %v",
			i+1, maxRetries, url, err)

		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}

		time.Sleep(retryDelay)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		"payments",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
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

func NewRabbitMQPaymentQueue(
	rabbitMQ *RabbitMQ,
	exchangeName,
	routingKey,
	queueName string,
) *RabbitMQPaymentQueue {
	return &RabbitMQPaymentQueue{
		rabbitMQ:     rabbitMQ,
		exchangeName: exchangeName,
		routingKey:   routingKey,
		queueName:    queueName,
	}
}

func (q *RabbitMQPaymentQueue) PublishPaymentRequest(ctx context.Context, message []byte) error {
	log.Printf("Publishing message to exchange '%s' with routing key '%s'",
		q.exchangeName, q.routingKey)
	log.Printf("Message content: %s", string(message))

	err := q.rabbitMQ.channel.Publish(
		q.exchangeName,
		q.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})

	if err != nil {
		log.Printf("Publish failed with error: %v", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Println("Message successfully published")
	return nil
}

func (q *RabbitMQPaymentQueue) SubscribeToPaymentUpdates(ctx context.Context, callback func(orderID string, success bool)) error {
	_, err := q.rabbitMQ.channel.QueueDeclare(
		q.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = q.rabbitMQ.channel.QueueBind(
		q.queueName,
		q.routingKey,
		q.exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := q.rabbitMQ.channel.Consume(
		q.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var result struct {
					OrderID string `json:"order_id"`
					Success bool   `json:"success"`
				}

				if err := json.Unmarshal(msg.Body, &result); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}

				callback(result.OrderID, result.Success)
			}
		}
	}()

	return nil
}
