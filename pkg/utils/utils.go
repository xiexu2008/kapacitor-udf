package utils

import (
	"strconv"

	"github.com/influxdata/kapacitor/udf/agent"
)

// StringifyPointByKey is to read the value of
// tag or field specified by the passed-in key
// and return the value as a string.
func StringifyPointByKey(key string, p *agent.Point) string {
	if val, ok := p.Tags[key]; ok {
		return val
	}
	if val, ok := p.FieldsString[key]; ok {
		return val
	}
	if val, ok := p.FieldsInt[key]; ok {
		return strconv.FormatInt(val, 10)
	}
	if val, ok := p.FieldsDouble[key]; ok {
		return strconv.FormatFloat(val, 'f', 2, 64)
	}
	if val, ok := p.FieldsBool[key]; ok {
		return strconv.FormatBool(val)
	}

	return ""
}
