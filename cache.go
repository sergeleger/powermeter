package main

import (
	"encoding/gob"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sergeleger/powermeter/power"
)

// Cache keeps track of the last power consumption seen for each meter.
type Cache struct {
	sync.Mutex
	cache map[int]cacheVal
}

type cacheVal struct {
	Value float64
	TS    time.Time
	Speed float64
}

// ReadCache reads the cache from the specified file.
func ReadCache(file string) (*Cache, error) {
	f, err := os.Open(file)

	cache := &Cache{
		cache: make(map[int]cacheVal),
	}

	// if the file does not exist use the empty cache object.
	if os.IsNotExist(err) {
		return cache, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "could not open cache file: %s", file)
	}
	defer f.Close()

	err = gob.NewDecoder(f).Decode(&cache.cache)
	return cache, errors.Wrap(err, "error decoding cache")
}

// WriteCache updates the cache on disk
func WriteCache(file string, cache *Cache) error {
	cache.Lock()
	defer cache.Unlock()

	f, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "could not create file: %s", file)
	}
	defer f.Close()

	err = gob.NewEncoder(f).Encode(cache.cache)
	return errors.Wrap(err, "could not encode cache object")
}

// Update updates the consumption cache for the specified meter ID. Returns the actual usage between
// the two measurements.
func (c *Cache) Update(usage *power.Usage) float64 {
	c.Lock()
	defer c.Unlock()

	previous, ok := c.cache[usage.MeterID]

	// return zero for the first entry of this meter.
	if !ok {
		c.cache[usage.MeterID] = cacheVal{usage.Consumption, usage.Time, 0}
		return 0
	}

	// Calculate and fix consumption
	consumption := consumption(previous.Value, usage.Consumption)

	// Calculate speed for sanity check
	speed := consumption / usage.Time.Sub(previous.TS).Seconds()
	if (previous.Speed > 0 && speed > previous.Speed*100) || speed > 10 {
		return 0
	}

	c.cache[usage.MeterID] = cacheVal{usage.Consumption, usage.Time, speed}
	return consumption
}

// consumption calculates the amount of power used since the last measurement. Also, corrects the
// value when it wraps around the meter's limit.
func consumption(old, new float64) float64 {
	consumption := new - old
	if consumption >= 0 {
		return consumption
	}

	for _, ceiling := range []float64{1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10} {
		if old < ceiling {
			return consumption + ceiling
		}
	}

	return 0
}
