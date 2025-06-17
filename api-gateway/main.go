package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	ordersURL, err := url.Parse("http://order-service:8080")
	if err != nil {
		log.Fatal("Failed to parse Order Service URL:", err)
	}
	ordersProxy := httputil.NewSingleHostReverseProxy(ordersURL)

	paymentsURL, err := url.Parse("http://payment-service:8081")
	if err != nil {
		log.Fatal("Failed to parse Payments Service URL:", err)
	}
	paymentsProxy := httputil.NewSingleHostReverseProxy(paymentsURL)

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
	log.Fatal(http.ListenAndServe(":8000", nil))
}
