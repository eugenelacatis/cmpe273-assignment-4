package store

import "testing"

func TestInventoryCheck(t *testing.T) {
	inv := NewInventory()
	cases := []struct {
		item         string
		qty          int
		known, avail bool
	}{
		{"widget", 2, true, true},
		{"widget", 10, true, true},
		{"widget", 11, true, false},
		{"gadget", 1, true, false},
		{"nope", 1, false, false},
	}
	for _, c := range cases {
		known, avail, _ := inv.Check(c.item, c.qty)
		if known != c.known || avail != c.avail {
			t.Errorf("%s x%d -> known=%v avail=%v", c.item, c.qty, known, avail)
		}
	}
}

func TestOrdersCreateGet(t *testing.T) {
	o := NewOrders()
	a := o.Create("widget", 1)
	b := o.Create("gizmo", 2)
	if a.ID == b.ID {
		t.Fatalf("duplicate id %s", a.ID)
	}
	got, ok := o.Get(b.ID)
	if !ok || got != b {
		t.Fatalf("get %s -> %+v %v", b.ID, got, ok)
	}
	if _, ok := o.Get("missing"); ok {
		t.Fatal("found missing order")
	}
}
