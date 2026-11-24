package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"sync"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
)

// JSONStore keeps the catalog in memory and writes it to a single JSON file
// on every change. It suits small catalogs and local development.
type JSONStore struct {
	mu   sync.RWMutex
	path string
	data snapshot
}

type snapshot struct {
	Products []catalog.Product `json:"products"`
	Sellers  []catalog.Seller  `json:"sellers"`
}

var _ Store = (*JSONStore)(nil)

// OpenJSON loads the catalog from path. A missing file is an empty catalog.
func OpenJSON(path string) (*JSONStore, error) {
	s := &JSONStore{path: path}
	b, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, fmt.Errorf("store: read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, fmt.Errorf("store: decode %s: %w", path, err)
	}
	return s, nil
}

func (s *JSONStore) Products() ([]catalog.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.data.Products), nil
}

func (s *JSONStore) Product(id string) (catalog.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i := s.productIndex(id)
	if i < 0 {
		return catalog.Product{}, fmt.Errorf("product %q: %w", id, ErrNotFound)
	}
	return s.data.Products[i], nil
}

// PutProduct inserts p, or replaces the product with the same ID.
func (s *JSONStore) PutProduct(p catalog.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i := s.productIndex(p.ID); i >= 0 {
		s.data.Products[i] = p
	} else {
		s.data.Products = append(s.data.Products, p)
	}
	return s.flush()
}

func (s *JSONStore) Sellers() ([]catalog.Seller, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.data.Sellers), nil
}

func (s *JSONStore) Seller(id string) (catalog.Seller, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i := slices.IndexFunc(s.data.Sellers, func(x catalog.Seller) bool { return x.ID == id })
	if i < 0 {
		return catalog.Seller{}, fmt.Errorf("seller %q: %w", id, ErrNotFound)
	}
	return s.data.Sellers[i], nil
}

// productIndex returns the index of the product with id, or -1. The caller
// must hold s.mu.
func (s *JSONStore) productIndex(id string) int {
	return slices.IndexFunc(s.data.Products, func(x catalog.Product) bool { return x.ID == id })
}

// flush writes the catalog to a temp file and renames it over s.path, so a
// crash mid-write never leaves a truncated file. The caller must hold s.mu.
func (s *JSONStore) flush() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}

	tmp, err := os.CreateTemp("", "marketplace-*.json")
	if err != nil {
		return fmt.Errorf("store: create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("store: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("store: close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("store: replace %s: %w", s.path, err)
	}
	return nil
}
