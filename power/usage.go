package power

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

// Measurement reflects the cummulative power usage at time Usage.Time
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
func (usage *Measurement) MarshalJSON() (buf []byte, err error) {
	var scm scmEncoding
	scm.Time.Time = usage.Time
	scm.Message.ID = usage.MeterID
	scm.Message.Consumption = usage.Consumption
	return json.Marshal(&scm)
}

// Measurements a list of Measurements
type Measurements []*Measurement

// ReadFrom reads measurements from the provided reader.
func (m *Measurements) ReadFrom(r io.Reader) (err error) {
	bufR := bufio.NewReader(r)
	var buf []byte

	for err == nil {
		buf, err = bufR.ReadBytes('\n')
		if err != nil {
			break
		}

		var measurement Measurement
		err = json.Unmarshal(buf, &measurement)
		if err != nil {
			log.Println(err, buf)
			err = nil
			continue
		}

		*m = append(*m, &measurement)
	}

	if err == io.EOF {
		err = nil
	}
	if err != nil {
		log.Println(err)
	}

	return err
}

// WriteTo writes the measurements to the writer.
func (m *Measurements) WriteTo(w io.Writer) (err error) {
	enc := json.NewEncoder(w)
	for _, m := range *m {
		if err = enc.Encode(&m); err != nil {
			return err
		}
	}

	return nil
}

// Less implements the sort.Interface and sorts entries by time.
func (m *Measurements) Less(i, j int) bool {
	return (*m)[i].Time.Before((*m)[j].Time)
}

// Swap implements the sort.Interface
func (m *Measurements) Swap(i, j int) { (*m)[i], (*m)[j] = (*m)[j], (*m)[i] }

// Len implements the sort.Interface
func (m *Measurements) Len() int { return len(*m) }

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
