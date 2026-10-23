package mathutil

import (
	"fmt"
	"strings"
)

// Polynomial represents a single-variable integer polynomial:
// P(x) = Coeffs[0] + Coeffs[1]*x + Coeffs[2]*x^2 + ...
type Polynomial struct {
	Coeffs []int
}

// NewPolynomial creates a polynomial from ascending coefficients:
// c[0] + c[1]*x + c[2]*x^2 + ...
func NewPolynomial(coeffs ...int) Polynomial {
	// Trim trailing zero coefficients while keeping at least one coefficient
	last := len(coeffs) - 1
	for last > 0 && coeffs[last] == 0 {
		last--
	}
	res := make([]int, last+1)
	copy(res, coeffs[:last+1])
	return Polynomial{Coeffs: res}
}

// PolyFromRoots constructs a monic polynomial with the specified integer roots:
// P(x) = (x - r_1)(x - r_2)...(x - r_k)
func PolyFromRoots(roots ...int) Polynomial {
	if len(roots) == 0 {
		return NewPolynomial(1)
	}
	// Start with (x - r_0) = -r_0 + 1*x
	curr := []int{-roots[0], 1}
	for _, r := range roots[1:] {
		next := make([]int, len(curr)+1)
		for i, c := range curr {
			next[i] += -r * c
			next[i+1] += c
		}
		curr = next
	}
	return Polynomial{Coeffs: curr}
}

// Degree returns the highest power of x with non-zero coefficient.
func (p Polynomial) Degree() int {
	return len(p.Coeffs) - 1
}

// Eval evaluates the polynomial at an integer value x using Horner's method.
func (p Polynomial) Eval(x int) int {
	if len(p.Coeffs) == 0 {
		return 0
	}
	res := p.Coeffs[len(p.Coeffs)-1]
	for i := len(p.Coeffs) - 2; i >= 0; i-- {
		res = res*x + p.Coeffs[i]
	}
	return res
}

// Derivative computes the formal derivative of the polynomial:
// P'(x) = c[1] + 2*c[2]*x + 3*c[3]*x^2 + ...
func (p Polynomial) Derivative() Polynomial {
	if len(p.Coeffs) <= 1 {
		return NewPolynomial(0)
	}
	deriv := make([]int, len(p.Coeffs)-1)
	for i := 1; i < len(p.Coeffs); i++ {
		deriv[i-1] = i * p.Coeffs[i]
	}
	return NewPolynomial(deriv...)
}

// Format renders the polynomial in conventional descending-power algebraic notation.
// For example, "x^2 - 5x + 6".
func (p Polynomial) Format(variable string) string {
	if len(p.Coeffs) == 0 {
		return "0"
	}
	if len(p.Coeffs) == 1 {
		return fmt.Sprintf("%d", p.Coeffs[0])
	}

	var sb strings.Builder
	first := true

	for deg := len(p.Coeffs) - 1; deg >= 0; deg-- {
		c := p.Coeffs[deg]
		if c == 0 {
			continue
		}

		if first {
			if c < 0 {
				sb.WriteString("-")
			}
			absC := Abs(c)
			if deg == 0 {
				sb.WriteString(fmt.Sprintf("%d", absC))
			} else if deg == 1 {
				if absC != 1 {
					sb.WriteString(fmt.Sprintf("%d", absC))
				}
				sb.WriteString(variable)
			} else {
				if absC != 1 {
					sb.WriteString(fmt.Sprintf("%d", absC))
				}
				sb.WriteString(fmt.Sprintf("%s^%d", variable, deg))
			}
			first = false
			continue
		}

		if c < 0 {
			sb.WriteString(" - ")
		} else {
			sb.WriteString(" + ")
		}

		absC := Abs(c)
		if deg == 0 {
			sb.WriteString(fmt.Sprintf("%d", absC))
		} else if deg == 1 {
			if absC != 1 {
				sb.WriteString(fmt.Sprintf("%d", absC))
			}
			sb.WriteString(variable)
		} else {
			if absC != 1 {
				sb.WriteString(fmt.Sprintf("%d", absC))
			}
			sb.WriteString(fmt.Sprintf("%s^%d", variable, deg))
		}
	}

	if first {
		return "0"
	}
	return sb.String()
}
