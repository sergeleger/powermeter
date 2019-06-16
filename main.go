package main

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "PowerMeter"
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		CollectorCommand,
		GraphCommand,
		SinkCommand,
		HTTPCommand,
		InfluxCommand,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "cache", Value: "cache.gob", Usage: "power consumption cache"},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
