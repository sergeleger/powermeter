package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sergeleger/powermeter/power"
	"github.com/urfave/cli"
)

type Server string

// HTTPCommand collects power usage information from the command line and outputs it to
// the correct data directory.
var HTTPCommand = cli.Command{
	Name:      "http",
	Usage:     "Starts the HTTP server.",
	ArgsUsage: "dataDirectory",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "port", Value: ":8080", Usage: "port to bind HTTP service"},
	},
	Action: httpAction,
}

func httpAction(ctx *cli.Context) error {
	server := Server(ctx.Args().First())

	muxRouter := mux.NewRouter()
	r := muxRouter.PathPrefix("/api").Subrouter()
	r.HandleFunc("/data", server.listDatafile)
	r.HandleFunc("/daily/{key}", func(w http.ResponseWriter, r *http.Request) { server.filter(w, r, dailyFilter) })
	r.HandleFunc("/hourly/{key}/{day}", func(w http.ResponseWriter, r *http.Request) { server.filter(w, r, hourlyFilter) })

	//TODO http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, ssi.Handler(fs.NewIgnoreDir(staticDir))))
	http.Handle("/api/", handlers.CombinedLoggingHandler(os.Stdout, muxRouter))

	return http.ListenAndServe(ctx.String("port"), nil)
}

// listDatafile returns a list of years/month entries which have data.
func (data Server) listDatafile(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(string(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]string, 0, len(files))
	for _, f := range files {
		result = append(result, f.Name())
	}

	json.NewEncoder(w).Encode(result)
}

// daily returns the daily power usage
func (data Server) filter(w http.ResponseWriter, r *http.Request, filter func(*power.Usage) string) {
	file := mux.Vars(r)["key"]
	f, err := os.Open(filepath.Join(string(data), file))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	defer f.Close()

	// read the data file
	results, err := power.Read(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	// extract the daily information
	sort.Slice(results, func(i, j int) bool {
		if results[i].MeterID == results[j].MeterID {
			return results[i].Time.Before(results[j].Time)
		}

		return results[i].MeterID < results[j].MeterID
	})

	results = findFirstEntries(results, filter)
	json.NewEncoder(w).Encode(results)
	return

	// n := len(indexes)

	// w.Write([]byte{'['})

	// enc := json.NewEncoder(w)
	// for i := 0; i < n; i++ {
	// 	first, last := indexes[i], -1
	// 	if i+1 < n {
	// 		last = indexes[i+1] - 1
	// 	} else {
	// 		last = len(results) - 1
	// 	}

	// 	// skip entries that only have a single entry
	// 	if first == last {
	// 		continue
	// 	}

	// 	results[last].Consumption -= results[first].Consumption
	// 	if err := enc.Encode(results[last]); err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}

	// 	if i+1 < n {
	// 		w.Write([]byte{','})
	// 	}
	// }

	// w.Write([]byte{']'})
}

func findFirstEntries(results []*power.Usage, fn func(*power.Usage) string) []*power.Usage {
	var lastKey string
	var lastMeterID int

	first, j := -1, 0

	for i := range results {
		key := fn(results[i])
		if results[i].MeterID != lastMeterID || key != lastKey {
			lastMeterID = results[i].MeterID
			lastKey = key
			last := i - 1
			if first == -1 || first == last {
				first = i
				continue
			}

			results[last].Consumption -= results[first].Consumption
			results[j] = results[last]
			first = i
			j++
		}
	}

	return results[:j]
}

func monthlyFilter(u *power.Usage) string {
	return u.Time.Format("200601")
}

func dailyFilter(u *power.Usage) string {
	return u.Time.Format("20060102")
}

func hourlyFilter(u *power.Usage) string {
	return u.Time.Format("2006010215")
}
