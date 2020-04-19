package main

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/sergeleger/powermeter/power"
)

const (
	insert int = iota
	monthly
	daily
	hourly
)

type Service struct {
	db    *sqlx.DB
	stmts [4]*sqlx.Stmt
}

// NewService creates a connection to a SQLite database.
func NewService(dbFile string) (*Service, error) {
	var err error
	srv := &Service{}

	var db *sql.DB
	if db, err = sql.Open("sqlite3", dbFile); err != nil {
		return nil, errors.Wrapf(err, "could not open database %q", dbFile)
	}

	// create database objects
	if _, err = db.Exec(ddl); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "could not create tables")
	}

	srv.db = sqlx.NewDb(db, "sqlite3")
	for i, stmt := range []string{insertStmt, queryByMonthStmt, queryByDayStmt, queryByHourStmt} {
		if srv.stmts[i], err = srv.db.Preparex(stmt); err != nil {
			break
		}
	}

	return srv, errors.Wrap(err, "could not prepare statements")
}

// Close releases all resources.
func (s *Service) Close() error {
	for i := range s.stmts {
		s.stmts[i].Close()
	}

	err := s.db.Close()
	return errors.Wrap(err, "error closing database")
}

// Insert adds new entries to the table
func (s *Service) Insert(usage []*power.Usage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.Wrap(err, "could not start a transaction")
	}

	st := tx.Stmt(s.stmts[insert].Stmt)
	for _, u := range usage {
		if u.Consumption == 0 {
			continue
		}

		st.Exec(u.MeterID, u.Time, u.Consumption)
	}

	return errors.Wrap(tx.Commit(), "could not commit transaction")
}

func (s *Service) QueryByMonth(meterID int, year string) ([]dbUsage, error) {
	var usage []dbUsage
	err := s.stmts[monthly].Select(&usage, meterID, year)
	log.Println(usage)
	return usage, err
}

// func (s *Service) QueryByDay(meterID, year, month int) {
// }

// func (s *Service) QueryByHour(meterID, year, month, day int) {
// }

type dbUsage struct {
	Time        string  `db:"Time"`
	MeterID     int     `db:"MeterID"`
	Consumption float64 `db:"Usage"`
}

var ddl = `
create table if not exists power (
	MeterID integer,
	Time integer,
	Usage real
);

create view if not exists hourly as
	select
		MeterID,
		strftime("%Y", Time, 'localtime') as Year,
		strftime("%m", Time, 'localtime') as Month,
		strftime("%d", Time, 'localtime') as Day,
		strftime("%H", Time, 'localtime') as Hour,
		sum(Usage) as Usage
	from
		Power
	group by 1, 2, 3, 4, 5;
`

var insertStmt = `insert into power(MeterID, Time, Usage) values(?,?,?)`

var queryByMonthStmt = `
select
	MeterID, Year || "-" || Month as Time, sum(Usage) as Usage
from
	hourly
where
	MeterID = ? and Year = ?
group by
	MeterID, Year, Month
`

var queryByDayStmt = `
select
	MeterID, Time, Usage
from
	power
where
	MeterID = ? and Time = ?
`
var queryByHourStmt = `
select
	MeterID, Time, Usage
from
	power
where
	MeterID = ? and Time = ?
`
