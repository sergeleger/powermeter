package main

import (
	"encoding/gob"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/sergeleger/powermeter/power"
)

// Cache keeps track of the last power consumption seen for each meter.
type Cache map[int]cacheVal

type cacheVal struct {
	Value float64
	TS    time.Time
	Speed float64
}

// ReadFrom reads cache from specified reader
func (c *Cache) ReadFrom(r io.Reader) error {
	if *c == nil {
		*c = make(map[int]cacheVal)
	}

	err := gob.NewDecoder(r).Decode(c)
	return errors.Wrap(err, "error decoding cache")
}

// WriteTo writes cache to the specified writer
func (c *Cache) WriteTo(w io.Writer) error {
	err := gob.NewEncoder(w).Encode(c)
	return errors.Wrap(err, "could not encode cache object")
}

// Update updates the consumption cache for the specified meter ID. Returns the actual usage between
// the two measurements.
func (c *Cache) Update(usage *power.Measurement) float64 {
	if *c == nil {
		*c = make(map[int]cacheVal)
	}

	previous, ok := (*c)[usage.MeterID]

	// return zero for the first entry of this meter.
	if !ok {
		(*c)[usage.MeterID] = cacheVal{usage.Consumption, usage.Time, 0}
		return 0
	}

	// ignore older entries
	if usage.Time.Before(previous.TS) || usage.Time.Equal(previous.TS) {
		return 0
	}

	// Calculate and fix consumption
	consumption := consumption(previous.Value, usage.Consumption)

	// Calculate speed for sanity check
	speed := consumption / usage.Time.Sub(previous.TS).Seconds()
	if (previous.Speed > 0 && speed > previous.Speed*100) || speed > 10 {
		return 0
	}

	(*c)[usage.MeterID] = cacheVal{usage.Consumption, usage.Time, speed}
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
