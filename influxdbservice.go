package main

import (
	"strconv"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"

	"github.com/sergeleger/powermeter/power"
)

type InfluxDBService struct {
	client client.Client
}

// NewInfluxDBService creates a connection to a SQLite database.
func NewInfluxDBService(connection string) (*InfluxDBService, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: connection})
	if err != nil {
		return nil, err
	}

	return &InfluxDBService{client: c}, nil
}

// Close releases all resources.
func (s *InfluxDBService) Close() error {
	return s.client.Close()
}

// Insert adds new entries to the table
func (s *InfluxDBService) Insert(usage []*power.Usage) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "mydb",
		Precision: "ms",
	})
	if err != nil {
		return err
	}

	for _, u := range usage {
		if u.Consumption == 0 {
			continue
		}

		p, err := client.NewPoint(
			"power",
			map[string]string{
				"meter": strconv.Itoa(u.MeterID),
			},
			map[string]interface{}{
				"consumption": u.Consumption,
			},
			u.Time.UTC(),
		)
		if err != nil {
			return err
		}

		bp.AddPoint(p)
	}

	return s.client.Write(bp)
}
