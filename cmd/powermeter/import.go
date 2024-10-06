package main

import (
	"context"
	"io"
	"slices"

	"github.com/jmoiron/sqlx"
	"github.com/sergeleger/powermeter"
	"github.com/sergeleger/powermeter/ioutil"
	"github.com/sergeleger/powermeter/storage/sqlite"
	"github.com/urfave/cli/v2"
)

// ImportCommand accepts power measurement from standard input and writes them to the
// database.
var ImportCommand = cli.Command{
	Name:      "import",
	Usage:     "import JSONL files",
	ArgsUsage: "[ files ]",
	Action:    importAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "power.db", Usage: "SQLite file", Required: true},
		&cli.IntFlag{Name: "batch", Value: 100, Usage: "transaction batch size"},
		&cli.Int64Flag{Name: "meter", Usage: "Accept only `meter_id` measurements."},
	},
}

func importAction(c *cli.Context) error {
	// if no input file is provided use standard input.
	args := c.Args().Slice()
	if len(args) == 0 {
		args = append(args, "-")
	}

	// Connect to SQLite db
	db := sqlite.NewDatabase(c.String("db"))
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	batchSize := max(1, c.Int("batch"))
	del := newDeleteFilter(c.Int64("meter"))

	var measurements []powermeter.Measurement
	var err error
	for _, f := range args {
		measurements, err = ioutil.ReadFrom(f, func(r io.Reader) ([]powermeter.Measurement, error) {
			measurements, err := ioutil.ReadJSONL(measurements, r)
			if err != nil {
				return nil, err
			}

			measurements = slices.DeleteFunc(measurements, del)
			var i int

			err = db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
				for i = 0; i+batchSize < len(measurements); i += batchSize {
					err := db.Insert(ctx, tx, measurements[i:i+batchSize])
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}

			return measurements[i:], nil
		})

		if err != nil {
			return err
		}
	}

	if len(measurements) > 0 {
		return db.Transaction(context.Background(), func(ctx context.Context, tx *sqlx.Tx) error {
			return db.Insert(ctx, tx, measurements)
		})
	}

	return nil
}

func newDeleteFilter(meter int64) func(m powermeter.Measurement) bool {
	if meter == 0 {
		return func(powermeter.Measurement) bool { return false }
	}

	return func(m powermeter.Measurement) bool { return meter != m.MeterID }
}
