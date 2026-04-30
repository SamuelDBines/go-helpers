package httpserver

import (
	"fmt"
	"net/http"
)

type HTTPResponse struct {
	Status  int         `json:"status",omitempty`
	Message string      `json:"message"`
	Data    interface{} `json:"data",omitempty`
}

type Config struct {
	Port int
	Name string
}

func NewServer(cfg Config, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", cfg.Port)
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated := true
		if !authenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func With(mux *http.ServeMux, path string, h http.Handler) {
	mux.Handle(path, Logger(h))
}

func HandleMiddleWare(mux *http.ServeMux, path string, next http.HandlerFunc) {
	mux.HandleFunc(path, LoggerMiddleware(AuthMiddleware(next)))
}
