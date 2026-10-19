package main

import (
	"log"
	"net/http"

	"cmpe273-assignment-4/internal/service"
	"cmpe273-assignment-4/internal/store"
)

func main() {
	log.Printf("inventory service on :8081")
	log.Fatal(http.ListenAndServe(":8081", service.NewInventoryHandler(store.NewInventory())))
}
