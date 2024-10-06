package sqlite

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sergeleger/powermeter"
)

type cacheEntry struct {
	MeterID     int64 `db:"meter_id"`
	Consumption int64 `db:"consumption"`
	LastEntry   int64 `db:"last_entry"`
}

func (db *Database) loadCache() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var cache []cacheEntry
	err := sqlx.Select(db, &cache, `select meter_id, consumption, last_entry from cache`)
	if err != nil {
		return err
	}

	for _, c := range cache {
		db.cache[c.MeterID] = powermeter.Measurement{
			Time:        time.Unix(c.LastEntry, 0),
			MeterID:     c.MeterID,
			Consumption: c.Consumption,
		}
	}

	return nil
}

func (db *Database) startCacheWorker() {
	timer := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-db.ctx.Done():
				timer.Stop()
				return

			case <-timer.C:
				db.saveCache()
			}
		}
	}()
}

func (db *Database) saveCache() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stmt := `insert into
		cache(meter_id, consumption, last_entry)
		values(:meter_id, :consumption, :last_entry)
		on conflict(meter_id) do
		update set
			consumption=excluded.consumption,
			last_entry=excluded.last_entry`

	return db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
		for _, cache := range db.cache {
			_, err := tx.NamedExec(stmt, cacheEntry{
				MeterID:     cache.MeterID,
				Consumption: cache.Consumption,
				LastEntry:   cache.Time.Unix(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *Database) adjustConsumption(usage powermeter.Measurement) int64 {
	db.mu.Lock()
	defer db.mu.Unlock()

	old, ok := db.cache[usage.MeterID]
	if !ok {
		// First entry for this meter.
		db.cache[usage.MeterID] = usage
		return 0
	}

	// Ignore older entries
	if old.Time.After(usage.Time) {
		return 0
	}

	db.cache[usage.MeterID] = usage
	consumption := consumption(old.Consumption, usage.Consumption)
	sameMonth := old.Time.Month() == usage.Time.Month() &&
		old.Time.Year() == usage.Time.Year()

	switch {
	// Refuse very large increment (normal household ~3000/month)
	case consumption > 1000:
		return 0

	// refuse large increment that span months
	case consumption > 100 && !sameMonth:
		return 0

	default:
		return consumption
	}
}

// consumption calculates the amount of power used since the last measurement. Also, corrects the
// value when it wraps around the meter's limit.
func consumption(old, new int64) int64 {
	consumption := new - old
	if consumption >= 0 {
		return consumption
	}

	for _, ceiling := range []int64{1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10} {
		if old < ceiling {
			return consumption + ceiling
		}
	}

	return 0
}
