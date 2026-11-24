// Package store persists the marketplace catalog.
package store

import (
	"errors"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
)

// ErrNotFound is returned when a product or seller does not exist.
var ErrNotFound = errors.New("store: not found")

// Store reads and writes catalog data. Implementations are safe for
// concurrent use.
type Store interface {
	Products() ([]catalog.Product, error)
	Product(id string) (catalog.Product, error)
	PutProduct(p catalog.Product) error
	Sellers() ([]catalog.Seller, error)
	Seller(id string) (catalog.Seller, error)
}
