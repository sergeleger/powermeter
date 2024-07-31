package handler

import (
	"io/fs"
	"net/http"

	"github.com/rs/cors"
	"github.com/sergeleger/powermeter/storage/sqlite"
)

func NewServer(db *sqlite.Service, htmlFS fs.FS) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db, htmlFS)

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept", "Content-Disposition"},
		ExposedHeaders: []string{"Content-Disposition"},
	})

	var handler http.Handler = mux
	handler = cors.Handler(handler)

	return handler
}
