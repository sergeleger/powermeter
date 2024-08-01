package handler

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"

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
