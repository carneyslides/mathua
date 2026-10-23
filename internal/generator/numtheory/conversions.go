package numtheory

// Three table-driven or multiple-choice number theory concepts turned into computed questions:
// the Legendre symbol (nt.adv.legendre), the Möbius function (nt.adv.mobius), and the
// multiplicative order modulo m (nt.basics.order_mod).
//
// These three concepts previously asked learners to identify statements or choose from fixed
// option sets. Each concept has an exact integer answer computed from randomized parameters.
// The generators publish all parameters in Facts, so test suites and solution schemas can
// recompute and verify answers deterministically.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func convInt(n int) string {
	return strconv.Itoa(n)
}

type legendreConvertedGen struct{}

func legendreSymbol(a, p int) int {
	a = ((a % p) + p) % p
	if a == 0 {
		return 0
	}
	exp := (p - 1) / 2
	result := 1
	base := a
	for exp > 0 {
		if exp%2 == 1 {
			result = (result * base) % p
		}
		base = (base * base) % p
		exp /= 2
	}
	if result == p-1 {
		return -1
	}
	return result
}

func (g *legendreConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	primes := []int{3, 5, 7, 11, 13, 17, 19, 23, 29, 31}
	maxIdx := 4 + int(ctx.Difficulty*float64(len(primes)-5))
	if maxIdx >= len(primes) {
		maxIdx = len(primes) - 1
	}
	p := primes[rand.Intn(maxIdx+1)]

	// Generate a variety of residues: 0 mod p, quadratic residues, and non-residues.
	var a int
	switch rand.Intn(4) {
	case 0:
		// Multiple of p, Legendre symbol is 0.
		a = p * (1 + rand.Intn(3))
	default:
		a = 1 + rand.Intn(p+5)
	}

	ans := legendreSymbol(a, p)

	var explanation string
	switch ans {
	case 0:
		explanation = fmt.Sprintf("%d is divisible by %d, so the Legendre symbol (%d/%d) is 0.", a, p, a, p)
	case 1:
		explanation = fmt.Sprintf("%d is a quadratic residue modulo %d, so the Legendre symbol (%d/%d) is 1.", a, p, a, p)
	case -1:
		explanation = fmt.Sprintf("%d is a quadratic non-residue modulo %d, so the Legendre symbol (%d/%d) is -1.", a, p, a, p)
	}

	return generator.Problem{
		Question:    fmt.Sprintf("Compute the value of the Legendre symbol (%d/%d).", a, p),
		Answer:      convInt(ans),
		Explanation: explanation,
		Facts: map[string]string{
			"a":      convInt(a),
			"p":      convInt(p),
			"answer": convInt(ans),
		},
	}
}

type mobiusConvertedGen struct{}

func mobiusMu(n int) int {
	if n == 1 {
		return 1
	}
	p := 2
	k := 0
	rem := n
	for p*p <= rem {
		if rem%p == 0 {
			rem /= p
			if rem%p == 0 {
				return 0
			}
			k++
		}
		p++
	}
	if rem > 1 {
		k++
	}
	if k%2 == 0 {
		return 1
	}
	return -1
}

func (g *mobiusConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxN := 20 + int(ctx.Difficulty*40)
	n := 1 + rand.Intn(maxN)
	mu := mobiusMu(n)

	var explanation string
	switch mu {
	case 0:
		explanation = fmt.Sprintf("%d has a squared prime factor, so μ(%d) is 0.", n, n)
	case 1:
		if n == 1 {
			explanation = "By definition, μ(1) is 1."
		} else {
			explanation = fmt.Sprintf("%d has an even number of distinct prime factors and no squared prime factor, so μ(%d) is 1.", n, n)
		}
	case -1:
		explanation = fmt.Sprintf("%d has an odd number of distinct prime factors and no squared prime factor, so μ(%d) is -1.", n, n)
	}

	return generator.Problem{
		Question:    fmt.Sprintf("What is the value of the Möbius function μ(%d)?", n),
		Answer:      convInt(mu),
		Explanation: explanation,
		Facts: map[string]string{
			"n":      convInt(n),
			"answer": convInt(mu),
		},
	}
}

type orderModConvertedGen struct{}

func multiplicativeOrder(a, m int) int {
	cur := a % m
	k := 1
	for cur != 1 {
		cur = (cur * a) % m
		k++
	}
	return k
}

func (g *orderModConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	moduli := []int{5, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	maxIdx := 5 + int(ctx.Difficulty*float64(len(moduli)-6))
	if maxIdx >= len(moduli) {
		maxIdx = len(moduli) - 1
	}
	m := moduli[rand.Intn(maxIdx+1)]

	// Find coprime residues a in [2, m-1]
	var coprimes []int
	for candidate := 2; candidate < m; candidate++ {
		if mathutil.GCD(candidate, m) == 1 {
			coprimes = append(coprimes, candidate)
		}
	}
	a := coprimes[rand.Intn(len(coprimes))]
	ord := multiplicativeOrder(a, m)

	return generator.Problem{
		Question:    fmt.Sprintf("What is the multiplicative order of %d modulo %d?", a, m),
		Answer:      convInt(ord),
		Explanation: fmt.Sprintf("The powers of %d modulo %d cycle with period %d, which is the smallest positive integer k with %d^k ≡ 1 (mod %d).", a, m, ord, a, m),
		Facts: map[string]string{
			"a":      convInt(a),
			"m":      convInt(m),
			"answer": convInt(ord),
		},
	}
}

// registerConversions registers the number theory concepts converted from multiple-choice
// to computed numeric generators.
func registerConversions(reg *generator.Registry) {
	reg.Register("nt.adv.legendre", &legendreConvertedGen{})
	reg.Register("nt.adv.mobius", &mobiusConvertedGen{})
	reg.Register("nt.basics.order_mod", &orderModConvertedGen{})
}
