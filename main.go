package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/sergeleger/powermeter/power"
)

func main() {
	var (
		cacheFile = flag.String("cache", "cache.gob", "cache file")
		host      = flag.String("host", "http://pi4:8086/", "InfluxDB address")
		db        = flag.String("db", "mydb", "InfluxDB database name")
	)
	flag.Parse()

	// Connect to influxDB
	service, err := NewInfluxDBService(*host, *db)
	if err != nil {
		log.Fatal(err)
	}
	defer service.Close()

	// Read the meter cache
	var cache Cache
	err = ReadFrom(&cache, *cacheFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Read a single entry at a time
	sc := bufio.NewScanner(os.Stdin)

	n := 1000
	i := 0
	batch := make([]power.Usage, n)
	for sc.Scan() {
		if err := json.Unmarshal(sc.Bytes(), &batch[i]); err != nil {
			log.Println(err)
			continue
		}

		batch[i].Consumption = cache.Update(&batch[i])

		i++
		if i < n {
			continue
		}

		if err := service.Insert(batch[0:i]); err != nil {
			log.Println(err)
		}

		if err := WriteTo(&cache, *cacheFile); err != nil {
			log.Println(err)
		}

		i = 0
	}

	if i > 0 {
		if err := service.Insert(batch[0:i]); err != nil {
			log.Println(err)
		}

		if err := WriteTo(&cache, *cacheFile); err != nil {
			log.Println(err)
		}
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}
