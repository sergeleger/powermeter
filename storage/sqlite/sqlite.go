package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/sergeleger/powermeter"
	_ "modernc.org/sqlite"
)

const homeMeterID = 18011759

//go:embed ddl/*
var ddlFS embed.FS

// Database represents the database connection.
type Database struct {
	*sqlx.DB
	dsn string

	mu     sync.RWMutex
	cache  map[int64]powermeter.Measurement
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDatabase creates a new SQLite client and creates the initial database schema.
func NewDatabase(filename string) *Database {
	db := &Database{
		dsn:   filename,
		cache: make(map[int64]powermeter.Measurement),
	}

	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

// NewMemoryDatabase creates an in-memory version of the SQLite database.
func NewMemoryDatabase() *Database {
	return NewDatabase(":memory:")
}

// Open opens the database connection.
func (db *Database) Open() (err error) {
	if db.dsn == "" {
		return errors.New("error: missing database file name")
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	if db.DB, err = sqlx.Open("sqlite", db.dsn); err != nil {
		return err
	}

	// Enable WAL, foreign key checks and set busy timeout.
	_, err = db.Exec(`PRAGMA journal_mode=wal;
		PRAGMA foreign_keys=ON;
		PRAGMA busy_timeout=5000;`)
	if err != nil {
		return fmt.Errorf("error: configuring the database: %w", err)
	}

	if err := migrateSchema(db); err != nil {
		return err
	}

	if err := db.loadCache(); err != nil {
		return err
	}

	db.startCacheWorker()
	return nil
}

func (db *Database) Close() error {
	db.cancel()
	if db.DB == nil {
		return nil
	}

	err := db.saveCache()
	if err != nil {
		log.Printf("cache error: %v", err)
	}

	return db.DB.Close()
}

func (db *Database) Transaction(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return err
	}

	if err = fn(ctx, tx); err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func (db *Database) Insert(ctx context.Context, tx *sqlx.Tx, measurements []powermeter.Measurement) error {
	stmt := `insert into
		power(meter_id, year, month, day, seconds, consumption)
		values (?, ?, ?, ?, ?, ?)`

	for _, m := range measurements {
		m.Consumption = db.adjustConsumption(m)
		if m.Consumption == 0 {
			continue
		}

		_, err := tx.ExecContext(ctx, stmt,
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

	return nil
}
