package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого origin (для разработки)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обрабатываем предварительные OPTIONS запросы
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Настройка прокси для Order Service
	ordersURL, err := url.Parse("http://order-service:8080")
	if err != nil {
		log.Fatal("Failed to parse Order Service URL:", err)
	}
	ordersProxy := httputil.NewSingleHostReverseProxy(ordersURL)

	// Настройка прокси для Payments Service
	paymentsURL, err := url.Parse("http://payment-service:8081")
	if err != nil {
		log.Fatal("Failed to parse Payments Service URL:", err)
	}
	paymentsProxy := httputil.NewSingleHostReverseProxy(paymentsURL)

	// Оборачиваем обработчики в CORS middleware
	http.HandleFunc("/orders/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Routing order request: %s %s", r.Method, r.URL.Path)
		ordersProxy.ServeHTTP(w, r)
	}))

	http.HandleFunc("/payments/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Routing payment request: %s %s", r.Method, r.URL.Path)
		paymentsProxy.ServeHTTP(w, r)
	}))

	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Gateway is healthy"))
	}))

	log.Println("API Gateway started on port 8000")
	log.Println("Order Service endpoint:", ordersURL)
	log.Println("Payment Service endpoint:", paymentsURL)

	port := "8000"
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start API Gateway:", err)
	}
}
