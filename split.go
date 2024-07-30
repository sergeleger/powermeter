package main

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli/v2"
)

// SplitCommand reads power measurement and splits the data into monthly files
var SplitCommand = cli.Command{
	Name:   "split",
	Usage:  "splits measurements into monthly files",
	Action: splitAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "destination", Required: true, Usage: "destination `directory`"},
	},
}

func splitAction(c *cli.Context) error {
	// if no input file is provided,
	args := c.Args().Slice()
	if len(args) == 0 {
		args = append(args, "-")
	}

	// create destination directory
	dest := c.String("destination")
	os.MkdirAll(dest, 0777)

	var err error
	var measurements power.Measurements
	for _, f := range args {
		if err = ReadFrom(&measurements, f); err != nil {
			return err
		}
	}

	sort.Stable(&measurements)

	var n = len(measurements)
	var j = 0
	var i = 0
	var year = measurements[i].Time.Year()
	var month = measurements[i].Time.Month()
	for i < n && j < n && err == nil {
		var y, m = measurements[j].Time.Year(), measurements[j].Time.Month()
		if y != year || m != month {
			err = createSplit(dest, measurements[i:j])
			year = y
			month = m
			i = j
		}

		j++
	}

	if err == nil && i != j {
		err = createSplit(dest, measurements[i:j])
	}

	return err
}

// createSplit creates/appends to a monthly file.
func createSplit(dest string, measurements power.Measurements) error {
	fname := measurements[0].Time.Format("data-2006-01.json")
	fpath := filepath.Join(dest, fname)

	w, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer w.Close()

	return measurements.WriteTo(w)
}
