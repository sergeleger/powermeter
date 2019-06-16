package power

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Usage reflects the cummulative power usage at time Usage.Time
type Usage struct {
	Time        time.Time
	MeterID     int
	Consumption float64
}

// Read reads all power reading from the reader.
func Read(r io.Reader) ([]*Usage, error) {
	results := make([]*Usage, 0)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var usage Usage
		if err := usage.UnmarshalJSON(sc.Bytes()); err != nil {
			return nil, err
		}

		results = append(results, &usage)
	}

	return results, nil
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
		return fmt.Errorf("error: could not decode SVM: %v", err)
	}

	if usage.Time, err = time.Parse("2006-01-02T15:04:05.999999999-07:00", scmEncoding.Time); err != nil {
		return fmt.Errorf("error: could not decode SVM time: %v", err)
	}

	usage.MeterID = scmEncoding.Message.ID
	usage.Consumption = scmEncoding.Message.Consumption
	return nil
}

// MarshalJSON implementes json.Marshaler
func (usage *Usage) MarshalJSON() ([]byte, error) {
	var jsonEncoding struct {
		Time        int64   `json:"time"`
		MeterID     int     `json:"meter"`
		Consumption float64 `json:"consumption"`
	}

	jsonEncoding.Time = usage.Time.Unix()
	jsonEncoding.MeterID = usage.MeterID
	jsonEncoding.Consumption = usage.Consumption
	return json.Marshal(jsonEncoding)
}
