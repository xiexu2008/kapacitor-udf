package interpolate

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/influxdata/kapacitor/udf/agent"
)

func interplolateString(str string, p *agent.Point) (string, error) {
	// To interpolate string like "Lower {lowerThresh} upper {upperThresh} within {withinSec}s"
	// with the fields or tags defined in Kapacitor pointer

	var sb strings.Builder
	var keyName strings.Builder
	isKeyName := false
	for _, ch := range str {
		if ch == '{' {
			isKeyName = true
			keyName.Reset()
		} else if ch == '}' {
			key := keyName.String()
			val := getValueByKey(key, p)
			sb.WriteString(val)

			isKeyName = false
			keyName.Reset()
		} else {
			if isKeyName {
				if !unicode.IsSpace(ch) {
					keyName.WriteRune(ch)
				}
			} else {
				sb.WriteRune(ch)
			}
		}
	}

	return sb.String(), nil
}

func getValueByKey(key string, p *agent.Point) string {
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
