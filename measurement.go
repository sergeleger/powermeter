package powermeter

import (
	"encoding/json"
	"fmt"
	"time"
)

// Measurement reflects the cumulative power usage at time Usage.Time
type Measurement struct {
	Time        time.Time `db:"Time"`
	MeterID     int64     `db:"MeterID"`
	Consumption int64     `db:"Usage"`
}

// UnmarshalJSON implements json.Unmarshaler and decodes the SCM protocol.
func (usage *Measurement) UnmarshalJSON(buf []byte) (err error) {
	var scm scmEncoding
	if err = json.Unmarshal(buf, &scm); err != nil {
		return fmt.Errorf("error: could not decode SCM: %w", err)
	}

	usage.Time = scm.Time.Time
	usage.MeterID = scm.Message.ID
	usage.Consumption = scm.Message.Consumption
	return nil
}

// MarshalJSON implements json.Unmarshaler and encodes the SCM protocol.
func (usage Measurement) MarshalJSON() (buf []byte, err error) {
	var scm scmEncoding
	scm.Time.Time = usage.Time
	scm.Message.ID = usage.MeterID
	scm.Message.Consumption = usage.Consumption
	return json.Marshal(&scm)
}

func (usage Measurement) Compare(other Measurement) int {
	return usage.Time.Compare(other.Time)
}

type scmEncoding struct {
	Time    Time `json:"Time"`
	Message struct {
		ID          int64 `json:"ID"`
		Consumption int64 `json:"Consumption"`
	} `json:"Message"`
}

const scmTimeFmt = `"2006-01-02T15:04:05.999999999-07:00"`

type Time struct {
	time.Time
}

func (t *Time) MarshalJSON() (buf []byte, err error) {
	str := t.Time.Format(scmTimeFmt)
	return []byte(str), nil
}

func (t *Time) UnmarshalJSON(buf []byte) (err error) {
	t.Time, err = time.Parse(scmTimeFmt, string(buf))
	return err
}
