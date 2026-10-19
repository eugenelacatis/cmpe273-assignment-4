package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cmpe273-assignment-4/internal/store"
)

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestInventoryStatuses(t *testing.T) {
	h := NewInventoryHandler(store.NewInventory())
	cases := []struct {
		path      string
		want      int
		available bool
	}{
		{"/inventory/widget?quantity=2", 200, true},
		{"/inventory/widget?quantity=99", 200, false},
		{"/inventory/gadget?quantity=1", 200, false},
		{"/inventory/nope?quantity=1", 404, false},
		{"/inventory/widget?quantity=0", 400, false},
		{"/inventory/widget", 400, false},
	}
	for _, c := range cases {
		rec := do(t, h, http.MethodGet, c.path, "")
		if rec.Code != c.want {
			t.Errorf("%s -> %d want %d", c.path, rec.Code, c.want)
			continue
		}
		if rec.Code == 200 {
			var resp AvailabilityResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Available != c.available {
				t.Errorf("%s -> available=%v", c.path, resp.Available)
			}
		}
	}
}

func orderAgainstInventory(t *testing.T) http.Handler {
	t.Helper()
	srv := httptest.NewServer(NewInventoryHandler(store.NewInventory()))
	t.Cleanup(srv.Close)
	return NewOrderHandler(store.NewOrders(), srv.URL, &http.Client{Timeout: time.Second})
}

func TestOrderStatuses(t *testing.T) {
	h := orderAgainstInventory(t)
	cases := []struct {
		body string
		want int
	}{
		{`{"item_id":"widget","quantity":2}`, 201},
		{`{"item_id":"gadget","quantity":1}`, 409},
		{`{"item_id":"widget","quantity":99}`, 409},
		{`{"item_id":"nope","quantity":1}`, 404},
		{`{"item_id":"widget","quantity":0}`, 400},
		{`{"item_id":"","quantity":1}`, 400},
		{`garbage`, 400},
	}
	for _, c := range cases {
		if got := do(t, h, http.MethodPost, "/orders", c.body).Code; got != c.want {
			t.Errorf("%s -> %d want %d", c.body, got, c.want)
		}
	}
}

func TestOrderCreatedThenTracked(t *testing.T) {
	h := orderAgainstInventory(t)
	rec := do(t, h, http.MethodPost, "/orders", `{"item_id":"gizmo","quantity":3}`)
	var resp OrderResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if rec.Code != 201 || resp.Order == nil {
		t.Fatalf("create -> %d %+v", rec.Code, resp)
	}
	if got := do(t, h, http.MethodGet, "/orders/"+resp.Order.ID, "").Code; got != 200 {
		t.Fatalf("track %s -> %d", resp.Order.ID, got)
	}
	if got := do(t, h, http.MethodGet, "/orders/missing", "").Code; got != 404 {
		t.Fatalf("missing order -> %d", got)
	}
}

func TestOrderInventoryDown503AndHandlerSurvives(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	h := NewOrderHandler(store.NewOrders(), url, &http.Client{Timeout: time.Second})
	if got := do(t, h, http.MethodPost, "/orders", `{"item_id":"widget","quantity":1}`).Code; got != 503 {
		t.Fatalf("got %d want 503", got)
	}
	if got := do(t, h, http.MethodPost, "/orders", `{"item_id":"widget","quantity":0}`).Code; got != 400 {
		t.Fatalf("handler dead after failure: got %d", got)
	}
}
