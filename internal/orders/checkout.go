package orders

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
	"github.com/deepsourcelabs/demo-go/internal/store"
)

var (
	ErrEmptyOrder    = errors.New("orders: order has no lines")
	ErrBadQuantity   = errors.New("orders: quantity must be positive")
	ErrOutOfStock    = errors.New("orders: not enough stock")
	ErrUnknownCoupon = errors.New("orders: unknown coupon")
)

// Coupons maps an upper-case coupon code to its percent discount.
type Coupons map[string]int

// Checkout places orders against a store.
type Checkout struct {
	Store   store.Store
	Coupons Coupons

	// Now and NewID are for tests. They default to time.Now and a random ID.
	Now   func() time.Time
	NewID func() string
}

// Place checks stock for every line, takes the stock and returns the order.
// When any line fails, no stock is taken.
func (c *Checkout) Place(lines []Line, coupon string) (Order, error) {
	if len(lines) == 0 {
		return Order{}, ErrEmptyOrder
	}
	lines = mergeLines(lines)

	var subtotal int64
	updated := make([]catalog.Product, 0, len(lines))
	for _, l := range lines {
		if l.Quantity <= 0 {
			return Order{}, fmt.Errorf("%w: %s", ErrBadQuantity, l.ProductID)
		}
		p, err := c.Store.Product(l.ProductID)
		if err != nil {
			return Order{}, err
		}
		if !p.InStock(l.Quantity) {
			return Order{}, fmt.Errorf("%w: %s has %d, want %d", ErrOutOfStock, p.ID, p.Stock, l.Quantity)
		}
		subtotal += p.PriceCents * int64(l.Quantity)
		p.Stock -= l.Quantity
		updated = append(updated, p)
	}

	discount, err := c.discount(subtotal, coupon)
	if err != nil {
		return Order{}, err
	}

	for _, p := range updated {
		if err := c.Store.PutProduct(p); err != nil {
			return Order{}, fmt.Errorf("orders: update stock for %s: %w", p.ID, err)
		}
	}

	return Order{
		ID:            c.newID(),
		Lines:         lines,
		Coupon:        strings.ToUpper(coupon),
		SubtotalCents: subtotal,
		DiscountCents: discount,
		TotalCents:    subtotal - discount,
		CreatedAt:     c.now(),
	}, nil
}

func (c *Checkout) discount(subtotal int64, code string) (int64, error) {
	if code == "" {
		return 0, nil
	}
	pct, ok := c.Coupons[strings.ToUpper(code)]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrUnknownCoupon, code)
	}
	return subtotal * int64(pct) / 100, nil
}

func (c *Checkout) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now().UTC()
}

func (c *Checkout) newID() string {
	if c.NewID != nil {
		return c.NewID()
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return "ord_" + hex.EncodeToString(b[:])
}
