package influxdb

// import (
// 	"strconv"
// 	"time"

// 	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
// 	"github.com/sergeleger/powermeter/power"
// )

// // InfluxDBService wraps the influxdb2 client to store measurements.
// type InfluxDBService struct {
// 	client influxdb2.Client
// 	org    string
// 	bucket string
// }

// // NewInfluxDBService creates a connection to a SQLite database.
// func NewInfluxDBService(connection, token, org, bucket string) (*InfluxDBService, error) {
// 	client := influxdb2.NewClientWithOptions(
// 		connection,
// 		token,
// 		influxdb2.DefaultOptions().SetPrecision(time.Millisecond),
// 	)

// 	return &InfluxDBService{
// 		client: client,
// 		bucket: bucket,
// 		org:    org,
// 	}, nil
// }

// // Close releases all resources.
// func (s *InfluxDBService) Close() error {
// 	s.client.Close()
// 	return nil
// }

// // Insert adds new entries to the table
// func (s *InfluxDBService) Insert(usage []power.Measurement) error {
// 	writer := s.client.WriteAPI(s.org, s.bucket)

// 	// // capture the first error and use it as the return value.
// 	// errCh := writer.Errors()
// 	// var err error
// 	// var wg sync.WaitGroup
// 	// wg.Add(1)
// 	// go func() {
// 	// 	for e := range errCh {
// 	// 		if err == nil {
// 	// 			err = e
// 	// 		}
// 	// 	}
// 	// 	log.Println("DONE!")
// 	// 	wg.Done()
// 	// }()

// 	// Write measurements.
// 	for _, u := range usage {
// 		if u.Consumption == 0 {
// 			continue
// 		}

// 		p := influxdb2.NewPointWithMeasurement("power").
// 			AddTag("meter", strconv.Itoa(u.MeterID)).
// 			AddTag("year", u.Time.Format("2006")).
// 			AddTag("month", u.Time.Format("01")).
// 			AddField("consumption", u.Consumption).
// 			SetTime(u.Time.UTC())

// 		writer.WritePoint(p)
// 	}

// 	writer.Flush()
// 	//	wg.Wait()
// 	return nil
// }
