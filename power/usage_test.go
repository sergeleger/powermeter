package power

import (
	"testing"
	"time"
)

func TestSCMUnmarshalJSON(t *testing.T) {
	var usage Usage
	var err = usage.UnmarshalJSON([]byte(`{"Time":"2018-01-06T14:44:17.051582803-04:00","Offset":0,"Length":0,"Message":{"ID":18553251,"Type":7,"TamperPhy":2,"TamperEnc":1,"Consumption":7652845,"ChecksumVal":59811}}`))

	if err != nil {
		t.Fatalf("error while decoding: %v", err)
	}

	if usage.Consumption != 7652845 {
		t.Fatal("incorrect Consumption.")
	}

	if usage.MeterID != 18553251 {
		t.Fatal("incorrect ID.")
	}

	if usage.Time != time.Date(2018, 1, 6, 14, 44, 17, 51582803, time.Local) {
		t.Fatal("incorrect Time.")
	}
}
