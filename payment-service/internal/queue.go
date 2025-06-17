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

	// Объявляем exchange и очередь
	err = ch.ExchangeDeclare(
		"payments",
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
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
	err := q.rabbitMQ.channel.Publish(
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
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (q *RabbitMQPaymentQueue) SubscribeToPaymentUpdates(
	ctx context.Context,
	callback func(orderID, userID string, amount float64),
) error {
	// Объявление и привязка очереди
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

	// Потребление сообщений
	msgs, err := q.rabbitMQ.channel.Consume(
		q.queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
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

				var request struct {
					OrderID string  `json:"order_id"`
					UserID  string  `json:"user_id"`
					Amount  float64 `json:"amount"`
				}

				if err := json.Unmarshal(msg.Body, &request); err != nil {
					log.Printf("Failed to unmarshal payment request: %v", err)
					continue
				}

				callback(request.OrderID, request.UserID, request.Amount)
			}
		}
	}()

	return nil
}
