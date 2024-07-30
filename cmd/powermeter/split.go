package main

import (
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/sergeleger/powermeter"
	"github.com/sergeleger/powermeter/ioutil"
	"github.com/urfave/cli/v2"
)

// SplitCommand reads power measurement and splits the data into monthly files
var SplitCommand = cli.Command{
	Name:      "split",
	Usage:     "splits measurements into monthly files",
	ArgsUsage: "[ files... ]",
	Action:    splitAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "destination", Required: true, Usage: "destination `directory`"},
	},
}

func splitAction(c *cli.Context) error {
	// if no input file is provided use standard input.
	args := c.Args().Slice()
	if len(args) == 0 {
		args = append(args, "-")
	}

	// create destination directory
	dest := c.String("destination")
	if err := os.MkdirAll(dest, 0777); err != nil {
		return err
	}

	var err error
	var measurements []powermeter.Measurement
	for _, f := range args {
		measurements, err = ioutil.ReadFrom(f, func(r io.Reader) ([]powermeter.Measurement, error) {
			return ioutil.ReadJSONL(measurements, r)
		})

		if err != nil {
			return err
		}
	}

	slices.SortStableFunc(measurements, func(a, b powermeter.Measurement) int {
		return a.Compare(b)
	})

	var n, i, j = len(measurements), 0, 0
	var year = measurements[i].Time.Year()
	var month = measurements[i].Time.Month()

	for i < n && j < n && err == nil {
		var y, m = measurements[j].Time.Year(), measurements[j].Time.Month()
		if y != year || m != month {
			err = createSplit(dest, measurements[i:j])
			year, month, i = y, m, j
		}

		j++
	}

	if err == nil && i != j {
		err = createSplit(dest, measurements[i:j])
	}

	return err
}

// createSplit creates/appends to a monthly file.
func createSplit(dest string, measurements []powermeter.Measurement) error {
	fname := measurements[0].Time.Format("data-2006-01.json")
	fpath := filepath.Join(dest, fname)

	w, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer w.Close()

	return ioutil.WriteJSONL(w, measurements)
}
