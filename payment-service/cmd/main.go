package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnechkaShv/KPO_BHW2/payment-service/internal"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize database
	db, err := sql.Open("postgres", "postgres://user:password@payments_db/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Wait for database to be ready
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

	// Create tables if not exists
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
		
		CREATE TABLE IF NOT EXISTS outbox_messages (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			payload TEXT NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT false
		);
	`); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Initialize repositories and services
	accountRepo := internal.NewAccountRepository(db)
	inboxRepo := internal.NewInboxRepository(db)
	outboxRepo := internal.NewOutboxRepository(db)
	paymentService := internal.NewPaymentService(accountRepo, inboxRepo, outboxRepo)
	paymentHandler := internal.NewPaymentHandler(paymentService)

	// Set up HTTP routes
	http.HandleFunc("/payments/create-account", paymentHandler.CreateAccount)
	http.HandleFunc("/payments/get-account", paymentHandler.GetAccount)
	http.HandleFunc("/payments/deposit", paymentHandler.Deposit)
	http.HandleFunc("/payments/process-payment", paymentHandler.ProcessPaymentTask)

	// Start inbox processor
	go processInboxMessages(db, paymentService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Payment service is running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func processInboxMessages(db *sql.DB, paymentService internal.PaymentService) {
	inboxRepo := internal.NewInboxRepository(db)
	outboxRepo := internal.NewOutboxRepository(db)

	for {
		// Get unprocessed messages
		messages, err := inboxRepo.GetUnprocessedMessages(context.Background())
		if err != nil {
			log.Printf("Failed to get unprocessed messages: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Process each message
		for _, msg := range messages {
			var task struct {
				OrderID string  `json:"order_id"`
				UserID  string  `json:"user_id"`
				Amount  float64 `json:"amount"`
			}

			if err := json.Unmarshal([]byte(msg.Payload), &task); err != nil {
				log.Printf("Failed to unmarshal message payload: %v", err)
				continue
			}

			// Process payment
			success, err := paymentService.ProcessPaymentTask(context.Background(), task.OrderID, task.UserID, task.Amount)
			if err != nil {
				log.Printf("Failed to process payment task: %v", err)
				continue
			}

			// Create payment event in outbox
			paymentEvent := internal.PaymentEvent{
				OrderID: task.OrderID,
				UserID:  task.UserID,
				Amount:  task.Amount,
				Success: success,
			}

			payload, err := json.Marshal(paymentEvent)
			if err != nil {
				log.Printf("Failed to marshal payment event: %v", err)
				continue
			}

			if err := outboxRepo.CreateOutboxMessage(context.Background(), task.OrderID, string(payload)); err != nil {
				log.Printf("Failed to create outbox message: %v", err)
				continue
			}

			// Mark as processed
			if err := inboxRepo.MarkMessageAsProcessed(context.Background(), msg.ID); err != nil {
				log.Printf("Failed to mark message as processed: %v", err)
				continue
			}
		}

		time.Sleep(10 * time.Second)
	}
}
