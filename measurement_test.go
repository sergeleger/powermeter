package powermeter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestMeasurement_MarshalJSON(t *testing.T) {
	is := is.New(t)

	var usage = Measurement{
		Time:        time.Date(2018, 1, 6, 14, 44, 17, 51582803, time.Local),
		MeterID:     18553251,
		Consumption: 7652845,
	}

	expected := []byte(`{"Time":"2018-01-06T14:44:17.051582803-04:00","Message":{"ID":18553251,"Consumption":7652845}}`)

	got, err := usage.MarshalJSON()
	is.NoErr(err)
	is.Equal(got, expected)
}

func TestMeasurement_UnmarshalJSON(t *testing.T) {
	is := is.New(t)

	var usage Measurement
	var buf = []byte(`{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}`)

	err := json.Unmarshal(buf, &usage)
	is.NoErr(err)
	is.Equal(usage.Consumption, int64(7652845))
	is.Equal(usage.MeterID, int64(18553251))
	is.Equal(usage.Time, time.Date(2018, 1, 6, 14, 44, 17, 51582803, time.Local))
}
