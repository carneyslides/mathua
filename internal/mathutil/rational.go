package mathutil

import (
	"fmt"
)

// Fraction represents an exact rational number num / den in lowest terms.
type Fraction struct {
	Num int
	Den int
}

// NewFraction creates a reduced fraction with a strictly positive denominator.
func NewFraction(num, den int) Fraction {
	if den == 0 {
		den = 1
	}
	if den < 0 {
		num = -num
		den = -den
	}
	g := GCD(Abs(num), den)
	if g > 1 {
		num /= g
		den /= g
	}
	return Fraction{Num: num, Den: den}
}

// Add computes f + other.
func (f Fraction) Add(other Fraction) Fraction {
	return NewFraction(f.Num*other.Den+other.Num*f.Den, f.Den*other.Den)
}

// Sub computes f - other.
func (f Fraction) Sub(other Fraction) Fraction {
	return NewFraction(f.Num*other.Den-other.Num*f.Den, f.Den*other.Den)
}

// Mul computes f * other.
func (f Fraction) Mul(other Fraction) Fraction {
	return NewFraction(f.Num*other.Num, f.Den*other.Den)
}

// Div computes f / other.
func (f Fraction) Div(other Fraction) Fraction {
	if other.Num == 0 {
		return Fraction{Num: 0, Den: 1}
	}
	return NewFraction(f.Num*other.Den, f.Den*other.Num)
}

// String formats the fraction as a clean string ("a/b" or "a").
func (f Fraction) String() string {
	if f.Den == 1 {
		return fmt.Sprintf("%d", f.Num)
	}
	return fmt.Sprintf("%d/%d", f.Num, f.Den)
}

// LaTeX formats the fraction for mathematical display.
func (f Fraction) LaTeX() string {
	if f.Den == 1 {
		return fmt.Sprintf("%d", f.Num)
	}
	if f.Num < 0 {
		return fmt.Sprintf("-\\frac{%d}{%d}", -f.Num, f.Den)
	}
	return fmt.Sprintf("\\frac{%d}{%d}", f.Num, f.Den)
}

// SimplifyRadical factors out perfect squares from sqrt(n).
// It returns coeff and radicand such that sqrt(n) = coeff * sqrt(radicand),
// where radicand is square-free.
func SimplifyRadical(n int) (coeff, radicand int) {
	if n <= 0 {
		return 0, 0
	}
	coeff = 1
	radicand = n
	for d := 2; d*d <= radicand; d++ {
		sq := d * d
		for radicand%sq == 0 {
			coeff *= d
			radicand /= sq
		}
	}
	return coeff, radicand
}

// FormatRadical renders sqrt(n) as simplified LaTeX.
// For example, 75 becomes "5\\sqrt{3}", 16 becomes "4" and 7 becomes "\\sqrt{7}".
func FormatRadical(n int) string {
	if n == 0 {
		return "0"
	}
	coeff, rad := SimplifyRadical(n)
	if rad == 1 {
		return fmt.Sprintf("%d", coeff)
	}
	if coeff == 1 {
		return fmt.Sprintf("\\sqrt{%d}", rad)
	}
	return fmt.Sprintf("%d\\sqrt{%d}", coeff, rad)
}
