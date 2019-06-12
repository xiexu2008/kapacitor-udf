package table

import (
	"fmt"
	"reflect"
	"testing"
)

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
