package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
)

func openTemp(t *testing.T) (*JSONStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "marketplace.json")
	s, err := OpenJSON(path)
	if err != nil {
		t.Fatalf("OpenJSON: %v", err)
	}
	return s, path
}

func TestOpenJSONMissingFile(t *testing.T) {
	s, _ := openTemp(t)
	products, err := s.Products()
	if err != nil {
		t.Fatalf("Products: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("Products = %v, want empty", products)
	}
}

func TestOpenJSONBadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "marketplace.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenJSON(path); err == nil {
		t.Error("OpenJSON on a truncated file returned nil error")
	}
}

func TestPutProductPersists(t *testing.T) {
	s, path := openTemp(t)
	p := catalog.Product{ID: "p1", Name: "Mug", PriceCents: 1499, Stock: 4}
	if err := s.PutProduct(p); err != nil {
		t.Fatalf("PutProduct: %v", err)
	}

	reopened, err := OpenJSON(path)
	if err != nil {
		t.Fatalf("OpenJSON: %v", err)
	}
	got, err := reopened.Product("p1")
	if err != nil {
		t.Fatalf("Product: %v", err)
	}
	if got != p {
		t.Errorf("Product = %+v, want %+v", got, p)
	}
}

func TestPutProductReplaces(t *testing.T) {
	s, _ := openTemp(t)
	_ = s.PutProduct(catalog.Product{ID: "p1", Name: "Mug", Stock: 4})
	_ = s.PutProduct(catalog.Product{ID: "p1", Name: "Mug", Stock: 2})

	products, _ := s.Products()
	if len(products) != 1 || products[0].Stock != 2 {
		t.Errorf("Products = %+v, want one product with stock 2", products)
	}
}

func TestProductNotFound(t *testing.T) {
	s, _ := openTemp(t)
	if _, err := s.Product("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Product error = %v, want ErrNotFound", err)
	}
	if _, err := s.Seller("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Seller error = %v, want ErrNotFound", err)
	}
}

func TestProductsReturnsCopy(t *testing.T) {
	s, _ := openTemp(t)
	_ = s.PutProduct(catalog.Product{ID: "p1", Name: "Mug", Stock: 4})

	products, _ := s.Products()
	products[0].Stock = 0

	got, _ := s.Product("p1")
	if got.Stock != 4 {
		t.Error("changing the Products result changed the store")
	}
}
