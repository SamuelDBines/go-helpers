package httpserver

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server wraps *http.Server for ListenAndServe / Shutdown ergonomics.
type Server struct {
	HTTPServer *http.Server
}

// Options configures a Chi-backed HTTP server (production-style defaults).
type Options struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	Router            http.Handler
}

// New builds an HTTP server around a Chi (or any) handler.
func New(opt Options) *Server {
	s := &http.Server{
		Addr:              opt.Addr,
		Handler:           opt.Router,
		ReadHeaderTimeout: opt.ReadHeaderTimeout,
	}
	return &Server{HTTPServer: s}
}

// NewBaseRouter returns a Chi router with common middleware: request ID, real
// IP, recover, logger, and a 15s timeout.
func NewBaseRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(15 * time.Second))
	return r
}
