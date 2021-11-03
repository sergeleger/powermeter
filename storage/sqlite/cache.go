package sqlite

import (
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/sergeleger/powermeter/power"
)

func (s *Service) loadCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn := s.db.Get(s.ctx)
	defer s.db.Put(conn)

	stmt, err := conn.Prepare(`select meter_id, consumption, last_entry from cache`)
	if err != nil {
		return err
	}

	var hasRow bool
	for {
		if hasRow, err = stmt.Step(); err != nil {
			return err
		}

		if !hasRow {
			break
		}

		s.cache[stmt.GetInt64("meter_id")] = power.Measurement{
			Time:        time.Unix(stmt.GetInt64("last_entry"), 0),
			Consumption: stmt.GetInt64("consumption"),
			MeterID:     stmt.GetInt64("meter_id"),
		}
	}
	return nil
}

func (s *Service) startCacheWorker() {
	go func() {
		timer := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return

			case <-timer.C:
				s.saveCache()
			}
		}
	}()
}

func (s *Service) saveCache() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn := s.db.Get(s.ctx)
	defer s.db.Put(conn)

	insert := `insert into cache(meter_id, consumption, last_entry) values( ?, ?, ? )
		on conflict(meter_id) do update set consumption = excluded.consumption, last_entry = excluded.last_entry`

	for k, v := range s.cache {
		err := sqlitex.Exec(conn, insert, nil, k, v.Consumption, v.Time.Unix())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) adjustConsumption(usage power.Measurement) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, ok := s.cache[usage.MeterID]

	// First entry for this meter.
	if !ok {
		s.cache[usage.MeterID] = usage
		return 0
	}

	// Ignore older entries
	if old.Time.After(usage.Time) {
		return 0
	}

	s.cache[usage.MeterID] = usage
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

	//
	// 	newConsumption := consumption(old.Consumption, usage.Consumption)
	// 	ok = newConsumption > 0 && newConsumption < 100
	// 	if ok {
	// 		s.cache[usage.MeterID] = *usage
	// 		usage.Consumption = newConsumption
	// 	}
	// 	return ok
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
