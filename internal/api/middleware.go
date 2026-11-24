package api

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// statusRecorder captures the status code a handler writes.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests logs one line per request with its status and duration.
func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

// recoverPanics turns a handler panic into a 500 response and logs the stack.
func (s *Server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.log.Error("handler panic",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", v,
					"stack", string(debug.Stack()),
				)
				writeError(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requireKey rejects requests that do not carry the API key as a bearer
// token.
func (s *Server) requireKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if key != s.apiKey {
			writeError(w, http.StatusUnauthorized, "missing or invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}
