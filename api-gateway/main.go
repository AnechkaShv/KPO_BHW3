package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	ordersURL, _ := url.Parse("http://order-service:8080")
	paymentsURL, _ := url.Parse("http://payment-service:8081")

	ordersProxy := httputil.NewSingleHostReverseProxy(ordersURL)
	paymentsProxy := httputil.NewSingleHostReverseProxy(paymentsURL)

	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len("/orders/"):]
		ordersProxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/payments/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len("/payments/"):]
		paymentsProxy.ServeHTTP(w, r)
	})

	fmt.Println("API Gateway is running on port 8000")
	http.ListenAndServe(":8000", nil)
}
