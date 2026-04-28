package httpserver

import (
	"fmt"
	"net/http"
)

func WriteCSS(w http.ResponseWriter, status int, css string) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprint(w, css)
}
