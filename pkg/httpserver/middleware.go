package httpserver

import (
	"fmt"
	"net/http"
	"time"
)

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(startTime))
	})
}

// Simple logging may use the logger pkg later
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
