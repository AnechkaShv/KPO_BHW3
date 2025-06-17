package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnechkaShv/KPO_BHW2/payment-service/internal"
	_ "github.com/lib/pq"
)

func main() {
	// Database setup
	db, err := sql.Open("postgres", "postgres://user:password@payments_db/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Wait for database
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... attempt %d", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create tables
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL UNIQUE,
			balance DECIMAL(10, 2) NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS inbox_messages (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			payload TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT false
		);
	`); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// RabbitMQ setup with retries
	var rabbitMQ *internal.RabbitMQ
	for i := 0; i < 5; i++ {
		rabbitMQ, err = internal.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to connect to RabbitMQ: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ after retries:", err)
	}
	defer rabbitMQ.Close()

	// Initialize services
	accountRepo := internal.NewAccountRepository(db)
	inboxRepo := internal.NewInboxRepository(db)
	paymentService := internal.NewPaymentService(db, accountRepo, inboxRepo) // Исправлено количество параметров
	paymentHandler := internal.NewPaymentHandler(paymentService)

	// Setup RabbitMQ queues
	paymentRequestQueue := internal.NewRabbitMQPaymentQueue(
		rabbitMQ,
		"payments",
		"payment.request",
		"payment_requests",
	)

	paymentResponseQueue := internal.NewRabbitMQPaymentQueue(
		rabbitMQ,
		"payments",
		"payment.response",
		"payment_responses",
	)

	// Start message processors
	ctx := context.Background()
	go processPaymentRequests(ctx, paymentRequestQueue, paymentService)
	go processInboxMessages(ctx, inboxRepo, paymentService, paymentResponseQueue)

	// HTTP routes
	http.HandleFunc("/payments/create-account", paymentHandler.CreateAccount)
	http.HandleFunc("/payments/get-account", paymentHandler.GetAccount)
	http.HandleFunc("/payments/deposit", paymentHandler.Deposit)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Payment service is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func processPaymentRequests(ctx context.Context, queue *internal.RabbitMQPaymentQueue, paymentService internal.PaymentService) {
	log.Println("Starting payment request processor...")

	err := queue.SubscribeToPaymentUpdates(ctx, func(orderID, userID string, amount float64) {
		log.Printf("Processing payment request: OrderID=%s, UserID=%s, Amount=%.2f", orderID, userID, amount)

		// Process payment
		result, err := paymentService.ProcessOrderPayment(ctx, orderID, userID, amount)
		if err != nil {
			log.Printf("Payment processing failed: %v", err)
			return
		}

		log.Printf("Payment processed: OrderID=%s, Success=%v", orderID, result.Success)
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to payment requests: %v", err)
	}
}

func processInboxMessages(ctx context.Context, inboxRepo internal.InboxRepository, paymentService internal.PaymentService, responseQueue *internal.RabbitMQPaymentQueue) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := inboxRepo.GetUnprocessedMessages(ctx)
			if err != nil {
				log.Printf("Failed to get unprocessed messages: %v", err)
				continue
			}

			for _, msg := range messages {
				var task struct {
					OrderID string  `json:"order_id"`
					UserID  string  `json:"user_id"`
					Amount  float64 `json:"amount"`
				}

				if err := json.Unmarshal([]byte(msg.Payload), &task); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}

				result, err := paymentService.ProcessOrderPayment(ctx, task.OrderID, task.UserID, task.Amount)
				if err != nil {
					log.Printf("Failed to process payment: %v", err)
					continue
				}

				response := map[string]interface{}{
					"order_id": task.OrderID,
					"success":  result.Success,
				}

				responseBytes, err := json.Marshal(response)
				if err != nil {
					log.Printf("Failed to marshal response: %v", err)
					continue
				}

				if err := responseQueue.PublishPaymentRequest(ctx, responseBytes); err != nil {
					log.Printf("Failed to publish response: %v", err)
					continue
				}

				if err := inboxRepo.MarkMessageAsProcessed(ctx, msg.ID); err != nil {
					log.Printf("Failed to mark message as processed: %v", err)
				}
			}
		}
	}
}
