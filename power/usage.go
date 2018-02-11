package power

import (
	"encoding/json"
	"fmt"
	"time"
)

// Usage reflects the cummulative power usage at time Usage.Time
type Usage struct {
	Time        time.Time
	MeterID     int
	Consumption float64
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
