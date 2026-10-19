package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"cmpe273-assignment-4/internal/service"
	"cmpe273-assignment-4/internal/store"
)

func main() {
	addr := os.Getenv("INVENTORY_ADDR")
	if addr == "" {
		addr = "localhost:8081"
	}
	log.Printf("order service on :8080 -> http://%s", addr)
	h := service.NewOrderHandler(store.NewOrders(), "http://"+addr, &http.Client{Timeout: 2 * time.Second})
	log.Fatal(http.ListenAndServe(":8080", h))
}
