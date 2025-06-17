package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnechkaShv/KPO_BHW2/payment-service/internal"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@payments_db/postgres?sslmode=disable")
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

	accountRepo := internal.NewAccountRepository(db)
	inboxRepo := internal.NewInboxRepository(db)
	paymentService := internal.NewPaymentService(db, accountRepo, inboxRepo, paymentResponseQueue)
	paymentHandler := internal.NewPaymentHandler(paymentService)

	r := mux.NewRouter()

	r.HandleFunc("/api/payments/create-account", paymentHandler.CreateAccount).Methods("POST")
	r.HandleFunc("/api/payments/get-account", paymentHandler.GetAccount).Methods("GET")
	r.HandleFunc("/api/payments/deposit", paymentHandler.Deposit).Methods("POST")
	r.HandleFunc("/api/payments/process", paymentHandler.ProcessPayment).Methods("POST")

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

	ctx := context.Background()
	go processPaymentRequests(ctx, paymentRequestQueue, paymentService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Payment service is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler.Handler(r)))
}

func processPaymentRequests(ctx context.Context, queue *internal.RabbitMQPaymentQueue, paymentService internal.PaymentService) {
	log.Println("Starting payment request processor...")

	err := queue.SubscribeToPaymentUpdates(ctx, func(orderID, userID string, amount float64) {
		log.Printf("Processing payment request: OrderID=%s, UserID=%s, Amount=%.2f", orderID, userID, amount)

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
