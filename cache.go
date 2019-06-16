package main

import (
	"encoding/gob"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sergeleger/powermeter/power"
)

const Threshold = 1000

type Cache struct {
	sync.Mutex

	cache map[int]float64
}

// ReadCache reads the cache from the specified file.
func ReadCache(file string) (*Cache, error) {
	f, err := os.Open(file)

	cache := &Cache{
		cache: make(map[int]float64),
	}

	// if the file does not exist create a new cache object.
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
func (c *Cache) Update(usage power.Usage) float64 {
	c.Lock()

	previous, ok := c.cache[usage.MeterID]
	c.cache[usage.MeterID] = usage.Consumption
	c.Unlock()

	if !ok {
		return usage.Consumption
	}

	return consumption(previous, usage.Consumption)
}

// consumption returns the amount of power consumption while taking care of overlow values.
func consumption(old, new float64) float64 {
	if new >= old {
		return new - old
	}

	return (new + 100000) - old
}
