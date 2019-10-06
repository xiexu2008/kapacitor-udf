package matchtime

import (
	"fmt"
	"testing"
	"time"
)

func TestMatchTimeWithMask(t *testing.T) {
	dt, _ := time.Parse(time.RFC3339, "2019-08-26T15:15:15Z")
	sunday, _ := time.Parse(time.RFC3339, "2019-08-25T11:30:00Z")
	for _, tc := range [...]struct {
		mask     string
		dt       *time.Time
		expected bool
	}{
		{"Y==2019", &dt, true},
		{"Y>=2019", &dt, true},
		{"Y<=2019", &dt, true},
		{"Y>2019", &dt, false},
		{"Y<2019", &dt, false},
		{"Y!=2019", &dt, false},
		{"M==8", &dt, true},
		{"M>=8", &dt, true},
		{"M<=8", &dt, true},
		{"M>8", &dt, false},
		{"M<8", &dt, false},
		{"M!=8", &dt, false},
		{"D==26", &dt, true},
		{"D>=26", &dt, true},
		{"D<=26", &dt, true},
		{"D>26", &dt, false},
		{"D<26", &dt, false},
		{"D!=26", &dt, false},
		{"h==15", &dt, true},
		{"h>=15", &dt, true},
		{"h<=15", &dt, true},
		{"h>15", &dt, false},
		{"h<15", &dt, false},
		{"h!=15", &dt, false},
		{"m==15", &dt, true},
		{"m>=15", &dt, true},
		{"m<=15", &dt, true},
		{"m>15", &dt, false},
		{"m<15", &dt, false},
		{"m!=15", &dt, false},
		{"s==15", &dt, true},
		{"s>=15", &dt, true},
		{"s<=15", &dt, true},
		{"s>15", &dt, false},
		{"s<15", &dt, false},
		{"s!=15", &dt, false},
		{"W==1", &dt, true},
		{"W>=1", &dt, true},
		{"W<=1", &dt, true},
		{"W>1", &dt, false},
		{"W<1", &dt, false},
		{"W!=1", &dt, false},

		{"W>=1 & W<=5 & h >= 9 & h <= 18", &dt, true},      // weekdays 7am to 7pm
		{"W>=1 & W<=5 & h >= 9 & h <= 18", &sunday, false}, // weekdays 7am to 7pm
		{"W>=1 & W<=5 & h==15 & m==15", &sunday, false},    // weekdays 3:15pm
		{"W>=1 & W<=5 & h==15 & m==15", &dt, true},         // weekdays 3:15pm
		{"W>=1 & W<=5 | (h==15 | h==11)", &sunday, true},
		{"W==0 & w==1 | (h==15 | h==11)", &dt, true},
		{"(W==0 & w==1) | (h==15 | h==11)", &dt, true},
		{"((W==0 & (w==1)) | (h==15 | h==11))", &dt, true},
	} {
		t.Run("Match time with mask", func(t *testing.T) {
			actual := MatchTimeWithMask(tc.mask, tc.dt)
			if actual != tc.expected {
				t.Errorf("Input %v, expected %v, actual %v", tc.mask, tc.expected, actual)
			}
		})
	}
}

func TestPopStackTillOpeningBracket(t *testing.T) {
	bracket := stack{
		s: []string{
			"(",
			"1",
			"+",
			"0",
			"+",
			"0",
		},
	}

	noBracket := stack{
		s: []string{
			"1",
			"+",
			"0",
			"+",
			"0",
		},
	}
	for _, tc := range [...]struct {
		s        *stack
		expected string
	}{
		{&bracket, "0+0+1"},
		{&noBracket, "0+0+1"},
	} {
		t.Run(fmt.Sprintf("Oop stack till bracket"), func(t *testing.T) {
			actual := popStackTillOpeningBracket(tc.s)
			if actual != tc.expected {
				t.Errorf("Input %v, expected %v, actual %v", tc.s, tc.expected, actual)
			}
		})

	}
}

func TestEvaluateArithmeticExpression(t *testing.T) {
	for _, tc := range [...]struct {
		expression string
		expected   int
	}{
		{"1+1+1+0+0+1", 4},
		{"1+1*1*1+1*0*1+0*1", 2},
		{"1+ 1*1* 1 + 0*1 + 1*1*1", 3},
	} {
		t.Run(fmt.Sprintf("Evaluate arithmetic expression"), func(t *testing.T) {
			actual := evaluateArithmeticExpression(tc.expression)
			if actual != tc.expected {
				t.Errorf("Input %v, expected %v, actual %v", tc.expression, tc.expected, actual)
			}
		})
	}
}

func TestEvaluateSingleBooleanExpression(t *testing.T) {
	dt, _ := time.Parse(time.RFC3339, "2019-08-26T15:15:15Z")
	for _, tc := range [...]struct {
		expression string
		dt         *time.Time
		expected   bool
	}{
		{"Y > 2018", &dt, true},
		{"Y <= 2018", &dt, false},
		{"M > 6", &dt, true},
		{"M < 10", &dt, true},
		{"D> 29", &dt, false},
		{"D > 20", &dt, true},
		{"h ==12", &dt, false},
		{"m > 30", &dt, false},
		{"s > 30", &dt, false},
		{"s ==15", &dt, true},
		{"W==1", &dt, true},
		{"W>1", &dt, false},
	} {
		t.Run(fmt.Sprintf("Evaluate single boolean expression"), func(t *testing.T) {
			actual := evaluateSingleBooleanExpression(tc.expression, tc.dt)
			if actual != tc.expected {
				t.Errorf("Input %v, expected %v, actual %v", tc.expression, tc.expected, actual)
			}
		})
	}
}

func TestMatchTime(t *testing.T) {
	dt, _ := time.Parse(time.RFC3339, "2019-08-26T15:15:15Z")
	for _, tc := range [...]struct {
		left     string
		operator string
		right    string
		dt       *time.Time
		expected bool
	}{
		{"Y", ">", "2018", &dt, true},
		{"Y", "<=", "2018", &dt, false},
		{"M", ">", "6", &dt, true},
		{"M", "<", "10", &dt, true},
		{"D", ">", "29", &dt, false},
		{"D", ">", "20", &dt, true},
		{"h", "==", "12", &dt, false},
		{"m", ">", "30", &dt, false},
		{"s", ">", "30", &dt, false},
		{"s", "==", "15", &dt, true},
	} {
		t.Run(fmt.Sprintf("Match time"), func(t *testing.T) {
			actual := matchTime(tc.left, tc.operator, tc.right, tc.dt)
			if actual != tc.expected {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func TestDoComparison(t *testing.T) {
	for _, tc := range [...]struct {
		leftOprd  int
		operator  string
		rightOprd int
		expected  bool
	}{
		{2, "==", 3, false},
		{2, "!=", 3, true},
		{2, ">=", 3, false},
		{2, "<=", 3, true},
		{2, ">", 3, false},
		{3, "==", 3, true},
		{3, "!=", 3, false},
		{3, ">=", 3, true},
		{3, "<=", 3, true},
		{3, ">", 3, false},
		{3, "<", 3, false},
	} {
		t.Run(fmt.Sprintf("Do comparison for %v %s %v", tc.leftOprd, tc.operator, tc.rightOprd), func(t *testing.T) {
			actual := doComparison(tc.leftOprd, tc.operator, tc.rightOprd)
			if actual != tc.expected {
				t.Errorf("expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}
