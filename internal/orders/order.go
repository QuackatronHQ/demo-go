// Package orders turns a cart into an order and updates stock.
package orders

import "time"

// Line is one product and quantity in an order.
type Line struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Order is a placed order. Amounts are in cents.
type Order struct {
	ID            string    `json:"id"`
	Lines         []Line    `json:"lines"`
	Coupon        string    `json:"coupon,omitempty"`
	SubtotalCents int64     `json:"subtotal_cents"`
	DiscountCents int64     `json:"discount_cents"`
	TotalCents    int64     `json:"total_cents"`
	CreatedAt     time.Time `json:"created_at"`
}

// mergeLines combines lines for the same product, keeping first-seen order.
func mergeLines(lines []Line) []Line {
	out := make([]Line, 0, len(lines))
	index := make(map[string]int, len(lines))
	for _, l := range lines {
		if i, ok := index[l.ProductID]; ok {
			out[i].Quantity += l.Quantity
			continue
		}
		index[l.ProductID] = len(out)
		out = append(out, l)
	}
	return out
}
