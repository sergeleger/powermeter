package handler

import (
	"io/fs"
	"net/http"

	"github.com/sergeleger/powermeter/storage/sqlite"
)

func addRoutes(mux *http.ServeMux, db *sqlite.Service, htmlFS fs.FS) {
	mux.HandleFunc("GET /api/{$}", byDate(db))
	mux.HandleFunc("GET /api/{year}/{$}", byDate(db))
	mux.HandleFunc("GET /api/{year}/{month}/{$}", byDate(db))
	mux.HandleFunc("GET /api/{year}/{month}/{day}/{$}", byDate(db))
	if htmlFS != nil {
		mux.Handle("GET /", http.FileServerFS(htmlFS))
	}
}
