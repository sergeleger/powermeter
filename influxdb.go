package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli"
)

// InfluxCommand collects power usage information from standard input and stores it to InfluxDB.
var InfluxCommand = cli.Command{
	Name:      "influxdb",
	Usage:     "Collects power usage details from the standard input, sends it to InfluxDB.",
	ArgsUsage: "",
	Action:    influxdbAction,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "host", Value: "http://localhost:8086", Usage: "InfluxDB HTTP server"},
	},
}

func influxdbAction(ctx *cli.Context) error {
	// Read the consumption cache
	cache, err := ReadCache(ctx.GlobalString("cache"))
	if err != nil {
		return errors.Wrap(err, "could not read cache")
	}

	// Create connection to InfluxDB
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: ctx.String("host"),
	})
	if err != nil {
		return fmt.Errorf("error: creating InfluxDB client: %v", err)
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "power_consumption",
		Precision: "ns",
	})

	// Read standard input.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		var usage power.Usage
		var err error
		if err = usage.UnmarshalJSON(sc.Bytes()); err != nil {
			log.Printf("could not marshall entry: %v", err)
			continue
		}

		// Skip households with very large consumptions -- tends to be errors
		if usage.Consumption > 100000 {
			continue
		}

		consumption := cache.Update(usage)

		// Create a point and add to batch
		tags := map[string]string{"meter_id": strconv.Itoa(usage.MeterID)}
		fields := map[string]interface{}{
			"consumption": consumption,
		}
		pt, err := client.NewPoint("power_consumption", tags, fields, usage.Time)
		if err != nil {
			fmt.Println("Error: ", err.Error())
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	c.Write(bp)

	err = WriteCache(ctx.GlobalString("cache"), cache)
	return errors.Wrap(err, "error while updating the cache")
}
