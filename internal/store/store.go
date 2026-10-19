package store

import (
	"fmt"
	"sync"
)

// Inventory is owned by the Inventory Service only.
type Inventory struct {
	stock map[string]int
}

func NewInventory() *Inventory {
	return &Inventory{stock: map[string]int{"widget": 10, "gadget": 0, "gizmo": 3}}
}

// Check reports whether qty of itemID is in stock. known is false for items
// the inventory has never heard of.
func (i *Inventory) Check(itemID string, qty int) (known, available bool, stock int) {
	have, ok := i.stock[itemID]
	if !ok {
		return false, false, 0
	}
	return true, qty <= have, have
}

type Order struct {
	ID       string `json:"id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

// Orders is owned by the Order Service only.
type Orders struct {
	mu     sync.Mutex
	next   int
	orders map[string]Order
}

func NewOrders() *Orders {
	return &Orders{orders: map[string]Order{}}
}

func (o *Orders) Create(itemID string, qty int) Order {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.next++
	ord := Order{ID: fmt.Sprintf("ord-%d", o.next), ItemID: itemID, Quantity: qty, Status: "created"}
	o.orders[ord.ID] = ord
	return ord
}

func (o *Orders) Get(id string) (Order, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	ord, ok := o.orders[id]
	return ord, ok
}
