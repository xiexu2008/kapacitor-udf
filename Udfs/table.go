package table

import (
	"strconv"
	"strings"
)

// Table to hold CSV-formatted table
type Table struct {
	colNames []string
	colTypes map[string]string

	bodyRows [][]interface{}
}

func ConvertStringToType(value string, valType string) (interface{}, error) {
	var result interface{}
	var err error

	switch valType {
	case "string":
		result = value
	case "int":
		tempVal, tempErr := strconv.Atoi(value)
		err = tempErr
		result = int64(tempVal)
	case "float":
		result, err = strconv.ParseFloat(value, 64)
	case "bool":
		result, err = strconv.ParseBool(value)
	}

	return result, err
}

func SplitAndTrimSpace(rawStr string, splitBy string) []string {
	rawStr = strings.TrimSpace(rawStr)
	if len(rawStr) == 0 {
		return []string{}
	}

	items := strings.Split(rawStr, splitBy)

	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}

	return items
}
