package matchtime

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type stack struct {
	s    []string
	lock sync.Mutex
}

func (s *stack) Push(val string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.s = append(s.s, val)
}

func (s *stack) Pop() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.s)
	if l == 0 {
		return "", errors.New("Stack is empty")
	}

	val := s.s[l-1]
	s.s = s.s[:l-1]

	return val, nil
}

func (s *stack) ToString(seperator string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return strings.Join(s.s, seperator)
}

// MatchTimeWithMask is to match the passed-in time 'dt' with
// the passed-in mask 'timeMask'.
// The time mask like "Y >= 2019 & (M==5 | M==8) | (h > 8 & h < 6)"
// defines a set of time values. This function is to check
// if the passed-in time 'dt' is in this set.
func MatchTimeWithMask(timeMask string, dt *time.Time) bool {
	resStack := &stack{
		lock: sync.Mutex{},
		s:    make([]string, 0),
	}

	for i := 0; i < len(timeMask); i++ {
		ch := timeMask[i]
		if ch == 'Y' || ch == 'M' || ch == 'D' || ch == 'h' || ch == 'm' || ch == 's' || ch == 'W' {
			var singleExprssn string
			singleExprssn, i = extractSingleBooleanExpression(i, timeMask)

			// Evaluate boolean expression like "Y <= 2019" and save the
			// result as string 1 or 0 to back to result stack
			val := 0
			if evaluateSingleBooleanExpression(singleExprssn, dt) {
				val = 1
			}
			resStack.Push(strconv.Itoa(val))
		} else if ch == '&' || ch == '|' {
			// Replace '&' with '*' and '|' with '+'
			op := "*"
			if ch == '|' {
				op = "+"
			}
			resStack.Push(op)
		} else if ch == ')' {
			// Evalute arithmetic expression like "1+0+0*1*1+0"
			// and save result back to result stack
			arithExprssn := popStackTillOpeningBracket(resStack)
			res := evaluateArithmeticExpression(arithExprssn)
			resStack.Push(strconv.Itoa(res))
		} else if !unicode.IsSpace(rune(ch)) {
			resStack.Push(string(ch))
		}
	}

	r := evaluateArithmeticExpression(resStack.ToString(""))
	return r != 0
}

func extractSingleBooleanExpression(index int, exprssn string) (string, int) {
	// Given the passed-in 'exprssn' with value "Y >= 2019 & (M==5 | M==8) | (h > 8 & h < 6)",
	// the function extracts "Y >= 2019" and return it back with current index.

	var signleExpr strings.Builder
	j := index

	for ; j < len(exprssn); j++ {
		if exprssn[j] != '&' && exprssn[j] != '|' && exprssn[j] != ')' {
			signleExpr.WriteByte(exprssn[j])
		} else {
			break
		}
	}

	return signleExpr.String(), j - 1
}

func popStackTillOpeningBracket(s *stack) string {
	var sb strings.Builder
	ch, err := s.Pop()
	for err == nil && ch != "(" {
		sb.WriteString(ch)
		ch, err = s.Pop()
	}

	return sb.String()
}

func evaluateArithmeticExpression(arithExprssn string) int {
	// To calulate expression like "1+0*1*1+1".
	// Note it only processes single digit, which means
	// expression like "1+22*1+0*1" will fail.

	var s stack

	// Process multiply like "0*1*1*0"
	for i := 0; i < len(arithExprssn); i++ {
		ch := arithExprssn[i]
		if ch == '*' {
			previous, _ := s.Pop()
			v, _ := strconv.Atoi(previous)
			for i++; i < len(arithExprssn); i++ {
				if !unicode.IsSpace(rune(arithExprssn[i])) {
					break
				}
			}

			c, _ := strconv.Atoi(string(arithExprssn[i]))
			r := v * c
			s.Push(strconv.Itoa(r))
		} else if !unicode.IsSpace(rune(ch)) {
			s.Push(string(ch))
		}
	}

	// Process addition like "1+0+1+0"
	r := 0
	v, e := s.Pop()
	for e == nil {
		if v != "+" {
			i, _ := strconv.Atoi(v)
			r += i
		}
		v, e = s.Pop()
	}

	return r
}

func evaluateSingleBooleanExpression(exprssn string, dt *time.Time) bool {
	// Evaluate string expression like "Y <= 2019" to true or false.
	// The left operand could be "Y" (year), "M" (month), "D" (day),
	// "h" (hour), "m" (minute), "s" (second) or "W" (weekday).
	// It uses the related value (year, month, ..., second) of the passed-in
	// time 'dt' to do the evaluate.

	re := regexp.MustCompile(`([Y,M,D,h,m,s,W]{1})\s*([!,=,>,<]+)\s*(\d+)`)
	match := re.FindStringSubmatch(exprssn)

	return matchTime(match[1], match[2], match[3], dt)
}

func matchTime(YMDhms string, operator string, rightOperand string, dt *time.Time) bool {
	var val = GetTimeField(YMDhms, dt)

	valInt, _ := strconv.Atoi(rightOperand)
	return doComparison(val, operator, valInt)
}

// GetTimeField is to get the specified filed of date and time.
func GetTimeField(fieldName string, dt *time.Time) int {
	var val = -1

	switch fieldName {
	case "Y": // year
		val = dt.Year()
	case "M": // month
		val = int(dt.Month())
	case "D": // day
		val = dt.Day()
	case "h": // hour
		val = dt.Hour()
	case "m": // minute
		val = dt.Minute()
	case "s": // second
		val = dt.Second()
	case "W": // weekday
		val = int(dt.Weekday())
	}

	return val
}

func doComparison(leftOperand int, operator string, rightOperand int) bool {
	// Calculate the bool value of expression like: 2 "<=" 3, in which the operator
	// is passed in as a string

	switch operator {
	case "==":
		return leftOperand == rightOperand
	case "!=":
		return leftOperand != rightOperand
	case ">=":
		return leftOperand >= rightOperand
	case "<=":
		return leftOperand <= rightOperand
	case ">":
		return leftOperand > rightOperand
	case "<":
		return leftOperand < rightOperand
	}

	return false
}
