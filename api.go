package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/sergeleger/powermeter/storage/sqlite"
	"github.com/urfave/cli/v2"
)

//go:embed SvelteUI/public/*
var htmlFS embed.FS

const prefixPath = "SvelteUI/public/"

// APICommand provides an API server for querying the database
var APICommand = cli.Command{
	Name:   "api",
	Usage:  "start the HTTP PowerMeter server",
	Action: apiAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "db", Value: "power.db", Usage: "SQLite file"},
		&cli.StringFlag{Name: "addr", Value: "localhost:8088", Usage: "HTTP Port"},
	},
}

func apiAction(c *cli.Context) (err error) {
	// Connect to SQLite service
	db, err := sqlite.Open(c.String("db"))
	if err != nil {
		return err
	}
	defer db.Close()

	service := NewAPIService(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup termination signals and wait for
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-stop:
			log.Println("Received interrupt")
			cancel()

		case <-ctx.Done():
			log.Println("Context has been canceled.")
		}

		err := service.Shutdown()
		if err != nil {
			log.Println(err)
		}
	}()

	return service.Listen(ctx, c.String("addr"))
}

type APIService struct {
	db  *sqlite.Service
	app *fiber.App
}

func NewAPIService(db *sqlite.Service) APIService {
	service := APIService{db: db}

	// Create fiber application
	service.app = fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ReadTimeout:           5 * time.Minute,
		WriteTimeout:          2 * time.Minute,
		IdleTimeout:           10 * time.Second,
	})

	service.app.Use(cors.New(
		cors.Config{
			AllowOrigins:  "*",
			AllowMethods:  "GET,POST,DELETE,PUT,PATCH",
			AllowHeaders:  "Content-Type, Authorization, Accept, Content-Disposition",
			ExposeHeaders: "Content-Disposition",
		},
	))

	// register routes
	api := service.app.Group("/api")
	api.Get("/", service.byDate)
	api.Get("/:year", service.byDate)
	api.Get("/:year/:month", service.byDate)
	api.Get("/:year/:month/:day", service.byDate)

	// Implement static HTTP server
	// Provide a minimal config
	service.app.Use(filesystem.New(filesystem.Config{
		Root: http.FS(prefixFS{
			FS:     htmlFS,
			prefix: prefixPath,
		}),
	}))

	return service
}

func (s APIService) Listen(ctx context.Context, addr string) error {
	return s.app.Listen(addr)
}

func (s APIService) Shutdown() error {
	if s.app == nil {
		return nil
	}

	// wait for app to shutdown
	ch := make(chan error)
	go func() {
		ch <- s.app.Shutdown()
	}()

	tick := time.After(10 * time.Second)
	select {
	case <-tick:
		log.Println(errors.New("error: app shutdown took too long"))
		os.Exit(1)
		return nil

	case err := <-ch:
		return err
	}
}

// byDate returns the data based on the URL pattern: /api[/year[/month[/day]]]
func (s APIService) byDate(c *fiber.Ctx) error {
	var args []int
	for _, k := range []string{"year", "month", "day"} {
		v := c.Params(k)
		if v == "" {
			break
		}

		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}

		args = append(args, i)
	}

	var rows interface{}
	var err error
	rows, err = s.db.Summary(c.Query("details") == "true", args...)
	if err != nil {
		return err
	}

	return c.JSON(rows)
}

// prefixFS adds the go:embed directory prefix to Open() operations
type prefixFS struct {
	prefix string
	embed.FS
}

func (fs prefixFS) Open(name string) (fs.File, error) {
	return fs.FS.Open(filepath.Join(fs.prefix, name))
}
