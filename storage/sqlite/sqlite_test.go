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

func TestNewMemoryDatabase(t *testing.T) {
	is := is.New(t)

	db := NewMemoryDatabase()
	err := db.Open()
	is.NoErr(err)

	err = migrateSchema(db)
	is.NoErr(err)

	err = db.Close()
	is.NoErr(err)
}

func TestInsert(t *testing.T) {
	is := is.New(t)

	db := NewMemoryDatabase()
	err := db.Open()
	is.NoErr(err)
	defer db.Close()

	var usage []powermeter.Measurement
	buf, err := os.ReadFile("testdata/consumption.json")
	is.NoErr(err)

	err = json.Unmarshal(buf, &usage)
	is.NoErr(err)

	err = db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
		return db.Insert(ctx, tx, usage)
	})
	is.NoErr(err)
}
