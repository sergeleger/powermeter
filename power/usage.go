package power

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"sort"
	"time"

	"github.com/pkg/errors"
)

// Usage reflects the cummulative power usage at time Usage.Time
type Usage struct {
	Time        time.Time `db:"Time"`
	MeterID     int       `db:"MeterID"`
	Consumption float64   `db:"Usage"`
}

// Read reads all power reading from the reader.
func Read(dest []*Usage, r io.Reader) ([]*Usage, error) {
	if dest == nil {
		dest = make([]*Usage, 0)
	}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var usage Usage
		if err := usage.UnmarshalJSON(sc.Bytes()); err != nil {
			log.Printf("error: decoding entry: %v", err)
			continue
		}

		dest = append(dest, &usage)
	}

	return dest, errors.Wrap(sc.Err(), "error occured while reading")
}

// UnmarshalJSON implements json.Unmarshaler and decodes the SCM protocol.
func (usage *Usage) UnmarshalJSON(buf []byte) (err error) {
	var scmEncoding struct {
		Time    string `json:"Time"`
		Message struct {
			ID          int     `json:"ID"`
			Consumption float64 `json:"Consumption"`
		} `json:"Message"`
	}

	if err = json.Unmarshal(buf, &scmEncoding); err != nil {
		return errors.Wrap(err, " could not decode SVM")
	}

	if usage.Time, err = time.Parse("2006-01-02T15:04:05.999999999-07:00", scmEncoding.Time); err != nil {
		return errors.Wrap(err, "could not decode SVM time")
	}

	usage.MeterID = scmEncoding.Message.ID
	usage.Consumption = scmEncoding.Message.Consumption
	return nil
}

// Sort sorts the entries by MeterID/Time
func Sort(usage []*Usage) {
	sort.Slice(usage, func(i, j int) bool {
		if usage[i].MeterID == usage[j].MeterID {
			return usage[i].Time.Before(usage[j].Time)
		}

		return usage[i].MeterID < usage[j].MeterID
	})
}
