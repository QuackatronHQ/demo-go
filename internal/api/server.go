// Package api serves the marketplace over HTTP.
package api

import (
	"log/slog"
	"net/http"

	"github.com/deepsourcelabs/demo-go/internal/orders"
	"github.com/deepsourcelabs/demo-go/internal/store"
)

// Server holds the dependencies shared by the handlers.
type Server struct {
	store    store.Store
	checkout *orders.Checkout
	apiKey   string
	log      *slog.Logger
}

// New returns the marketplace HTTP handler. Routes that change data need
// apiKey as a bearer token.
//
//	GET   /healthz
//	GET   /products?offset=&limit=
//	GET   /products/{id}
//	PATCH /products/{id}        (key)
//	GET   /sellers?city=
//	POST  /orders               (key)
func New(st store.Store, co *orders.Checkout, apiKey string, log *slog.Logger) http.Handler {
	s := &Server{store: st, checkout: co, apiKey: apiKey, log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /products", s.listProducts)
	mux.HandleFunc("GET /products/{id}", s.getProduct)
	mux.Handle("PATCH /products/{id}", s.requireKey(http.HandlerFunc(s.updateProduct)))
	mux.HandleFunc("GET /sellers", s.listSellers)
	mux.Handle("POST /orders", s.requireKey(http.HandlerFunc(s.placeOrder)))

	return s.logRequests(s.recoverPanics(mux))
}
