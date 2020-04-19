package main

import (
	"database/sql/driver"
	"log"
	"strconv"
	"time"
)

// Time is a wrapper around the time.Time object. It implements database/sql.Scanner and
// database/sql/driver.Valuer interfaces.
type Time struct {
	time.Time
}

// NewTime wraps a time.Time object.
func NewTime(t time.Time) Time {
	return Time{t}
}

// Value implements database/sql/driver.Valuer and converts the time object into an Unix time value.
func (t Time) Value() (driver.Value, error) {
	if !t.IsZero() {
		return driver.Value(int64(t.UTC().Unix())), nil
	}

	return driver.Value(int64(0)), nil
}

// Scan implements database/sql.Scanner converts a incoming Unix time representation to a Time object.
func (t *Time) Scan(src interface{}) error {
	log.Println(string(src.([]byte)))
	i64, _ := strconv.ParseInt(string(src.([]byte)), 0, 64)
	t.Time = time.Unix(i64, 0).Local()
	log.Println(t.Time)
	return nil
}

// parseTime parses the time string and returns a drug.Time value.
func parseTime(format, value string) (Time, error) {
	t, err := time.Parse(format, value)
	return Time{t}, err
}
