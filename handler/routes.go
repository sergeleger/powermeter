package handler

import (
	"io/fs"
	"net/http"

	"github.com/sergeleger/powermeter/storage/sqlite"
)

func addRoutes(mux *http.ServeMux, meterID int64, db *sqlite.Database, htmlFS fs.FS) {
	mux.HandleFunc("GET /api/{$}", byDate(db, meterID))
	mux.HandleFunc("GET /api/{year}/{$}", byDate(db, meterID))
	mux.HandleFunc("GET /api/{year}/{month}/{$}", byDate(db, meterID))
	mux.HandleFunc("GET /api/{year}/{month}/{day}/{$}", byDate(db, meterID))
	if htmlFS != nil {
		mux.Handle("GET /", http.FileServerFS(htmlFS))
	}
}
