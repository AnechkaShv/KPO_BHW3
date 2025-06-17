package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnechkaShv/KPO_BHW2/order-service/internal"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
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

	rabbitMQ, err := internal.NewRabbitMQ("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitMQ.Close()

	paymentQueue := internal.NewRabbitMQPaymentQueue(
		rabbitMQ,
		"payments",
		"payment.request",
		"payment_requests",
	)

	orderRepo := internal.NewOrderRepository(db)
	outboxRepo := internal.NewOutboxRepository(db)
	orderService := internal.NewOrderService(orderRepo, outboxRepo, paymentQueue)
	orderHandler := internal.NewOrderHandler(orderService)

	go processOutboxMessages(context.Background(), db, paymentQueue)

	go consumePaymentUpdates(context.Background(), rabbitMQ, orderService)

	r := mux.NewRouter()

	r.HandleFunc("/api/orders/create", orderHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/api/orders/get", orderHandler.GetOrder).Methods("GET")
	r.HandleFunc("/api/orders/list", orderHandler.ListOrders).Methods("GET")
	r.HandleFunc("/api/orders/process-payment", orderHandler.ProcessPaymentEvent).Methods("POST")

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order service is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler.Handler(r)))
}

func processOutboxMessages(ctx context.Context, db *sql.DB, queue *internal.RabbitMQPaymentQueue) {
	outboxRepo := internal.NewOutboxRepository(db)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := outboxRepo.GetUnprocessedMessages(ctx)
			if err != nil {
				log.Printf("Failed to get unprocessed messages: %v", err)
				continue
			}

			for _, msg := range messages {
				if err := queue.PublishPaymentRequest(ctx, []byte(msg.Payload)); err != nil {
					log.Printf("Failed to publish message: %v", err)
					continue
				}

				if err := outboxRepo.MarkMessageAsProcessed(ctx, msg.ID); err != nil {
					log.Printf("Failed to mark message as processed: %v", err)
				}
			}
		}
	}
}

func consumePaymentUpdates(ctx context.Context, rabbitMQ *internal.RabbitMQ, orderService internal.OrderService) {
	queue := internal.NewRabbitMQPaymentQueue(
		rabbitMQ,
		"payments",
		"payment.response",
		"payment_responses",
	)

	err := queue.SubscribeToPaymentUpdates(ctx, func(orderID string, success bool) {
		if err := orderService.ProcessPaymentEvent(context.Background(), orderID, success); err != nil {
			log.Printf("Failed to process payment event: %v", err)
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to payment updates:", err)
	}
}
