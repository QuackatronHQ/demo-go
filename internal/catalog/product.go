package catalog

import (
	"errors"
	"strings"
)

var (
	ErrEmptyName     = errors.New("catalog: product name is empty")
	ErrNegativePrice = errors.New("catalog: product price is negative")
	ErrNegativeStock = errors.New("catalog: product stock is negative")
)

// Product is an item a seller lists. Prices are in cents to avoid float
// rounding.
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PriceCents  int64  `json:"price_cents"`
	Description string `json:"description"`
	SellerID    string `json:"seller_id"`
	Stock       int    `json:"stock"`
}

// ProductUpdate holds the fields a PATCH request may change. A nil field
// keeps the current value.
type ProductUpdate struct {
	Name        *string `json:"name,omitempty"`
	PriceCents  *int64  `json:"price_cents,omitempty"`
	Description *string `json:"description,omitempty"`
	Stock       *int    `json:"stock,omitempty"`
}

// Validate checks the product fields that callers can set.
func (p Product) Validate() error {
	switch {
	case strings.TrimSpace(p.Name) == "":
		return ErrEmptyName
	case p.PriceCents < 0:
		return ErrNegativePrice
	case p.Stock < 0:
		return ErrNegativeStock
	}
	return nil
}

// Apply returns a copy of p with the non-nil fields of u set.
func (p Product) Apply(u ProductUpdate) Product {
	if u.Name != nil {
		p.Name = *u.Name
	}
	if u.PriceCents != nil {
		p.PriceCents = *u.PriceCents
	}
	if u.Description != nil {
		p.Description = *u.Description
	}
	if u.Stock != nil {
		p.Stock = *u.Stock
	}
	return p
}

// InStock reports whether qty units can be sold.
func (p Product) InStock(qty int) bool {
	return qty > 0 && p.Stock >= qty
}
