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

// LoadFromCsvString to load CSV-formatted table string.
// The table should follow the pattern:
//   `
//   nameCol1,nameCol2,nameCol3
//   typeCol1,typeCol2,typeCol3
//   valCol11,valCol21,valCol31
//   `
//
// For example:
//   `
//   client,region,domestic,total,compliance
//   string,string,bool,int,float
//   TT,APAC,false,100,0.9
//   ...
//   `
func (tbl *Table) LoadFromCsvString(tableStr string) error {
	// TODO: validate the input and error handling

	tableStr = strings.TrimSpace(tableStr)
	rows := strings.Split(tableStr, "\n")

	// Load column names from the first row
	tbl.colNames = SplitAndTrimSpace(rows[0], ",")

	// Load column types from the second row
	tbl.loadColumnTypes(rows[1])

	// Load all data (body) rows
	tbl.loadBodyRows(rows[2:])

	return nil
}

func (tbl *Table) loadColumnTypes(typesStr string) {
	colTypes := SplitAndTrimSpace(typesStr, ",")

	tbl.colTypes = make(map[string]string)
	for i, col := range tbl.colNames {
		tbl.colTypes[col] = colTypes[i]
	}
}

func (tbl *Table) loadBodyRows(bodyRows []string) {
	tbl.bodyRows = make([][]interface{}, len(bodyRows))

	for i, row := range bodyRows {
		tbl.bodyRows[i] = tbl.loadSingleRow(row)
	}
}

func (tbl *Table) loadSingleRow(row string) []interface{} {
	fieldsInStr := SplitAndTrimSpace(row, ",")
	fields := make([]interface{}, len(fieldsInStr))

	for i, valStr := range fieldsInStr {
		valType := tbl.colTypes[tbl.colNames[i]]
		fields[i], _ = ConvertStringToType(valStr, valType)
	}

	return fields
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
