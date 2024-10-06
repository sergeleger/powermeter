package handler

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/rs/cors"
	"github.com/sergeleger/powermeter/storage/sqlite"
)

func NewServer(db *sqlite.Database, meterID int64, htmlFS fs.FS) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, meterID, db, htmlFS)

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

func byDate(db *sqlite.Database, meterID int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		details, _ := strconv.ParseBool(r.FormValue("details"))

		var (
			rows       any
			err        error
			detailReq  sqlite.DetailRequest
			summaryReq sqlite.SummaryRequest
		)

		switch {
		case details:
			detailReq, err = newDetailRequest(meterID, r)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			rows, err = sqlite.Detail(r.Context(), db, detailReq)

		default:
			summaryReq, err = newSummaryRequest(meterID, r)
			if err != nil {
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			rows, err = sqlite.Summary(r.Context(), db, summaryReq)
		}

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(rows)
	}
}

func newSummaryRequest(meterID int64, r *http.Request) (sqlite.SummaryRequest, error) {
	req := sqlite.SummaryRequest{MeterID: meterID}
	for _, k := range []string{"year", "month", "day"} {
		v := r.PathValue(k)
		if v == "" {
			break
		}

		i, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return req, err
		}

		switch k {
		case "year":
			req.Year = uint(i)
		case "month":
			req.Month = uint(i)
		case "day":
			req.Day = uint(i)
		}
	}

	return req, nil
}

func newDetailRequest(meterID int64, r *http.Request) (sqlite.DetailRequest, error) {
	req := sqlite.DetailRequest{MeterID: meterID}
	for _, k := range []string{"year", "month", "day"} {
		v := r.PathValue(k)
		if v == "" {
			break
		}

		i, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return req, err
		}

		switch k {
		case "year":
			req.Year = uint(i)
		case "month":
			req.Month = uint(i)
		case "day":
			req.Day = uint(i)
		}
	}

	return req, nil
}
