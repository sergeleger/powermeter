package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/sergeleger/powermeter/power"
)

func main() {
	var (
		cacheFile = flag.String("cache", "cache.gob", "cache file")
		host      = flag.String("host", "http://pi4:8086/", "InfluxDB address")
		db        = flag.String("db", "mydb", "InfluxDB database name")
		batchSize = flag.Int("batch", 5, "transaction batch size, small values for live updating")
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
	i := 0
	timer := time.NewTicker(10 * time.Minute)
	batch := make([]power.Usage, *batchSize)
	for sc.Scan() {
		select {
		case <-timer.C:
			if err := WriteTo(&cache, *cacheFile); err != nil {
				log.Println(err)
			}

		default:
		}

		if err := json.Unmarshal(sc.Bytes(), &batch[i]); err != nil {
			log.Println(err)
			continue
		}

		batch[i].Consumption = cache.Update(&batch[i])

		i++
		if i < *batchSize {
			continue
		}

		if err := service.Insert(batch[0:i]); err != nil {
			log.Println(err)
		}
		i = 0
	}
	timer.Stop()

	// attempt to update remaining batch
	if i > 0 {
		if err := service.Insert(batch[0:i]); err != nil {
			log.Println(err)
		}
	}

	if err := WriteTo(&cache, *cacheFile); err != nil {
		log.Println(err)
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}
