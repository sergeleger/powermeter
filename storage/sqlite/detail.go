package sqlite

import (
	"context"
	"errors"
)

type DetailRequest struct {
	MeterID int64
	Details bool
	Year    uint
	Month   uint
	Day     uint
	Hour    *uint
}

type DetailResult struct {
	Consumption int64 `json:"consumption" db:"consumption"`
	MeterID     int64 `json:"meter" db:"meter_id"`
	Year        int   `json:"year" db:"year"`
	Month       int   `json:"month,omitempty" db:"month"`
	Day         int   `json:"day,omitempty" db:"day"`
	Seconds     int   `json:"seconds,omitempty" db:"seconds"`
}

func Detail(ctx context.Context, db *Database, req DetailRequest) ([]DetailResult, error) {
	if req.MeterID <= 0 {
		return nil, errors.New("error: bad request, missing meter ID")
	}

	stmt := `select
		meter_id, consumption, year, month, day, seconds from power
		where meter_id=? `

	var params []any
	switch {
	case req.Year > 0 && req.Month > 0 && req.Day > 0 && req.Hour != nil:
		stmt += `and year=? and month=? and day=? and seconds>=? and seconds<?`
		params = []any{req.MeterID, req.Year, req.Month, req.Day, *req.Hour * 3600, (*req.Hour + 1) * 3600}

	case req.Year > 0 && req.Month > 0 && req.Day > 0:
		stmt += `and year=? and month=? and day=?`
		params = []any{req.MeterID, req.Year, req.Month, req.Day}

	case req.Year > 0 && req.Month > 0:
		stmt += `and year=? and month=?`
		params = []any{req.MeterID, req.Year, req.Month}

	case req.Year > 0:
		stmt += `and year=?`
		params = []any{req.MeterID, req.Year}

	default:
		params = []any{req.MeterID}
	}

	var dest []DetailResult
	return dest, db.SelectContext(ctx, &dest, stmt, params...)
}
