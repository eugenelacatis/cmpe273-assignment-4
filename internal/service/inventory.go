package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cmpe273-assignment-4/internal/store"
)

type AvailabilityResponse struct {
	ItemID    string `json:"item_id"`
	Quantity  int    `json:"quantity"`
	Available bool   `json:"available"`
	Stock     int    `json:"stock"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// NewInventoryHandler serves the availability query:
// GET /inventory/{item_id}?quantity=N
func NewInventoryHandler(inv *store.Inventory) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /inventory/{item_id}", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PathValue("item_id")
		qty, err := strconv.Atoi(r.URL.Query().Get("quantity"))
		if err != nil || qty <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{"positive quantity required"})
			return
		}
		known, available, stock := inv.Check(itemID, qty)
		if !known {
			writeJSON(w, http.StatusNotFound, errorResponse{"unknown item"})
			return
		}
		writeJSON(w, http.StatusOK, AvailabilityResponse{ItemID: itemID, Quantity: qty, Available: available, Stock: stock})
	})
	return mux
}
