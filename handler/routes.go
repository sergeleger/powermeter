package handler

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"

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

func byDate(db *sqlite.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var args []int
		for _, k := range []string{"year", "month", "day"} {
			v := r.PathValue(k)
			if v == "" {
				break
			}

			i, err := strconv.Atoi(v)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			args = append(args, i)
		}

		var rows interface{}
		var err error
		rows, err = db.Summary(r.FormValue("details") == "true", args...)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(rows)
	}
}
