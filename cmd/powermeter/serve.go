package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sergeleger/powermeter"
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
		&cli.StringFlag{Name: "html", Usage: "HTML content directory"},
	},
}

func serveAction(c *cli.Context) (err error) {
	db := sqlite.NewDatabase(c.String("db"))
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()

	var htmlFS fs.FS
	if html := c.String("html"); html != "" {
		htmlFS = os.DirFS(html)
	}

	srv := handler.NewServer(db, c.Int64("meter"), htmlFS)
	httpServer := &http.Server{
		Addr:         c.String("http"),
		Handler:      srv,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  10 * time.Second,
	}

	// Listen for interrupt signal to stop the HTTP server
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Listen for new events
	go func() {
		delFilter := newDeleteFilter(c.Int64("meter"))

		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() && ctx.Err() == nil && sc.Err() == nil {
			var m powermeter.Measurement
			if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
				log.Println(err)
				continue
			}

			if delFilter(m) {
				continue
			}

			err := db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
				return db.Insert(ctx, tx, []powermeter.Measurement{m})
			})
			if err != nil {
				log.Println(err)
			}
		}
	}()

	// Start the http server
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
