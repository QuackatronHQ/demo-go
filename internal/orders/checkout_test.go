package orders

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
	"github.com/deepsourcelabs/demo-go/internal/store"
)

var placedAt = time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)

func newCheckout(t *testing.T, products ...catalog.Product) (*Checkout, *store.JSONStore) {
	t.Helper()
	st, err := store.OpenJSON(filepath.Join(t.TempDir(), "marketplace.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range products {
		if err := st.PutProduct(p); err != nil {
			t.Fatal(err)
		}
	}
	c := &Checkout{
		Store:   st,
		Coupons: Coupons{"SPRING10": 10},
		Now:     func() time.Time { return placedAt },
		NewID:   func() string { return "ord_test" },
	}
	return c, st
}

func stockOf(t *testing.T, st store.Store, id string) int {
	t.Helper()
	p, err := st.Product(id)
	if err != nil {
		t.Fatal(err)
	}
	return p.Stock
}

var (
	mug  = catalog.Product{ID: "mug", Name: "Mug", PriceCents: 1499, Stock: 10}
	lamp = catalog.Product{ID: "lamp", Name: "Lamp", PriceCents: 4900, Stock: 1}
)

func TestPlaceAppliesCoupon(t *testing.T) {
	c, st := newCheckout(t, mug)
	o, err := c.Place([]Line{{ProductID: "mug", Quantity: 2}}, "spring10")
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	want := Order{
		ID:            "ord_test",
		Lines:         []Line{{ProductID: "mug", Quantity: 2}},
		Coupon:        "SPRING10",
		SubtotalCents: 2998,
		DiscountCents: 299,
		TotalCents:    2699,
		CreatedAt:     placedAt,
	}
	if o.ID != want.ID || o.Coupon != want.Coupon || o.SubtotalCents != want.SubtotalCents ||
		o.DiscountCents != want.DiscountCents || o.TotalCents != want.TotalCents || !o.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("Place = %+v, want %+v", o, want)
	}
	if got := stockOf(t, st, "mug"); got != 8 {
		t.Errorf("mug stock = %d, want 8", got)
	}
}

func TestPlaceMergesDuplicateLines(t *testing.T) {
	c, st := newCheckout(t, mug)
	o, err := c.Place([]Line{{"mug", 1}, {"mug", 2}}, "")
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if len(o.Lines) != 1 || o.Lines[0].Quantity != 3 {
		t.Errorf("Lines = %+v, want one line of 3", o.Lines)
	}
	if got := stockOf(t, st, "mug"); got != 7 {
		t.Errorf("mug stock = %d, want 7", got)
	}
}

func TestPlaceOutOfStockTakesNothing(t *testing.T) {
	c, st := newCheckout(t, mug, lamp)
	_, err := c.Place([]Line{{"mug", 1}, {"lamp", 2}}, "")
	if !errors.Is(err, ErrOutOfStock) {
		t.Fatalf("Place error = %v, want ErrOutOfStock", err)
	}
	if got := stockOf(t, st, "mug"); got != 10 {
		t.Errorf("mug stock = %d, want 10", got)
	}
}

func TestPlaceUnknownCouponTakesNothing(t *testing.T) {
	c, st := newCheckout(t, mug)
	_, err := c.Place([]Line{{"mug", 1}}, "BOGUS")
	if !errors.Is(err, ErrUnknownCoupon) {
		t.Fatalf("Place error = %v, want ErrUnknownCoupon", err)
	}
	if got := stockOf(t, st, "mug"); got != 10 {
		t.Errorf("mug stock = %d, want 10", got)
	}
}

func TestPlaceRejectsBadInput(t *testing.T) {
	c, _ := newCheckout(t, mug)
	if _, err := c.Place(nil, ""); !errors.Is(err, ErrEmptyOrder) {
		t.Errorf("Place(nil) error = %v, want ErrEmptyOrder", err)
	}
	if _, err := c.Place([]Line{{"mug", 0}}, ""); !errors.Is(err, ErrBadQuantity) {
		t.Errorf("Place(qty 0) error = %v, want ErrBadQuantity", err)
	}
	if _, err := c.Place([]Line{{"missing", 1}}, ""); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Place(missing) error = %v, want store.ErrNotFound", err)
	}
}

func TestNewIDDefault(t *testing.T) {
	c := &Checkout{}
	a, b := c.newID(), c.newID()
	if a == b || len(a) != len("ord_")+16 {
		t.Errorf("newID returned %q and %q", a, b)
	}
}
