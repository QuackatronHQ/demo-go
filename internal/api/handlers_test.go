package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
	"github.com/deepsourcelabs/demo-go/internal/orders"
	"github.com/deepsourcelabs/demo-go/internal/store"
)

const testKey = "test-key"

func newTestServer(t *testing.T) (http.Handler, *store.JSONStore) {
	t.Helper()
	seed, err := os.ReadFile("../../testdata/marketplace.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "marketplace.json")
	if err := os.WriteFile(path, seed, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := store.OpenJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	co := &orders.Checkout{Store: st, Coupons: orders.Coupons{"SPRING10": 10}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(st, co, testKey, log), st
}

func do(t *testing.T, h http.Handler, method, path, body, key string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealth(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(t, h, "GET", "/healthz", "", ""); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestListProductsPaginates(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(t, h, "GET", "/products?offset=1&limit=1", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	got := decode[page[catalog.Product]](t, rec)
	if got.Total != 3 || len(got.Items) != 1 || got.Items[0].ID != "p-lamp" {
		t.Errorf("page = %+v", got)
	}
}

func TestGetProduct(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(t, h, "GET", "/products/p-mug", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := decode[catalog.Product](t, rec); got.Name != "Bat Mug" {
		t.Errorf("product = %+v", got)
	}

	if rec := do(t, h, "GET", "/products/nope", "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("missing product status = %d, want 404", rec.Code)
	}
}

func TestUpdateProductRequiresKey(t *testing.T) {
	h, _ := newTestServer(t)
	body := `{"price_cents": 1299}`
	if rec := do(t, h, "PATCH", "/products/p-mug", body, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no key status = %d, want 401", rec.Code)
	}
	if rec := do(t, h, "PATCH", "/products/p-mug", body, "wrong"); rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong key status = %d, want 401", rec.Code)
	}
}

func TestUpdateProduct(t *testing.T) {
	h, st := newTestServer(t)
	rec := do(t, h, "PATCH", "/products/p-mug", `{"price_cents": 1299}`, testKey)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	p, _ := st.Product("p-mug")
	if p.PriceCents != 1299 || p.Name != "Bat Mug" {
		t.Errorf("stored product = %+v", p)
	}
}

func TestUpdateProductBadBody(t *testing.T) {
	h, _ := newTestServer(t)
	if rec := do(t, h, "PATCH", "/products/p-mug", `{"colour": "red"}`, testKey); rec.Code != http.StatusBadRequest {
		t.Errorf("unknown field status = %d, want 400", rec.Code)
	}
	if rec := do(t, h, "PATCH", "/products/p-mug", `{"price_cents": -5}`, testKey); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("negative price status = %d, want 422", rec.Code)
	}
}

func TestListSellersByCity(t *testing.T) {
	h, _ := newTestServer(t)
	got := decode[page[catalog.Seller]](t, do(t, h, "GET", "/sellers?city=gotham", "", ""))
	if got.Total != 1 || got.Items[0].ID != "s-gotham-goods" {
		t.Errorf("page = %+v", got)
	}
	got = decode[page[catalog.Seller]](t, do(t, h, "GET", "/sellers?city=Atlantis", "", ""))
	if got.Total != 0 || got.Items == nil {
		t.Errorf("empty page = %+v, want items []", got)
	}
}

func TestPlaceOrder(t *testing.T) {
	h, st := newTestServer(t)
	body := `{"lines": [{"product_id": "p-mug", "quantity": 2}], "coupon": "spring10"}`
	rec := do(t, h, "POST", "/orders", body, testKey)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body)
	}
	o := decode[orders.Order](t, rec)
	if o.TotalCents != 2699 || o.Coupon != "SPRING10" {
		t.Errorf("order = %+v", o)
	}
	if p, _ := st.Product("p-mug"); p.Stock != 8 {
		t.Errorf("mug stock = %d, want 8", p.Stock)
	}
}

func TestPlaceOrderErrors(t *testing.T) {
	h, _ := newTestServer(t)
	tests := []struct {
		name string
		body string
		want int
	}{
		{"out of stock", `{"lines": [{"product_id": "p-notebook", "quantity": 1}]}`, http.StatusConflict},
		{"unknown product", `{"lines": [{"product_id": "nope", "quantity": 1}]}`, http.StatusNotFound},
		{"empty", `{"lines": []}`, http.StatusBadRequest},
		{"bad coupon", `{"lines": [{"product_id": "p-mug", "quantity": 1}], "coupon": "X"}`, http.StatusBadRequest},
		{"bad json", `{"lines": `, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(t, h, "POST", "/orders", tt.body, testKey); rec.Code != tt.want {
				t.Errorf("status = %d, want %d: %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
}
