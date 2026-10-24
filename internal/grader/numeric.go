package grader

import (
	"math"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

var (
	plusZeroRe   = regexp.MustCompile(`^\+`)
	andRe        = regexp.MustCompile(`(?i)(\d)\s*and\s*(\d)`)
	mixedRe      = regexp.MustCompile(`^(-?\d+)\s+(\d+/\d+)$`)
	fractionRe   = regexp.MustCompile(`^(-?\d+)/(\d+)$`)
	leadingDotRe = regexp.MustCompile(`^(-?)\.`)
	sciRe        = regexp.MustCompile(`^(-?\d+\.?\d*)[eE](-?\d+)$`)
)

const floatTolerance = 1e-9

type numericGrader struct{}

func (g *numericGrader) grade(expected, answer string) Result {
	e := normalise(expected)
	a := normalise(answer)
	if e == "" || a == "" {
		return emptyResult()
	}
	if res, ok := compareScalars(e, a); ok {
		return res
	}
	if res, ok := compareLists(e, a); ok {
		return res
	}
	// A comma used as a decimal separator is how most of Europe writes 3.5.
	// It cannot be normalised up front, because `6,8` is a two-element answer
	// in this corpus and not sixty-eight. It is only tried after both other
	// readings fail, so a list is never mistaken for a number.
	if strings.Contains(a, ",") && !strings.Contains(e, ",") {
		if res, ok := compareScalars(e, strings.ReplaceAll(a, ",", ".")); ok {
			return res
		}
	}
	// Nothing could be read as a number or a list, which is a different problem
	// from a wrong value and worth saying so.
	if _, _, eOK := parseNumber(e); !eOK {
		return Result{Correct: false, Score: 0, Feedback: "Could not parse numeric values"}
	}
	if _, _, aOK := parseNumber(a); !aOK {
		return Result{Correct: false, Score: 0, Feedback: "Could not parse numeric values"}
	}
	return Result{Correct: false, Score: 0, Feedback: "Incorrect"}
}

func compareScalars(expected, answer string) (Result, bool) {
	eVal, eIsFloat, eOK := parseNumber(expected)
	aVal, aIsFloat, aOK := parseNumber(answer)
	if !aOK {
		// A learner who works the last step out in their head sometimes types the
		// working: "3-4" where the answer is -1. Reading it as a number is the
		// difference between grading the answer and grading its presentation.
		if v, ok := evalExpression(answer); ok {
			aVal, aIsFloat, aOK = v, true, true
		}
	}
	if !eOK || !aOK {
		return Result{}, false
	}
	if eIsFloat || aIsFloat {
		if math.Abs(eVal-aVal) <= numericEpsilon(expected, eVal) {
			return Result{Correct: true, Score: 1}, true
		}
		return Result{Correct: false, Score: 0, Feedback: "Incorrect"}, true
	}
	if eVal == aVal {
		return Result{Correct: true, Score: 1}, true
	}
	return Result{Correct: false, Score: 0, Feedback: "Incorrect"}, true
}

// compareLists grades an answer that names several values, such as the roots of
// a quadratic, the elements of a set or the terms of an expanded number. Order
// does not matter: a learner who writes the roots of x^2 + x - 12 = 0 as 4,-3
// has the same answer as -3,4, and three reports failed on exactly that.
//
// The bool is false when either side is not a list, so the caller can try the
// next reading. A single value is not a list, which keeps `12` from matching
// `1,2`.
func compareLists(expected, answer string) (Result, bool) {
	e := splitList(expected)
	a := splitList(answer)
	if len(e) < 2 || len(a) < 2 {
		return Result{}, false
	}
	if len(e) != len(a) {
		return Result{Correct: false, Score: 0, Feedback: "Incorrect"}, true
	}
	eVals := make([]float64, 0, len(e))
	aVals := make([]float64, 0, len(a))
	for _, part := range e {
		v, _, ok := parseNumber(part)
		if !ok {
			return Result{}, false
		}
		eVals = append(eVals, v)
	}
	for _, part := range a {
		v, _, ok := parseNumber(part)
		if !ok {
			return Result{}, false
		}
		aVals = append(aVals, v)
	}
	sort.Float64s(eVals)
	sort.Float64s(aVals)
	for i := range eVals {
		if math.Abs(eVals[i]-aVals[i]) > floatTolerance*math.Max(1, math.Abs(eVals[i])) {
			return Result{Correct: false, Score: 0, Feedback: "Incorrect"}, true
		}
	}
	return Result{Correct: true, Score: 1}, true
}

// splitList reads the parts of a multi-value answer, dropping the brackets a
// learner may type around a set or a coordinate pair.
func splitList(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "()[]{}")
	var parts []string
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' }) {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) < 2 {
		parts = splitBeforeMinusAfterDot(s)
	}
	return parts
}

// splitBeforeMinusAfterDot reads the list a learner typed when the separator
// arrived as a full stop, so the roots -1, -2 reach the server as "-1.-2".
// A decimal never has a minus sign straight after its point, so a "." followed
// by "-" is a separator and `3.5` is left alone.
func splitBeforeMinusAfterDot(s string) []string {
	runes := []rune(s)
	var parts []string
	start := 0
	for i := 0; i+1 < len(runes); i++ {
		if runes[i] == '.' && runes[i+1] == '-' {
			if part := strings.TrimSpace(string(runes[start:i])); part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	if len(parts) == 0 {
		return nil
	}
	if tail := strings.TrimSpace(string(runes[start:])); tail != "" {
		parts = append(parts, tail)
	}
	return parts
}

// stripVariablePrefix removes the "x=" a learner often types in front of a
// value the question asked them to find.
func stripVariablePrefix(s string) string {
	if i := strings.IndexByte(s, '='); i > 0 {
		left := strings.TrimSpace(s[:i])
		left = strings.TrimPrefix(left, `\`)
		if len(left) >= 1 && ((left[0] >= 'a' && left[0] <= 'z') || (left[0] >= 'A' && left[0] <= 'Z')) {
			return strings.TrimSpace(s[i+1:])
		}
	}
	return s
}

// stripThousands removes the grouping commas from a number and leaves every
// other comma where it is. normalise used to delete all of them, which turned
// `6,8` into `68`, `7,10` into `710` and a decimal comma into a wrong integer.
func stripThousands(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i := 0; i < len(runes); i++ {
		if runes[i] == ',' && i+4 <= len(runes) &&
			isASCIIDigit(runes[i+1]) && isASCIIDigit(runes[i+2]) && isASCIIDigit(runes[i+3]) &&
			(i+4 == len(runes) || !isASCIIDigit(runes[i+4])) {
			continue
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

func isASCIIDigit(r rune) bool { return r >= '0' && r <= '9' }

// parseNumber reads one value, trying the grouping commas of a large number as
// a second reading. It is per value rather than per answer, because `1000,200`
// is a list whose first part is four digits, and stripping commas from the whole
// answer would join the two parts into one number.
func parseNumber(s string) (float64, bool, bool) {
	s = stripVariablePrefix(strings.TrimSpace(s))
	if v, isFloat, ok := parseNumeric(s); ok {
		return v, isFloat, true
	}
	return parseNumeric(stripThousands(s))
}

func normalise(s string) string {
	s = plusZeroRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// numericEpsilon is the window inside which an answer counts as the same number
// as the expected one. For an expected value printed with a decimal point it is
// half a unit in the last printed place: "12.61" accepts 12.610001, because
// that is the same answer rounded differently, and rejects 12.62, which is a
// different number at that precision. Expected values without a decimal point,
// integers, fractions, mixed numbers and scientific notation, keep the flat
// tolerance, because widening those would accept a different value of the same
// shape. The relative term keeps large magnitudes comparable without making the
// window unlimited.
func numericEpsilon(expected string, value float64) float64 {
	eps := floatTolerance
	if i := strings.IndexByte(expected, '.'); i >= 0 && !strings.ContainsAny(expected, "/eE") {
		if places := len(expected) - i - 1; places > 0 {
			if half := 0.5 / math.Pow(10, float64(places)); half > eps {
				eps = half
			}
		}
	}
	if relative := floatTolerance * math.Abs(value); relative > eps {
		eps = relative
	}
	return eps
}

func parseNumeric(s string) (float64, bool, bool) {
	if s == "" {
		return 0, false, false
	}

	// Accept the spoken form of a mixed number: "4 and 1/10" / "2and3/10"
	// normalize to "4 1/10" before matching. (The English reading is how
	// students naturally type it, and matches the worked explanation.)
	s = andRe.ReplaceAllString(s, "$1 $2")

	// Mixed number: "1 1/2", check before stripping whitespace. A negative
	// whole number subtracts its fraction ("-1 1/2" = -(1 + 1/2) = -1.5), not
	// adds it ("-1 1/2" must not parse as -1 + 0.5).
	if m := mixedRe.FindStringSubmatch(s); m != nil {
		whole := new(big.Rat)
		whole.SetString(m[1])
		frac := new(big.Rat)
		if _, ok := frac.SetString(m[2]); !ok {
			return 0, false, false
		}
		total := new(big.Rat)
		if whole.Sign() < 0 {
			total.Sub(whole, frac)
		} else {
			total.Add(whole, frac)
		}
		val, _ := total.Float64() // aislop-ignore-line ai-slop/swallowed-exception -- Float64's 2nd result is an exactness flag, not an error
		return val, true, true
	}

	// Strip whitespace for all other formats
	s = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, s)

	if strings.HasPrefix(s, "-.") {
		s = "-0." + s[2:]
	}
	if s == "." || strings.HasPrefix(s, ".") {
		s = "0" + s
	}

	// Scientific notation
	if m := sciRe.FindStringSubmatch(s); m != nil {
		coeff := new(big.Rat)
		if _, ok := coeff.SetString(m[1]); !ok {
			return 0, false, false
		}
		exp := new(big.Int)
		if _, ok := exp.SetString(m[2], 10); !ok {
			return 0, false, false
		}
		ten := big.NewInt(10)
		pow := new(big.Int).Exp(ten, new(big.Int).Abs(exp), nil)
		if exp.Sign() >= 0 {
			num := new(big.Int).Mul(coeff.Num(), pow)
			val, _ := new(big.Rat).SetFrac(num, coeff.Denom()).Float64() // aislop-ignore-line ai-slop/swallowed-exception -- Float64's 2nd result is an exactness flag, not an error
			return val, true, true
		}
		denom := new(big.Int).Mul(coeff.Denom(), pow)
		val, _ := new(big.Rat).SetFrac(coeff.Num(), denom).Float64() // aislop-ignore-line ai-slop/swallowed-exception -- Float64's 2nd result is an exactness flag, not an error
		return val, true, true
	}

	// Fraction
	if m := fractionRe.FindStringSubmatch(s); m != nil {
		frac := new(big.Rat)
		if _, ok := frac.SetString(m[1] + "/" + m[2]); !ok {
			return 0, false, false
		}
		val, _ := frac.Float64()
		return val, true, true
	}

	// Integer or float (including big.Rat for exact parsing)
	r := new(big.Rat)
	if _, ok := r.SetString(s); ok {
		val, _ := r.Float64()
		if r.IsInt() {
			return val, false, true
		}
		return val, true, true
	}

	return 0, false, false
}
