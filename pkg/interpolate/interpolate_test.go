package interpolate

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/influxdata/kapacitor/udf/agent"
)

func TestInterpolateString(t *testing.T) {
	pnt := getKapacitorPoint()

	for _, tc := range [...]struct {
		strToInterpolate string
		pntKapacitor     *agent.Point
		expected         string
	}{
		{"start {fieldIntPos} end", pnt, "start 3 end"},
		{"start {fieldIntNeg} end", pnt, "start -22 end"},
		{"start {fieldFloatPos} end", pnt, "start 0.12 end"},
		{"start {fieldFloatPosRound} end", pnt, "start 0.13 end"},
		{"start {fieldFloatNeg} end", pnt, "start -0.12 end"},
		{"start {fieldFloatNegRound} end", pnt, "start -0.13 end"},
		{"start {fieldStr} end", pnt, "start good end"},
		{"start {fieldBoolTrue} end", pnt, "start true end"},
		{"start {fieldBoolFalse} end", pnt, "start false end"},
		{"start {notExisting} end", pnt, "start  end"},
		{"start {tag} end", pnt, "start tagValue end"},
		{"start int field {fieldIntPos}, " +
			"float field {fieldFloatNegRound}, " +
			"string field {fieldStr}, " +
			"tag {tag} and bool field {fieldBoolTrue}",
			pnt,
			"start int field 3, " +
				"float field -0.13, " +
				"string field good, " +
				"tag tagValue and bool field true"},
	} {
		t.Run(fmt.Sprintf("Interpolate string with Kapacitor point fields and tags"), func(t *testing.T) {
			actual, _ := interplolateString(tc.strToInterpolate, tc.pntKapacitor)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func TestGetValueByKey(t *testing.T) {
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

			actual := getValueByKey(tc.key, tc.point)
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
