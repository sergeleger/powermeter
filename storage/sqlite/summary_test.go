package sqlite

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/matryer/is"
	"github.com/sergeleger/powermeter"
)

func TestSummary(t *testing.T) {
	ctx := context.Background()

	var db *Database
	t.Run("Setup", func(t *testing.T) {
		db = testSetup(t)
	})
	defer db.Close()

	t.Run("InvalidRequest", func(t *testing.T) {
		is := is.New(t)
		_, err := Summary(ctx, db, SummaryRequest{})
		is.True(err != nil)
	})

	tests := []struct {
		Name     string
		Expected int
		Request  SummaryRequest
	}{
		{"ByYear", 1, SummaryRequest{MeterID: 9999}},
		{"ByMonth", 1, SummaryRequest{MeterID: 9999, Year: 2020}},
		{"ByDay", 27, SummaryRequest{MeterID: 9999, Year: 2020, Month: 1}},
		{"ByHour", 15, SummaryRequest{MeterID: 9999, Year: 2020, Month: 1, Day: 22}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			is := is.New(t)
			got, err := Summary(ctx, db, tc.Request)
			is.NoErr(err)
			is.Equal(len(got), tc.Expected)
		})
	}
}

func testSetup(t *testing.T) *Database {
	is := is.New(t)

	db := NewMemoryDatabase()
	err := db.Open()
	is.NoErr(err)

	var usage []powermeter.Measurement
	buf, err := os.ReadFile("testdata/consumption.json")
	is.NoErr(err)

	err = json.Unmarshal(buf, &usage)
	is.NoErr(err)

	err = db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
		return db.Insert(ctx, tx, usage)
	})
	is.NoErr(err)

	return db
}
