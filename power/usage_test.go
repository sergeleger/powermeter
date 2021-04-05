package power

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestReadFrom(t *testing.T) {
	is := is.New(t)

	// Read the data (in original SCM format.)
	var measurements Measurements
	err := measurements.ReadFrom(bytes.NewReader(scmTestData))
	is.NoErr(err)
	is.Equal(len(measurements), 3)

	// Write it back
	var buf bytes.Buffer
	err = measurements.WriteTo(&buf)
	is.NoErr(err)

	// Read it again in modified SCM format.
	var got Measurements
	err = got.ReadFrom(&buf)
	is.NoErr(err)
	is.Equal(got, measurements)
}

func TestSCMUnmarshalJSON(t *testing.T) {
	is := is.New(t)

	var usage Measurement
	var buf = []byte(`{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}`)

	err := json.Unmarshal(buf, &usage)
	is.NoErr(err)
	is.Equal(usage.Consumption, float64(7652845))
	is.Equal(usage.MeterID, 18553251)
	is.Equal(usage.Time, time.Date(2018, 1, 6, 14, 44, 17, 51582803, time.Local))
}

var scmTestData = []byte(
	`{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}
{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}
{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}
`)
