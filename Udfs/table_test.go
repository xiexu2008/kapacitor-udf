package table

import (
	"fmt"
	"reflect"
	"testing"
)

func getTable() *Table {
	tableCsv := `client,region,domestic,total,compliance
	string,string,bool,int,float
	TT,APAC,false,100,0.9
	ATPIUK,EU,true,99,0.8
	TT,APAC,true,66,0.95	`

	t := Table{}
	t.LoadFromCsvString(tableCsv)

	return &t
}

func TestLoadFromCsvString(t *testing.T) {
	// TODO: add negative test cases

	// test cases
	for _, tc := range [...]struct {
		tableCsvStr string
		table       *Table
		expected    *Table
	}{
		{"client,total,compliance,domestic\nstring,int,float,bool\nTT,100,0.9,true\nFCL,200,0.95,false",
			&Table{},
			&Table{
				[]string{"client", "total", "compliance", "domestic"},
				map[string]string{"client": "string", "total": "int", "compliance": "float", "domestic": "bool"},
				[][]interface{}{{"TT", int64(100), 0.9, true}, {"FCL", int64(200), 0.95, false}},
			},
		},
		{`client,total, compliance,domestic
		string,int,float, bool
		TT,100, 0.9,true
		FCL,200,0.95, false`,
			&Table{},
			&Table{
				[]string{"client", "total", "compliance", "domestic"},
				map[string]string{"client": "string", "total": "int", "compliance": "float", "domestic": "bool"},
				[][]interface{}{{"TT", int64(100), 0.9, true}, {"FCL", int64(200), 0.95, false}},
			},
		},
	} {
		t.Run(fmt.Sprintf("Load table from CSV-formatted string %s", tc.tableCsvStr), func(t *testing.T) {
			tc.table.LoadFromCsvString(tc.tableCsvStr)
			if !reflect.DeepEqual(tc.expected.colNames, tc.table.colNames) {
				t.Errorf("expected colNames %v, actual colNames %v", tc.expected.colNames, tc.table.colNames)
			}
			if !reflect.DeepEqual(tc.expected.colTypes, tc.table.colTypes) {
				t.Errorf("expected colTypes %v, actual colTypes %v", tc.expected.colTypes, tc.table.colTypes)
			}
			if !reflect.DeepEqual(tc.expected.bodyRows, tc.table.bodyRows) {
				t.Errorf("expected bodyRows %v, actual bodyRows %v", tc.expected.bodyRows, tc.table.bodyRows)
			}
		})
	}
}

func TestGetRowByColumns(t *testing.T) {
	// TODO: add negative test cases
	tbl := getTable()

	// test cases
	for _, tc := range [...]struct {
		query    map[string]interface{}
		table    *Table
		expected map[string]interface{}
	}{
		{map[string]interface{}{"client": "TT", "region": "APAC"},
			tbl,
			map[string]interface{}{"client": "TT", "region": "APAC", "domestic": false, "total": int64(100), "compliance": 0.9},
		},
		{map[string]interface{}{"client": "Unknown", "region": "APAC"},
			tbl,
			nil,
		},
	} {
		t.Run(fmt.Sprintf("Get table row by cloumns %v", tc.query), func(t *testing.T) {
			actual := tc.table.GetRowByColumns(tc.query)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected colNames %v, actual colNames %v", tc.expected, actual)
			}
		})
	}
}

func TestConvertStringToType(t *testing.T) {
	// test cases
	for _, tc := range [...]struct {
		value    string
		valType  string
		expected interface{}
	}{
		{"TT", "string", "TT"},
		{"100", "int", int64(100)},
		{"999.999", "float", 999.999},
		{"101.00", "float", 101.0},
		{"true", "bool", true},
		{"false", "bool", false},
		{"Unknown", "unknown", nil},
	} {
		t.Run(fmt.Sprintf("Convert string '%s' to type '%s'", tc.value, tc.valType), func(t *testing.T) {
			actual, _ := ConvertStringToType(tc.value, tc.valType)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func TestSplitAndTrimSpace(t *testing.T) {
	type checkFunc func([]string) error

	// checkers
	isEmptyList := func(have []string) error {
		if len(have) > 0 {
			return fmt.Errorf("expected empty list, found %v", have)
		}
		return nil
	}
	is := func(want ...string) checkFunc {
		return func(have []string) error {
			if !reflect.DeepEqual(have, want) {
				return fmt.Errorf("expected list %v, found %v", want, have)
			}
			return nil
		}
	}

	// test cases
	for _, tc := range [...]struct {
		strToSplit string
		splitBy    string
		check      checkFunc
	}{
		{"", ",", isEmptyList},
		{"client,domain,slo", ",", is("client", "domain", "slo")},
		{"client,domain,slo", ":", is("client,domain,slo")},
		{" T T : 123: 11.99:11.0 :false ", ":", is("T T", "123", "11.99", "11.0", "false")},
		{"\tT T , 123,11.99,11.0 ,false\r\n,", ",", is("T T", "123", "11.99", "11.0", "false", "")},
		{`client,
		domain,
		slo`, ",", is("client", "domain", "slo")},
	} {
		t.Run(fmt.Sprintf("Split %s by '%s' and trim space", tc.strToSplit, tc.splitBy), func(t *testing.T) {
			result := SplitAndTrimSpace(tc.strToSplit, tc.splitBy)
			if err := tc.check(result); err != nil {
				t.Error(err)
			}
		})
	}
}
