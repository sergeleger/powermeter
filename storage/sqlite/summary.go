package sqlite

import (
	"context"
	"errors"
)

const (
	summaryByHour = `select
		meter_id, sum(consumption) as consumption, year, month, day, (seconds / 3600) as hour
	from power where
		meter_id=? and year=? and month=? and day=?
	group by year, month, day, hour`

	summaryByDay = `select
		meter_id, sum(consumption) as consumption, year, month, day
	from power where
		meter_id=? and year=? and month=?
	group by year, month, day`

	summaryByMonth = `select
		meter_id, sum(consumption) as consumption, year, month
	from power where
		meter_id=? and year=?
	group by year, month`

	summaryByYear = `select
		meter_id, sum(consumption) as consumption, year
	from power where
		meter_id=?
	group by year`
)

type SummaryRequest struct {
	MeterID int64
	Details bool
	Year    uint
	Month   uint
	Day     uint
	Hour    uint
}

type SummaryResult struct {
	Consumption int64 `json:"consumption" db:"consumption"`
	MeterID     int64 `json:"meter" db:"meter_id"`
	Year        int   `json:"year" db:"year"`
	Month       *int  `json:"month,omitempty" db:"month"`
	Day         *int  `json:"day,omitempty" db:"day"`
	Hour        *int  `json:"hour,omitempty" db:"hour"`
}

func Summary(ctx context.Context, db *Database, req SummaryRequest) ([]SummaryResult, error) {
	if req.MeterID <= 0 {
		return nil, errors.New("error: bad request, missing meter ID")
	}

	var stmt string
	var params []any

	switch {
	case req.Year > 0 && req.Month > 0 && req.Day > 0:
		stmt = summaryByHour
		params = []any{req.MeterID, req.Year, req.Month, req.Day}

	case req.Year > 0 && req.Month > 0:
		stmt = summaryByDay
		params = []any{req.MeterID, req.Year, req.Month}

	case req.Year > 0:
		stmt = summaryByMonth
		params = []any{req.MeterID, req.Year}

	default:
		stmt = summaryByYear
		params = []any{req.MeterID}
	}

	var dest []SummaryResult
	return dest, db.SelectContext(ctx, &dest, stmt, params...)
}
