package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli"
)

// ImportCommand collects power usage information from the command line.
var ImportCommand = cli.Command{
	Name:      "import",
	Usage:     "imports raw SCM data",
	ArgsUsage: "",
	Action:    importAction,
}

func importAction(ctx *cli.Context) error {
	files := []string(ctx.Args())
	if ctx.NArg() == 0 {
		files = append(files, "-")
	}

	// Read all files provided
	var r io.ReadCloser
	var err error
	var usage []*power.Usage
	for _, f := range files {
		if r, err = open(f); err != nil {
			return errors.Wrapf(err, "could not open %s", f)
		}

		if usage, err = power.Read(usage, r); err != nil {
			return errors.Wrapf(err, "error while reading %s", f)
		}
	}

	// Read the meter cache
	cache, err := ReadCache(ctx.GlobalString("cache"))
	if err != nil {
		return errors.Wrap(err, "could not read cache")
	}

	// Create the sqlite service
	srv, err := NewService(ctx.GlobalString("db"))
	if err != nil {
		return errors.Wrap(err, "could not connect to the database")
	}

	for _, u := range usage {
		u.Consumption = cache.Update(u)
	}

	if err = srv.Insert(usage); err != nil {
		return errors.Wrap(err, "error while updating the database")
	}

	err = WriteCache(ctx.GlobalString("cache"), cache)
	return errors.Wrap(err, "error while updating the cache")
}

func open(file string) (io.ReadCloser, error) {
	if file == "-" {
		return ioutil.NopCloser(os.Stdin), nil
	} else {
		return os.Open(file)
	}
}
