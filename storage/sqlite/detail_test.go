package sqlite

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestDetail(t *testing.T) {
	ctx := context.Background()

	var db *Database
	t.Run("Setup", func(t *testing.T) {
		db = testSetup(t)
	})
	defer db.Close()

	t.Run("InvalidRequest", func(t *testing.T) {
		is := is.New(t)
		_, err := Detail(ctx, db, DetailRequest{})
		is.True(err != nil)
	})

	tests := []struct {
		Name     string
		Expected int
		Request  DetailRequest
	}{
		{"Everything", 684, DetailRequest{MeterID: 9999}},
		{"ByYear", 684, DetailRequest{MeterID: 9999, Year: 2020}},
		{"ByMonth", 684, DetailRequest{MeterID: 9999, Year: 2020, Month: 1}},
		{"ByDay", 31, DetailRequest{MeterID: 9999, Year: 2020, Month: 1, Day: 22}},
		{"ByHour", 3, DetailRequest{MeterID: 9999, Year: 2020, Month: 1, Day: 22, Hour: ptr(uint(17))}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			is := is.New(t)
			got, err := Detail(ctx, db, tc.Request)
			is.NoErr(err)
			is.Equal(len(got), tc.Expected)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
