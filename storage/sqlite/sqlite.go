package sqlite

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sync"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/sergeleger/powermeter/power"
)

const homeMeterID = 18011759

//go:embed ddl/*
var ddlFS embed.FS

// Service implements methods for storing
type Service struct {
	db     *sqlitex.Pool
	ctx    context.Context
	cancel context.CancelFunc

	mu    sync.RWMutex
	cache map[int64]power.Measurement
}

// Open opens the specified SQLite file.
func Open(file string) (s *Service, err error) {
	defer func() {
		if err != nil {
			if s.cancel != nil {
				s.cancel()
			}

			if s.db != nil {
				s.db.Close()
			}
		}
	}()

	s = &Service{cache: make(map[int64]power.Measurement)}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	if s.db, err = sqlitex.Open(file, 0, 10); err != nil {
		return s, err
	}

	// Create/update the database schema
	if err = s.initDDL(); err != nil {
		return s, err
	}

	// Load cache information
	if err = s.loadCache(); err != nil {
		return s, err
	}

	s.startCacheWorker()
	return s, err
}

// Close releases the database resources.
func (s *Service) Close() error {
	defer s.db.Close()

	err := s.saveCache()
	if err != nil {
		log.Printf("cache: %v", err)
	}

	s.cancel()
	return err
}

func (s *Service) Insert(measurements []power.Measurement) (err error) {
	conn := s.db.Get(context.Background())
	defer s.db.Put(conn)
	defer sqlitex.Save(conn)(&err)

	insert := `insert into
		power(meter_id, year, month, day, seconds, consumption)
		values (?, ?, ?, ?, ?, ?)`

	for _, m := range measurements {
		m.Consumption = s.adjustConsumption(m)
		if m.Consumption == 0 {
			continue
		}

		err = sqlitex.Exec(
			conn,
			insert,
			nil,
			m.MeterID,
			m.Time.Year(),
			m.Time.Month(),
			m.Time.Day(),
			m.Time.Hour()*3600+m.Time.Minute()*60+m.Time.Second(),
			m.Consumption,
		)

		if err != nil {
			return err
		}
	}

	return err
}

var fields []string
var where []string
var groupBy []string

func init() {
	fields = make([]string, 4)
	fields[0] = ""
	fields[1] = ", month"
	fields[2] = fields[1] + ", day"
	fields[3] = fields[2] + ", (seconds / 3600) as hour"

	where = make([]string, 4)
	where[0] = ""
	where[1] = " and year=$2"
	where[2] = where[1] + " and month=$3"
	where[3] = where[2] + " and day=$4"

	groupBy = make([]string, 4)
	groupBy[0] = ""
	groupBy[1] = ", month"
	groupBy[2] = groupBy[1] + ", day"
	groupBy[3] = groupBy[2] + ", hour"
}

func (s *Service) Summary(details bool, args ...int) (interface{}, error) {
	conn := s.db.Get(context.Background())
	defer s.db.Put(conn)

	n := len(args)
	var query string
	if details {
		query = `select
			meter_id, consumption, year, month, day, seconds
		from
			power
		where
			meter_id = $1 ` + where[n]
	} else {
		query = `select
			meter_id, sum(consumption) as consumption, year ` + fields[n] + `
		from
			power
		where
			meter_id = $1 ` + where[n] + `
		group by
			meter_id, year ` + groupBy[n]
	}

	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	stmt.SetInt64("$1", homeMeterID)
	for i, a := range args {
		stmt.SetInt64(fmt.Sprintf("$%d", i+2), int64(a))
	}

	var hasRow bool
	var measurements = make([]powerJson, 0, 100)
	for {
		if hasRow, err = stmt.Step(); err != nil {
			return nil, err
		}

		if !hasRow {
			break
		}

		var u powerJson
		u.Consumption = stmt.GetInt64("consumption")
		u.MeterID = stmt.GetInt64("meter_id")
		u.Year = int(stmt.GetInt64("year"))
		u.Month = int(stmt.GetInt64("month"))
		u.Day = int(stmt.GetInt64("day"))
		if details {
			var t int = int(stmt.GetInt64("seconds"))
			u.Seconds = &t
		} else if n >= 3 {
			var t int = int(stmt.GetInt64("hour"))
			u.Hour = &t
		}
		measurements = append(measurements, u)
	}

	return measurements, nil
}

type powerJson struct {
	Consumption int64 `json:"consumption"`
	MeterID     int64 `json:"meter"`
	Year        int   `json:"year"`
	Month       int   `json:"month,omitempty"`
	Day         int   `json:"day,omitempty"`
	Seconds     *int  `json:"seconds,omitempty"`
	Hour        *int  `json:"hour,omitempty"`
}
