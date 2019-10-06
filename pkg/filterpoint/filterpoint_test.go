package filterpoint

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/kapacitor/udf/agent"
)

func TestParseTimeZone(t *testing.T) {
	pnt := getKapacitorPoint()

	for _, tc := range [...]struct {
		timezone string
		pntKap   *agent.Point
		expected string
	}{
		{"Pacific/Auckland", pnt, "Pacific/Auckland"},
		{"{timezone}", pnt, "Pacific/Auckland"},
		{"{NotExisting}", pnt, ""},
		{"", pnt, ""},
		{"{}", pnt, "{}"},
	} {
		t.Run(fmt.Sprintf("Test Parse Time Zone"), func(t *testing.T) {
			actual := parseTimeZone(tc.timezone, tc.pntKap)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func TestConverTimeToTimeZone(t *testing.T) {
	dtNZ, _ := time.Parse(time.RFC3339, "2019-08-26T15:15:15+12:00")
	dtUTC, _ := time.Parse(time.RFC3339, "2019-08-26T03:15:15Z")

	dtDaylightSavingNZ, _ := time.Parse(time.RFC3339, "2019-10-26T15:15:15+13:00")
	dtUTC1, _ := time.Parse(time.RFC3339, "2019-10-26T02:15:15Z")

	for _, tc := range [...]struct {
		dt       time.Time
		timezone string
		expected time.Time
	}{
		{dtUTC, "", dtUTC},
		{dtNZ, "NotExisting", dtNZ},
		{dtUTC, "Pacific/Auckland", dtNZ},
		{dtUTC1, "Pacific/Auckland", dtDaylightSavingNZ},
	} {
		t.Run(fmt.Sprintf("Test Parse Time Zone"), func(t *testing.T) {
			actual := converTimeToTimeZone(&tc.dt, tc.timezone)
			if actual.Sub(tc.expected) != 0 {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func getKapacitorPoint() *agent.Point {
	return &agent.Point{
		FieldsInt: map[string]int64{
			"fieldIntPos": 3,
			"fieldIntNeg": -22,
		},
		FieldsDouble: map[string]float64{
			"fieldFloatPos":      0.123,
			"fieldFloatPosRound": 0.126,
			"fieldFloatNeg":      -0.123,
			"fieldFloatNegRound": -0.126,
		},
		FieldsString: map[string]string{
			"timezone": "Pacific/Auckland",
		},
		FieldsBool: map[string]bool{
			"fieldBoolTrue":  true,
			"fieldBoolFalse": false,
		},
		Tags: map[string]string{
			"tag": "tagValue",
		},
	}
}
