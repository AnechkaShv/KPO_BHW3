package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnechkaShv/KPO_BHW2/order-service/internal"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@orders_db/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
		CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL
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
	orderRepo := internal.NewOrderRepository(db)
	outboxRepo := internal.NewOutboxRepository(db)
	orderService := internal.NewOrderService(orderRepo, outboxRepo)
	orderHandler := internal.NewOrderHandler(orderService)

	// Set up HTTP routes
	http.HandleFunc("/orders/create", orderHandler.CreateOrder)
	http.HandleFunc("/orders/get", orderHandler.GetOrder)
	http.HandleFunc("/orders/list", orderHandler.ListOrders)
	http.HandleFunc("/orders/process-payment", orderHandler.ProcessPaymentEvent)

	// Start outbox processor
	go processOutboxMessages(db, orderService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order service is running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func processOutboxMessages(db *sql.DB, orderService internal.OrderService) {
	outboxRepo := internal.NewOutboxRepository(db)

	for {
		// Get unprocessed messages
		messages, err := outboxRepo.GetUnprocessedMessages(context.Background())
		if err != nil {
			log.Printf("Failed to get unprocessed messages: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Process each message
		for _, msg := range messages {
			// In a real implementation, we would send this to a message queue
			// For simplicity, we'll just log it here
			log.Printf("Processing outbox message: %s", msg.Payload)

			// Mark as processed
			if err := outboxRepo.MarkMessageAsProcessed(context.Background(), msg.ID); err != nil {
				log.Printf("Failed to mark message as processed: %v", err)
				continue
			}
		}

		time.Sleep(10 * time.Second)
	}
}
