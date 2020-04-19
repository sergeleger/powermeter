package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli"
)

// SplitCommand collects power usage information from standard input and outputs it to monthly file
// structure.
var SplitCommand = cli.Command{
	Name:      "split",
	Usage:     "Collects power usage details from stdin and writes it to monthly files",
	ArgsUsage: "dataDirectory",
	Action:    splitAction,
}

func splitAction(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return errors.New("error: not enough arguments")
	}

	output := ctx.Args().First()

	var f *os.File
	var month string

	// Read standard input.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		var usage power.Usage
		var err error
		if err = usage.UnmarshalJSON(sc.Bytes()); err != nil {
			log.Printf("could not marshall entry: %v", err)
			continue
		}

		if newMonth := usage.Time.Format("2006-01"); newMonth != month || f == nil {
			if f != nil {
				f.Close()
				f = nil
			}

			month = newMonth
			file := fmt.Sprintf("%s/data-%s.json", output, month)
			if f, err = os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err != nil {
				log.Printf("could not open new output file: %v", err)
				continue
			}
		}

		f.Write(sc.Bytes())
		f.WriteString("\n")
	}

	f.Close()

	return nil
}
