package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeleger/powermeter/power"
	"github.com/sergeleger/powermeter/storage/sqlite"
	"github.com/urfave/cli/v2"
)

// ServeCommand accepts power measurement from standard input and writes them to the
// database.
var ServeCommand = cli.Command{
	Name:   "serve",
	Usage:  "start the PowerMeter server",
	Action: serveAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "power.db", Usage: "SQLite file"},
		&cli.StringFlag{Name: "cache", Value: "power.gob", Usage: "cache file"},
		&cli.StringFlag{Name: "http", Usage: "start HTTP service at `addr`ess. (example: localhost:8088)"},
		&cli.IntFlag{Name: "meter", Usage: "Accept only `meter_id` measurements.", EnvVars: []string{"POWERMETER_FILTER"}},
		&cli.IntFlag{Name: "batch", Value: 1, Usage: "transaction batch size, small values for live updating"},
	},
}

func serveAction(c *cli.Context) (err error) {
	// Read the meter cache
	var cache Cache
	var cacheFile = c.String("cache")
	err = ReadFrom(&cache, cacheFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Connect to SQLite service
	service, err := sqlite.Open(c.String("db"))
	if err != nil {
		return err
	}
	defer service.Close()

	// create shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup termination signals and wait for
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-stop:
			cancel()

		case <-ctx.Done():
			return
		}
	}()

	// start API service in the background
	var httpService APIService
	if c.IsSet("http") {
		go func() {
			httpService = NewAPIService(service)
			err := httpService.Listen(ctx, c.String("http"))
			if err != nil {
				log.Println(err)
			}
		}()
	}

	// Start accepting entries from standard input
	ch := make(chan power.Measurement, 5)
	go func() {
		log.Println("Start scanning.")

		// Create filtering method
		accept := func(id int) bool { return true }
		if c.IsSet("meter") {
			meter := c.Int("meter")
			accept = func(id int) bool {
				return id == meter
			}
		}

		var err error

		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			var m power.Measurement
			if err = json.Unmarshal(sc.Bytes(), &m); err != nil {
				log.Println(err)
				continue
			}

			if !accept(m.MeterID) {
				continue
			}

			ch <- m
		}

		if err := sc.Err(); err != nil {
			log.Println(err)
		}

		close(ch)
	}()

	i, n := 0, c.Int("batch")
	batch := make([]power.Measurement, n)
	timer := time.NewTicker(10 * time.Minute)
	var stop bool
	for !stop {
		select {
		case <-ctx.Done():
			stop = true

		case <-timer.C:
			if err := WriteTo(&cache, cacheFile); err != nil {
				log.Println(err)
			}

		case m, ok := <-ch:
			if !ok {
				stop = true
				break
			}
			batch[i] = m
			batch[i].Consumption = cache.Update(&batch[i])
			i++
			if i < n {
				continue
			}

			if err := service.Insert(batch[0:i]); err != nil {
				log.Println(err)
			}
			i = 0
		}
	}
	timer.Stop()

	os.Stdin.Close()

	// attempt to update remaining batch
	if i > 0 {
		if err := service.Insert(batch[0:i]); err != nil {
			log.Println(err)
		}
	}

	// update cache
	if err := WriteTo(&cache, cacheFile); err != nil {
		log.Println(err)
	}

	log.Println("Shutting down services.")
	err = httpService.Shutdown()
	return err
}
