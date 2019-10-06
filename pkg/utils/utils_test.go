package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/influxdata/kapacitor/udf/agent"
)

func TestStringifyPointByKey(t *testing.T) {
	pnt := getKapacitorPoint()

	for _, tc := range [...]struct {
		key       string
		point     *agent.Point
		exptected string
	}{
		{"fieldIntPos", pnt, "3"},
		{"fieldIntNeg", pnt, "-22"},
		{"fieldFloatPos", pnt, "0.12"},
		{"fieldFloatPosRound", pnt, "0.13"},
		{"fieldFloatNeg", pnt, "-0.12"},
		{"fieldFloatNegRound", pnt, "-0.13"},
		{"fieldStr", pnt, "good"},
		{"fieldBoolTrue", pnt, "true"},
		{"fieldBoolFalse", pnt, "false"},
		{"notExisting", pnt, ""},
		{"tag", pnt, "tagValue"},
	} {
		t.Run(fmt.Sprintf("Interpolate string with positive integer"), func(t *testing.T) {

			actual := StringifyPointByKey(tc.key, tc.point)
			if !reflect.DeepEqual(tc.exptected, actual) {
				t.Errorf("expected %v, actual %v", tc.exptected, actual)
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
			"fieldStr": "good",
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
