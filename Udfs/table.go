package table

import "strings"

// Table to hold CSV-formatted table
type Table struct {
	colNames []string
	colTypes map[string]string

	bodyRows [][]interface{}
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
