package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"cmpe273-assignment-4/internal/store"
)

type OrderRequest struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type OrderResponse struct {
	Status string       `json:"status"`
	Detail string       `json:"detail"`
	Order  *store.Order `json:"order,omitempty"`
}

// NewOrderHandler serves POST /orders (create) and GET /orders/{id} (track).
// Creating an order first queries the Inventory Service at inventoryURL.
func NewOrderHandler(orders *store.Orders, inventoryURL string, client *http.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var req OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" || req.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, OrderResponse{Status: "error", Detail: "item_id and positive quantity required"})
			return
		}
		u := fmt.Sprintf("%s/inventory/%s?quantity=%d", inventoryURL, url.PathEscape(req.ItemID), req.Quantity)
		resp, err := client.Get(u)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				log.Printf("order %q x%d: inventory timed out: %v", req.ItemID, req.Quantity, err)
				writeJSON(w, http.StatusGatewayTimeout, OrderResponse{Status: "error", Detail: "inventory service timed out"})
				return
			}
			log.Printf("order %q x%d: inventory unavailable: %v", req.ItemID, req.Quantity, err)
			writeJSON(w, http.StatusServiceUnavailable, OrderResponse{Status: "error", Detail: "inventory service unavailable, try again later"})
			return
		}
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var avail AvailabilityResponse
			if err := json.NewDecoder(resp.Body).Decode(&avail); err != nil {
				writeJSON(w, http.StatusBadGateway, OrderResponse{Status: "error", Detail: "bad inventory response"})
				return
			}
			if !avail.Available {
				writeJSON(w, http.StatusConflict, OrderResponse{Status: "rejected", Detail: "item unavailable"})
				return
			}
			ord := orders.Create(req.ItemID, req.Quantity)
			log.Printf("order %s created: %q x%d", ord.ID, ord.ItemID, ord.Quantity)
			writeJSON(w, http.StatusCreated, OrderResponse{Status: "created", Detail: "item available", Order: &ord})
		case http.StatusNotFound:
			writeJSON(w, http.StatusNotFound, OrderResponse{Status: "rejected", Detail: "unknown item"})
		default:
			writeJSON(w, http.StatusBadGateway, OrderResponse{Status: "error", Detail: "unexpected inventory response"})
		}
	})
	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		ord, ok := orders.Get(r.PathValue("id"))
		if !ok {
			writeJSON(w, http.StatusNotFound, OrderResponse{Status: "error", Detail: "order not found"})
			return
		}
		writeJSON(w, http.StatusOK, ord)
	})
	return mux
}
