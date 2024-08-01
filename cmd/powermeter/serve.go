package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sergeleger/powermeter/handler"
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
		&cli.StringFlag{Name: "db", Value: "power.db", Usage: "SQLite file", Required: true},
		&cli.StringFlag{Name: "http", Usage: "start HTTP service at `addr`ess. (example: localhost:8088)"},
		&cli.Int64Flag{Name: "meter", Usage: "Accept only `meter_id` measurements.", EnvVars: []string{"POWERMETER_FILTER"}},
		&cli.IntFlag{Name: "batch", Value: 1, Usage: "transaction batch size, small values for live updating"},
		&cli.StringFlag{Name: "html", Usage: "HTML content directory"},
	},
}

func serveAction(c *cli.Context) (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Connect to SQLite service
	service, err := sqlite.Open(c.String("db"))
	if err != nil {
		return err
	}
	defer service.Close()

	var htmlFS fs.FS
	if html := c.String("html"); html != "" {
		htmlFS = os.DirFS(html)
	}

	srv := handler.NewServer(service, htmlFS)
	httpServer := &http.Server{
		Addr:         c.String("http"),
		Handler:      srv,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

//
// func serveAction(c *cli.Context) (err error) {

//
// 	// start API service in the background
// 	var httpService APIService
// 	if c.IsSet("http") {
// 		go func() {
// 			httpService = NewAPIService(service)
// 			err := httpService.Listen(ctx, c.String("http"))
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		}()
// 	}
//
// 	// Start accepting entries from standard input
// 	ch := make(chan power.Measurement, 5)
// 	go func() {
// 		log.Println("Start scanning.")
//
// 		// Create filtering method
// 		accept := func(id int64) bool { return true }
// 		if c.IsSet("meter") {
// 			meter := c.Int64("meter")
// 			accept = func(id int64) bool {
// 				return id == meter
// 			}
// 		}
//
// 		var err error
// 		sc := bufio.NewScanner(os.Stdin)
// 		for sc.Scan() {
// 			var m power.Measurement
// 			if err = json.Unmarshal(sc.Bytes(), &m); err != nil {
// 				log.Println(err)
// 				continue
// 			}
//
// 			if !accept(m.MeterID) {
// 				continue
// 			}
//
// 			ch <- m
// 		}
//
// 		if err := sc.Err(); err != nil {
// 			log.Println(err)
// 		}
//
// 		close(ch)
// 	}()
//
// 	i, n := 0, c.Int("batch")
// 	batch := make([]power.Measurement, n)
// 	var stop bool
// 	for !stop {
// 		select {
// 		case <-ctx.Done():
// 			stop = true
//
// 		case m, ok := <-ch:
// 			if !ok {
// 				stop = true
// 				break
// 			}
// 			batch[i] = m
// 			i++
// 			if i < n {
// 				continue
// 			}
//
// 			if err := service.Insert(batch[0:i]); err != nil {
// 				log.Println(err)
// 			}
// 			i = 0
// 		}
// 	}
//
// 	os.Stdin.Close()
//
// 	// attempt to update remaining batch
// 	if i > 0 {
// 		if err := service.Insert(batch[0:i]); err != nil {
// 			log.Println(err)
// 		}
// 	}
//
// 	log.Println("Shutting down services.")
// 	err = httpService.Shutdown()
// 	return err
// }
