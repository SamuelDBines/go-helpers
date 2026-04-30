package httpserver

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
)

func staticHandler(assets fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		relativePath := strings.TrimPrefix(r.URL.Path, "/")
		if relativePath == "" {
			serveIndex(w, assets)
			return
		}

		if entry, err := fs.Stat(assets, relativePath); err == nil && !entry.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		serveIndex(w, assets)
	})
}

func serveIndex(w http.ResponseWriter, assets fs.FS) {
	file, err := assets.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not available", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, file)
}
