package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/deepsourcelabs/demo-go/internal/catalog"
	"github.com/deepsourcelabs/demo-go/internal/orders"
	"github.com/deepsourcelabs/demo-go/internal/store"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
	maxBodyBytes    = 1 << 20
)

type page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
}

type orderRequest struct {
	Lines  []orders.Line `json:"lines"`
	Coupon string        `json:"coupon,omitempty"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.store.Products()
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	offset := intParam(r, "offset", 0)
	limit := min(intParam(r, "limit", defaultPageSize), maxPageSize)
	end := min(offset+limit, len(products))

	writeJSON(w, http.StatusOK, page[catalog.Product]{
		Items:  products[offset:end],
		Total:  len(products),
		Offset: offset,
	})
}

func (s *Server) getProduct(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.Product(r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updateProduct(w http.ResponseWriter, r *http.Request) {
	var u catalog.ProductUpdate
	if err := decodeJSON(w, r, &u); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	p, err := s.store.Product(r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	p = p.Apply(u)
	if err := p.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := s.store.PutProduct(p); err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) listSellers(w http.ResponseWriter, r *http.Request) {
	sellers, err := s.store.Sellers()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if city := r.URL.Query().Get("city"); city != "" {
		sellers = catalog.SellersIn(sellers, city)
	}
	if sellers == nil {
		sellers = []catalog.Seller{}
	}
	writeJSON(w, http.StatusOK, page[catalog.Seller]{Items: sellers, Total: len(sellers)})
}

func (s *Server) placeOrder(w http.ResponseWriter, r *http.Request) {
	var req orderRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	o, err := s.checkout.Place(req.Lines, req.Coupon)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, o)
	case errors.Is(err, orders.ErrOutOfStock):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, orders.ErrEmptyOrder),
		errors.Is(err, orders.ErrBadQuantity),
		errors.Is(err, orders.ErrUnknownCoupon):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		s.internalError(w, r, err)
	}
}

// intParam returns the integer query parameter name, or def when it is
// missing or not a number.
func intParam(r *http.Request, name string, def int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
