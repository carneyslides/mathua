package abstract

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

// batchLCM returns the least common multiple of two positive integers.
//
// The batch conversions need the order of a product of disjoint cycles and the
// order of a product of cyclic groups, both of which are least common
// multiples. mathutil exports GCD only, so this is defined here rather than
// widening that package's surface for one caller.
func batchLCM(a, b int) int { return a / mathutil.GCD(a, b) * b }

func Register(reg *generator.Registry) {
	reg.Register("abstract.group.def", &groupDefGen{})
	reg.Register("abstract.group.examples", &groupExamplesGen{})
	reg.Register("abstract.group.subgroup", &subgroupGen{})
	reg.Register("abstract.rings.def", &ringGen{})
	reg.Register("abstract.structures.field", &fieldGen{})
	reg.Register("abstract.structures.module", &moduleGen{})
	reg.Register("abstract.rings.ideal", &idealGen{})
	reg.Register("abstract.rings.quotient", &quotientGen{})
	reg.Register("abstract.group.normal", &normalSubgroupGen{})
	reg.Register("abstract.group.quotient_group", &quotientGroupGen{})
	reg.Register("abstract.rings.polynomial", &polynomialRingGen{})
	reg.Register("abstract.rings.ufd", &ufdGen{})
	reg.Register("abstract.field.extension", &fieldExtensionGen{})
	reg.Register("abstract.structures.galois", &galoisGen{})
	reg.Register("abstract.rings.localization", &localizationGen{})
	reg.Register("abstract.group.nilpotent", &nilpotentGen{})
	reg.Register("abstract.structures.tensor", &tensorGen{})
	reg.Register("abstract.group.solvable", &solvableGen{})
	reg.Register("abstract.group.free", &freeGroupGen{})
	reg.Register("abstract.rings.noetherian", &noetherianGen{})
	reg.Register("abstract.field.galois_group", &galoisGroupGen{})
	reg.Register("abstract.structures.exact_sequence", &exactSequenceGen{})
	reg.Register("abstract.group.commutator", &commutatorGen{})
	registerBatch(reg)
	registerConversions(reg)
}

// ----- 1. group.def -----

type groupDefGen struct{}

func (g *groupDefGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Dynamic multiples kZ under addition
		k := 2 + rand.Intn(10)
		q := fmt.Sprintf("Is the set of multiples %d\\mathbb{Z} under addition a group? (yes/no)", k)
		ex := fmt.Sprintf("Multiples of %d under addition are closed, associative, have identity 0 = %d(0) and each %dn has inverse %d(-n), satisfying all group axioms.", k, k, k, k)
		return groupYesNo(q, "yes", ex)
	case 1:
		// Integers modulo m without zero under multiplication
		primes := []int{3, 5, 7, 11, 13, 17, 19}
		composites := []int{4, 6, 8, 9, 10, 12, 14, 15, 16}
		if rand.Intn(2) == 0 {
			p := primes[rand.Intn(len(primes))]
			q := fmt.Sprintf("Is the set \\(\\mathbb{Z}_{%d} \\setminus \\{0\\}\\) under multiplication modulo %d a group? (yes/no)", p, p)
			ex := fmt.Sprintf("Because %d is prime, every non-zero residue is coprime to %d and possesses a modular inverse, forming a group under multiplication.", p, p)
			return groupYesNo(q, "yes", ex)
		}
		m := composites[rand.Intn(len(composites))]
		q := fmt.Sprintf("Is the set \\(\\mathbb{Z}_{%d} \\setminus \\{0\\}\\) under multiplication modulo %d a group? (yes/no)", m, m)
		ex := fmt.Sprintf("Because %d is composite, it contains non-trivial factors whose product is 0 mod %d, violating closure and inverse axioms.", m, m)
		return groupYesNo(q, "no", ex)
	case 2:
		// Failing axioms identification
		type failItem struct {
			set       string
			operation string
			axiom     string
			reason    string
		}
		fails := []failItem{
			{"natural numbers \\(\\mathbb{N}\\)", "addition", "inverses", "Positive natural numbers have no additive inverses in N."},
			{"odd integers", "addition", "closure", "Adding two odd integers produces an even integer, which leaves the set."},
			{"integers \\(\\mathbb{Z}\\)", "multiplication", "inverses", "Integers other than 1 and -1 lack multiplicative inverses in Z."},
			{"positive integers", "multiplication", "inverses", "Integers greater than 1 have no multiplicative inverse in the positive integers."},
			{"positive rational numbers", "addition", "identity", "The additive identity 0 is not positive, so no identity exists in the set."},
			{"even integers", "multiplication", "identity", "The multiplicative identity 1 is odd, so the set has no identity element."},
		}
		item := fails[rand.Intn(len(fails))]
		q := fmt.Sprintf("The set of %s under %s is NOT a group. Which axiom fails?", item.set, item.operation)
		return groupChoice(q, item.axiom, item.reason, "closure", "identity", "inverses")
	default:
		// Standard structures
		type stdItem struct {
			set       string
			operation string
			isGroup   string
			reason    string
		}
		stds := []stdItem{
			{"rational numbers \\(\\mathbb{Q}\\)", "addition", "yes", "Rationals under addition satisfy closure, associativity, identity 0 and inverses -a."},
			{"real numbers \\(\\mathbb{R}\\)", "addition", "yes", "Real numbers under addition satisfy closure, associativity, identity 0 and inverses -a."},
			{"non-zero rational numbers \\(\\mathbb{Q} \\setminus \\{0\\}\\)", "multiplication", "yes", "Non-zero rationals under multiplication satisfy all group properties with identity 1 and inverses 1/a."},
			{"non-zero real numbers \\(\\mathbb{R} \\setminus \\{0\\}\\)", "multiplication", "yes", "Non-zero reals under multiplication form a group with identity 1 and inverses 1/a."},
			{"invertible 2x2 real matrices \\(\\operatorname{GL}_2(\\mathbb{R})\\)", "multiplication", "yes", "Invertible matrices under multiplication form the general linear group with identity I and matrix inverses."},
			{"all 2x2 real matrices \\(M_2(\\mathbb{R})\\)", "multiplication", "no", "Singular matrices with zero determinant have no multiplicative inverse."},
		}
		item := stds[rand.Intn(len(stds))]
		q := fmt.Sprintf("Is the set of %s under %s a group? (yes/no)", item.set, item.operation)
		return groupYesNo(q, item.isGroup, item.reason)
	}
}

// ----- 2. group.examples -----

type groupExamplesGen struct{}

func (g *groupExamplesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Order of Dihedral group D_{2n}
		n := 3 + rand.Intn(10)
		ans := strconv.Itoa(2 * n)
		q := fmt.Sprintf("What is the order of the group of symmetries of a regular %d-gon? (enter a number)", n)
		ex := fmt.Sprintf("A regular %d-gon has %d rotations and %d reflections, giving total order 2(%d) = %d.", n, n, n, n, 2*n)
		return groupChoice(q, ans, ex, strconv.Itoa(n), strconv.Itoa(2*n-2), strconv.Itoa(2*n+2))
	case 1:
		// Order of unit groups U(p)
		primes := []int{3, 5, 7, 11, 13, 17}
		p := primes[rand.Intn(len(primes))]
		ans := strconv.Itoa(p - 1)
		q := fmt.Sprintf("What is the order of the group of units \\(U(%d) = \\mathbb{Z}_{%d}^\\times\\) under multiplication? (enter a number)", p, p)
		ex := fmt.Sprintf("Because %d is prime, all residues from 1 to %d are coprime to %d, so the order is %d - 1 = %d.", p, p-1, p, p, p-1)
		return groupChoice(q, ans, ex, strconv.Itoa(p), strconv.Itoa(p-2), strconv.Itoa(p+1))
	case 2:
		// Additive inverse in Z_n
		n := 5 + rand.Intn(15)
		a := 1 + rand.Intn(n-2)
		for 2*a == n {
			a = 1 + rand.Intn(n-2)
		}
		inv := n - a
		ans := strconv.Itoa(inv)
		q := fmt.Sprintf("What is the additive inverse of %d in the group \\(\\mathbb{Z}_{%d}\\)? (enter a number)", a, n)
		ex := fmt.Sprintf("In \\(\\mathbb{Z}_{%d}\\), %d + %d = %d \\equiv 0 \\pmod{%d}, so the inverse is %d.", n, a, inv, n, n, inv)
		return groupChoice(q, ans, ex, strconv.Itoa(a), strconv.Itoa((inv+1)%n), strconv.Itoa((inv+n-1)%n))
	case 3:
		// Abelian property questions
		if rand.Intn(2) == 0 {
			n := 3 + rand.Intn(5)
			q := fmt.Sprintf("Is the symmetric group \\(S_{%d}\\) abelian? (yes/no)", n)
			ex := fmt.Sprintf("For n >= 3, disjoint transposition and 3-cycle products do not commute, so S_%d is non-abelian.", n)
			return groupYesNo(q, "no", ex)
		}
		n := 3 + rand.Intn(5)
		q := fmt.Sprintf("Is the dihedral group \\(D_{%d}\\) of symmetries of a regular %d-gon abelian? (yes/no)", 2*n, n)
		ex := fmt.Sprintf("For n >= 3, reflecting and then rotating does not equal rotating and then reflecting, so D_%d is non-abelian.", 2*n)
		return groupYesNo(q, "no", ex)
	default:
		// Order of alternating group A_n
		n := 3 + rand.Intn(4) // 3 to 6
		order := abstractFactorial(n) / 2
		ans := strconv.Itoa(order)
		q := fmt.Sprintf("What is the order of the alternating group \\(A_{%d}\\)? (enter a number)", n)
		ex := fmt.Sprintf("The alternating group contains exactly half the permutations of S_%d, giving order %d! / 2 = %d / 2 = %d.", n, n, abstractFactorial(n), order)
		return groupChoice(q, ans, ex, strconv.Itoa(abstractFactorial(n)), strconv.Itoa(order-1), strconv.Itoa(order+2))
	}
}

// ----- 3. subgroup -----

type subgroupGen struct{}

func (g *subgroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		// Cyclic subgroup <d> in Z_n where n = d * k -> order is k
		d := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		k := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		n := d * k
		return generator.Problem{
			Question: fmt.Sprintf("In the additive cyclic group \\(\\mathbb{Z}_{%d}\\), find the order of the subgroup \\(\\langle %d \\rangle\\).", n, d),
			Answer:   fmt.Sprintf("%d", k),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula: the order of \\(\\langle d \\rangle\\) in \\(\\mathbb{Z}_n\\) is \\(\\frac{n}{\\gcd(d, n)}\\).\n"+
				"Subgoal 2: Substitute \\(n = %d\\) and \\(d = %d\\), and \\(\\gcd(%d, %d) = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", n, d, d, n, d, n, d, k),
		}
	case 1:
		// Number of left cosets [G:H] = |G| / |H|
		hOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		index := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := hOrder * index
		return generator.Problem{
			Question: fmt.Sprintf("A finite group \\(G\\) of order %d contains a subgroup \\(H\\) of order %d. How many distinct left cosets of \\(H\\) exist in \\(G\\)?", gOrder, hOrder),
			Answer:   fmt.Sprintf("%d", index),
			Explanation: fmt.Sprintf("Subgoal 1: State Lagrange's Theorem: the number of distinct left cosets is the index \\([G:H] = \\frac{|G|}{|H|}\\).\n"+
				"Subgoal 2: Substitute \\(|G| = %d\\) and \\(|H| = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", gOrder, hOrder, gOrder, hOrder, index),
		}
	case 2:
		// A product of disjoint cycles has order lcm of the cycle lengths.
		a := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		b := a + mathutil.ScaleInt(ctx.Difficulty, 1, 4)
		n := a + b + mathutil.ScaleInt(ctx.Difficulty, 0, 3)
		ans := batchLCM(a, b)
		return generator.Problem{
			Question: fmt.Sprintf("In the symmetric group \\(S_{%d}\\), find the order of the cyclic subgroup generated by a permutation that is a product of two disjoint cycles of lengths %d and %d.", n, a, b),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula: disjoint cycles commute, so the order of their product is \\(\\operatorname{lcm}\\) of the cycle lengths.\n"+
				"Subgoal 2: Substitute the lengths %d and %d.\n"+
				"Subgoal 3: Evaluate: \\(\\operatorname{lcm}(%d, %d) = %d\\).", a, b, a, b, ans),
		}
	case 3:
		// Number of generators of the subgroup of prime order p is phi(p) = p - 1.
		primes := []int{2, 3, 5, 7, 11, 13}
		p := primes[rand.Intn(len(primes))]
		m := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		n := p * m
		ans := p - 1
		return generator.Problem{
			Question: fmt.Sprintf("In \\(\\mathbb{Z}_{%d}\\), how many distinct elements generate the unique subgroup of prime order %d?", n, p),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: A cyclic group of order \\(m\\) has \\(\\phi(m)\\) generators.\n"+
				"Subgoal 2: For prime order \\(p = %d\\), \\(\\phi(p) = p - 1\\).\n"+
				"Subgoal 3: Evaluate: \\(\\phi(%d) = %d\\).", p, p, ans),
		}
	default:
		// Order of an element a in Z_n is n / gcd(a, n).
		a := mathutil.ScaleInt(ctx.Difficulty, 2, 12)
		n := mathutil.ScaleInt(ctx.Difficulty, 4, 24)
		ans := n / mathutil.GCD(a, n)
		return generator.Problem{
			Question: fmt.Sprintf("In the additive group \\(\\mathbb{Z}_{%d}\\), find the order of the element %d.", n, a),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula: the order of \\(a\\) in \\(\\mathbb{Z}_n\\) is \\(\\frac{n}{\\gcd(a, n)}\\).\n"+
				"Subgoal 2: Substitute \\(n = %d\\) and \\(a = %d\\), and \\(\\gcd(%d, %d) = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", n, a, a, n, mathutil.GCD(a, n), n, mathutil.GCD(a, n), ans),
		}
	}
}

// ----- 4. ring -----

type ringGen struct{}

func (g *ringGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Modular field checks
		primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23}
		composites := []int{4, 6, 8, 9, 10, 12, 14, 15, 16, 18, 20, 21, 25}
		if rand.Intn(2) == 0 {
			p := primes[rand.Intn(len(primes))]
			q := fmt.Sprintf("Is \\(\\mathbb{Z}_{%d}\\) a field under addition and multiplication modulo %d? (yes/no)", p, p)
			ex := fmt.Sprintf("Because %d is prime, every non-zero element has a multiplicative inverse, so \\(\\mathbb{Z}_{%d}\\) is a field.", p, p)
			return groupYesNo(q, "yes", ex)
		}
		m := composites[rand.Intn(len(composites))]
		q := fmt.Sprintf("Is \\(\\mathbb{Z}_{%d}\\) a field under addition and multiplication modulo %d? (yes/no)", m, m)
		ex := fmt.Sprintf("Because %d is composite, its non-trivial divisors are zero divisors and lack multiplicative inverses, so it is not a field.", m)
		return groupYesNo(q, "no", ex)
	case 1:
		// Witness non-unit element
		type witEntry struct {
			set    string
			wit    string
			d1, d2 string
			reason string
		}
		witTable := []witEntry{
			{"\\(\\mathbb{Z}\\) (integers)", "2", "0", "1", "2 has no integer multiplicative inverse: no integer times 2 is 1."},
			{"\\(\\mathbb{Z}_4\\)", "2", "1", "3", "2 * 2 = 4 = 0 mod 4, so 2 is a zero divisor with no inverse."},
			{"\\(\\mathbb{Z}_6\\)", "2", "1", "5", "2 * 3 = 6 = 0 mod 6, so 2 has no multiplicative inverse mod 6."},
			{"\\(\\mathbb{Z}_8\\)", "2", "1", "3", "2 * 4 = 8 = 0 mod 8, so 2 has no inverse mod 8."},
			{"\\(\\mathbb{Z}_9\\)", "3", "1", "2", "3 * 3 = 9 = 0 mod 9, so 3 has no multiplicative inverse mod 9."},
			{"\\(\\mathbb{Z}_{10}\\)", "2", "1", "3", "2 * 5 = 10 = 0 mod 10, so 2 has no inverse mod 10."},
			{"\\(\\mathbb{Z}_{12}\\)", "2", "1", "5", "2 * 6 = 12 = 0 mod 12, so 2 has no inverse mod 12."},
			{"\\(\\mathbb{Z}_{14}\\)", "2", "1", "3", "2 * 7 = 14 = 0 mod 14, so 2 has no inverse mod 14."},
			{"\\(\\mathbb{Z}_{15}\\)", "3", "1", "2", "3 * 5 = 15 = 0 mod 15, so 3 has no inverse mod 15."},
			{"\\(\\mathbb{Z}_{21}\\)", "3", "1", "2", "3 * 7 = 21 = 0 mod 21, so 3 has no inverse mod 21."},
			{"\\(\\mathbb{Z}_{25}\\)", "5", "1", "2", "5 * 5 = 25 = 0 mod 25, so 5 has no inverse mod 25."},
			{"\\(\\mathbb{Z}[i]\\) (Gaussian integers)", "2", "0", "1", "2 has no inverse in Z[i] because (a+bi)(c+di)=1 requires norm 1 while N(2)=4."},
		}
		w := witTable[rand.Intn(len(witTable))]
		q := fmt.Sprintf("What is the smallest positive integer with no multiplicative inverse in %s? (enter a number)", w.set)
		return groupChoice(q, w.wit, w.reason, w.d1, w.d2, "4")
	case 2:
		// Standard set ring and field queries
		type setEntry struct {
			set     string
			isRing  string
			isField string
			rReason string
			fReason string
		}
		sets := []setEntry{
			{"\\(\\mathbb{Q}\\) (rational numbers)", "yes", "yes", "Q is a commutative ring with unity under addition and multiplication.", "Q is a field because every non-zero rational a/b has inverse b/a."},
			{"\\(\\mathbb{R}\\) (real numbers)", "yes", "yes", "R is a commutative ring with unity under standard arithmetic.", "R is a field because every non-zero real number has a multiplicative inverse."},
			{"\\(\\mathbb{Z}\\) (integers)", "yes", "no", "Z is a ring with unity under addition and multiplication.", "Z is not a field because elements other than 1 and -1 lack multiplicative inverses in Z."},
			{"\\(2\\mathbb{Z}\\) (even integers)", "yes", "no", "2Z is a ring without unity (a rng) under addition and multiplication.", "2Z is not a field because it has no multiplicative identity."},
			{"\\(\\mathbb{Z}[i]\\) (Gaussian integers)", "yes", "no", "Gaussian integers a+bi form an integral domain under addition and multiplication.", "Gaussian integers do not form a field because non-units like 2 have no inverse in Z[i]."},
			{"\\(M_2(\\mathbb{R})\\) (2x2 real matrices)", "yes", "no", "Square matrices under matrix addition and multiplication form a non-commutative ring with identity.", "M_2(R) is not a field because matrix multiplication is non-commutative and singular matrices lack inverses."},
			{"\\(\\mathbb{R}[x]\\) (polynomials with real coefficients)", "yes", "no", "Polynomials form a commutative ring with unity under polynomial addition and multiplication.", "R[x] is not a field because non-constant polynomials have no polynomial multiplicative inverse."},
		}
		s := sets[rand.Intn(len(sets))]
		if rand.Intn(2) == 0 {
			q := fmt.Sprintf("Is %s a ring under addition and multiplication? (yes/no)", s.set)
			return groupYesNo(q, s.isRing, s.rReason)
		}
		q := fmt.Sprintf("Is %s a field under addition and multiplication? (yes/no)", s.set)
		return groupYesNo(q, s.isField, s.fReason)
	default:
		// Modular ring verification
		m := 2 + rand.Intn(20)
		q := fmt.Sprintf("Is \\(\\mathbb{Z}_{%d}\\) a commutative ring with unity under addition and multiplication modulo %d? (yes/no)", m, m)
		ex := fmt.Sprintf("For every integer n >= 2, the integers modulo n form a commutative ring with identity element 1.")
		return groupYesNo(q, "yes", ex)
	}
}

// ----- 5. homomorphism -----

type homomorphismGen struct{}

func (g *homomorphismGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		f        string
		domain   string
		codomain string
		isHomo   string
		reason   string
	}
	table := []entry{
		{"f(x)=2x", "Z", "Z", "yes", "f(a+b)=2(a+b)=2a+2b=f(a)+f(b), so it preserves the group operation."},
		{"f(x)=x+1", "Z", "Z", "no", "f(a+b)=a+b+1, but f(a)+f(b)=(a+1)+(b+1)=a+b+2. These are not equal."},
		{"f(x)=0", "Z", "Z", "yes", "f(a+b)=0 and f(a)+f(b)=0+0=0, so it is the trivial homomorphism."},
		{"f(x)=−x", "Z", "Z", "yes", "f(a+b)=−(a+b)=−a−b=(−a)+(−b)=f(a)+f(b)."},
		{"f(x)=|x|", "Z", "Z", "no", "f(1+(−1))=f(0)=0, but f(1)+f(−1)=1+1=2. These are not equal."},
		{"f(x)=x²", "Z", "Z", "no", "f(a+b)=(a+b)²=a²+2ab+b², but f(a)+f(b)=a²+b². These differ unless 2ab=0."},
		{"f(x)=eˣ", "R", "R⁺", "yes", "f(a+b)=e^{a+b}=e^a·e^b, so it maps the additive group of R to the multiplicative group of R⁺."},
		{"f(x)=ln(x)", "R⁺", "R", "yes", "f(ab)=ln(ab)=ln(a)+ln(b), so it maps the multiplicative group of R⁺ to the additive group of R."},
		{"f(x)=det(x)", "GL(2,R)", "R\\{0}", "yes", "det(AB)=det(A)det(B), so it is a group homomorphism from GL(2,R) to R\\{0}."},
		{"f(x)=x mod 2", "Z", "Z₂", "yes", "f(a+b)=(a+b) mod 2 = (a mod 2)+(b mod 2) mod 2 = f(a)+f(b)."},
	}
	e := table[rand.Intn(len(table))]
	if rand.Intn(3) == 0 {
		// production: evaluate a homomorphism at a point.
		type evalEntry struct {
			f      string
			point  string
			value  string
			reason string
		}
		evalTable := []evalEntry{
			{"f(x)=2x from Z to Z", "3", "6", "f(3)=2·3=6."},
			{"f(x)=x mod 2 from Z to Z_2", "5", "1", "5 mod 2 = 1."},
			{"f(x)=det(x) on GL(2,R)", "I (identity)", "1", "det(I)=1."},
		}
		v := evalTable[rand.Intn(len(evalTable))]
		return generator.Problem{
			Question:    fmt.Sprintf("For %s, what is f(%s)? (enter the value)", v.f, v.point),
			Answer:      v.value,
			Explanation: v.reason,
		}
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Is %s from %s to %s a group homomorphism? (yes/no)", e.f, e.domain, e.codomain),
		Answer:      e.isHomo,
		Explanation: e.reason,
	}
}

// ----- 6. field -----

type fieldGen struct{}

func (g *fieldGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		set     string
		isField string
		reason  string
	}
	table := []entry{
		{"Q (rational numbers)", "yes", "Q is a field: every non-zero rational a/b has inverse b/a, and all field axioms are satisfied."},
		{"R (real numbers)", "yes", "R is a field: every non-zero real has a multiplicative inverse, satisfying all field axioms."},
		{"C (complex numbers)", "yes", "C is a field: every non-zero complex number a+bi has inverse (a-bi)/(a²+b²)."},
		{"Z (integers)", "no", "Z is not a field: most integers lack multiplicative inverses (e.g., 2 has no inverse in Z)."},
		{"2Z (even integers)", "no", "2Z is not a field: it has no multiplicative identity and lacks inverses."},
		{"Z₅ (integers mod 5)", "yes", "Z₅ is a field because 5 is prime; every non-zero element has a multiplicative inverse mod 5."},
		{"Z₆ (integers mod 6)", "no", "Z₆ is not a field: 2 and 3 are zero divisors with no multiplicative inverses mod 6."},
		{"Z₁₁ (integers mod 11)", "yes", "Z₁₁ is a field because 11 is prime; every non-zero element is invertible mod 11."},
		{"n×n matrices over R (n≥2)", "no", "Matrix multiplication is not commutative and many matrices have no inverse, so it fails field axioms."},
		{"Z[i] (Gaussian integers)", "no", "Gaussian integers form a ring but not a field: 2 has no inverse in Z[i]."},
		{"F₄ (finite field with 4 elements)", "yes", "F₄ exists because 4=2² is a prime power. Every non-zero element has an inverse."},
	}
	e := table[rand.Intn(len(table))]
	if rand.Intn(2) == 0 {
		return generator.Problem{
			Question:    fmt.Sprintf("Is %s a field under addition and multiplication? (yes/no)", e.set),
			Answer:      e.isField,
			Explanation: e.reason,
		}
	}
	// characteristic question
	charTable := []struct {
		field  string
		char   string
		reason string
	}{
		{"Q", "0", "Q has characteristic 0: no finite sum of 1s equals 0."},
		{"R", "0", "R has characteristic 0: no finite sum of 1s equals 0."},
		{"C", "0", "C has characteristic 0: no finite sum of 1s equals 0."},
		{"Z₅", "5", "Z₅ has characteristic 5: 1+1+1+1+1 = 5 ≡ 0 mod 5."},
		{"Z₁₁", "11", "Z₁₁ has characteristic 11: 1+1+...+1 (11 times) ≡ 0 mod 11."},
		{"F₄", "2", "F₄ has characteristic 2: 1+1 = 0 in any field of order 2ⁿ."},
	}
	c := charTable[rand.Intn(len(charTable))]
	return generator.Problem{
		Question:    fmt.Sprintf("What is the characteristic of %s?", c.field),
		Answer:      c.char,
		Explanation: c.reason,
	}
}

// ----- 7. module -----

// ----- 8. cyclic -----

type cyclicGen struct{}

func (g *cyclicGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		group  string
		cyclic string
		reason string
	}
	table := []entry{
		{"Z_n (integers mod n)", "yes", "Z_n under addition is cyclic, generated by 1."},
		{"Z (integers under addition)", "yes", "Z is cyclic, generated by 1 (or -1)."},
		{"Klein four-group V_4", "no", "V_4 is not cyclic: every non-identity element has order 2, none generates the whole group."},
		{"Z_2 × Z_2", "no", "Direct product of two order-2 cyclic groups is not cyclic (no element of order 4)."},
		{"S_3 (symmetric group on 3)", "no", "S_3 is not cyclic: no element has order 6; max order is 3."},
		{"U(5) = {1,2,3,4} under multiplication mod 5", "yes", "U(5) is cyclic of order 4, generated by 2 (2^1=2, 2^2=4, 2^3=3, 2^4=1)."},
		{"Q under addition", "no", "Q is not cyclic: no single rational generates all rationals."},
	}
	e := table[rand.Intn(len(table))]
	if rand.Intn(3) == 0 {
		// production: orders and generators in cyclic groups.
		type cycEntry struct {
			question string
			answer   string
			reason   string
		}
		cycTable := []cycEntry{
			{"U(5) = {1,2,3,4} under multiplication mod 5 is cyclic. Name a generator (enter a number 1-4)", "2", "2^1=2, 2^2=4, 2^3=3, 2^4=1: 2 generates all of U(5)."},
			{"What is the order of the Klein four-group V_4? (enter a number)", "4", "V_4 = {e,a,b,ab} has 4 elements."},
			{"S_3 is not cyclic. What is the largest element order in S_3? (enter a number)", "3", "3-cycles have order 3; nothing has order 6."},
		}
		c := cycTable[rand.Intn(len(cycTable))]
		return generator.Problem{Question: c.question, Answer: c.answer, Explanation: c.reason}
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Is the group %s cyclic? (yes/no)", e.group),
		Answer:      e.cyclic,
		Explanation: e.reason,
	}
}

// ----- 9. order -----

type orderGen struct{}

func (g *orderGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		elem   string
		group  string
		order  string
		reason string
	}
	table := []entry{
		{"2", "Z_6 (addition mod 6)", "3", "2+2=4, 2+2+2=0 mod 6, so order of 2 is 3."},
		{"3", "Z_6 (addition mod 6)", "2", "3+3=0 mod 6, so order of 3 is 2."},
		{"1", "Z_6 (addition mod 6)", "6", "1 generates Z_6, so its order is 6."},
		{"2", "Z_4 (addition mod 4)", "2", "2+2=0 mod 4, so order is 2."},
		{"(1 2)", "S_3", "2", "Transposition (1 2) has order 2: applying twice returns identity."},
		{"(1 2 3)", "S_3", "3", "3-cycle (1 2 3) has order 3."},
		{"-1", "Z (addition)", "infinite", "No finite n gives n*(-1)=0 in Z (except 0), so order is infinite."},
	}
	e := table[rand.Intn(len(table))]
	return generator.Problem{
		Question:    fmt.Sprintf("What is the order of %s in %s?", e.elem, e.group),
		Answer:      e.order,
		Explanation: e.reason,
	}
}

// ----- 10. ideal -----

type idealGen struct{}

func (g *idealGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		// |(d)| in Z_n where n = d * k is k
		d := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		k := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		n := d * k
		return generator.Problem{
			Question: fmt.Sprintf("In the commutative ring \\(\\mathbb{Z}_{%d}\\), how many elements belong to the principal ideal \\((%d)\\)?", n, d),
			Answer:   fmt.Sprintf("%d", k),
			Explanation: fmt.Sprintf("Subgoal 1: The ideal \\((%d)\\) in \\(\\mathbb{Z}_{%d}\\) consists of the multiples of %d, so it has \\(\\frac{n}{\\gcd(d, n)}\\) elements.\n"+
				"Subgoal 2: Here \\(%d \\mid %d\\), so \\(\\gcd(%d, %d) = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", d, n, d, d, n, d, n, d, n, d, k),
		}
	case 1:
		// Index of (d) in Z_n when d does not divide n is gcd(d, n).
		d := mathutil.ScaleInt(ctx.Difficulty, 3, 9)
		n := mathutil.ScaleInt(ctx.Difficulty, 4, 20)
		for n%d == 0 {
			n++
		}
		ans := mathutil.GCD(d, n)
		return generator.Problem{
			Question: fmt.Sprintf("Find the number of distinct cosets (index) of the ideal \\((%d)\\) in the ring \\(\\mathbb{Z}_{%d}\\).", d, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The index of \\(I\\) in \\(R\\) is \\(|R/I|\\), and \\(|\\mathbb{Z}_n / (d)| = \\gcd(d, n)\\).\n"+
				"Subgoal 2: Substitute \\(d = %d\\) and \\(n = %d\\); note that %d does not divide %d, so the index is not %d.\n"+
				"Subgoal 3: Evaluate: \\(\\gcd(%d, %d) = %d\\).", d, n, d, n, d, d, n, ans),
		}
	case 2:
		// |(a)| in Z_n is n / gcd(a, n)
		a := mathutil.ScaleInt(ctx.Difficulty, 2, 10)
		n := mathutil.ScaleInt(ctx.Difficulty, 5, 24)
		ans := n / mathutil.GCD(a, n)
		return generator.Problem{
			Question: fmt.Sprintf("How many elements belong to the principal ideal \\((%d)\\) in the ring \\(\\mathbb{Z}_{%d}\\)?", a, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The ideal \\((%d)\\) is the set of multiples of %d modulo %d, which is the subgroup generated by \\(\\gcd(a, n)\\).\n"+
				"Subgoal 2: Its size is \\(\\frac{n}{\\gcd(a, n)} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d elements.", a, a, n, n, mathutil.GCD(a, n), ans),
		}
	case 3:
		// Sum of ideals (a) + (b) = (gcd(a, b)) in Z
		g := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		x := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		y := mathutil.ScaleInt(ctx.Difficulty, 3, 7)
		for mathutil.GCD(x, y) != 1 {
			y++
		}
		a := g * x
		b := g * y
		return generator.Problem{
			Question: fmt.Sprintf("In the principal ideal domain \\(\\mathbb{Z}\\), find the positive generator \\(d\\) of the sum ideal \\((%d) + (%d) = (d)\\).", a, b),
			Answer:   fmt.Sprintf("%d", g),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula for PID ideal sums: \\((a) + (b) = (\\gcd(a, b))\\).\n"+
				"Subgoal 2: Substitute \\(a = %d\\) and \\(b = %d\\).\n"+
				"Subgoal 3: Compute: \\(\\gcd(%d, %d) = %d\\).", a, b, a, b, g),
		}
	default:
		// Intersection of ideals (a) cap (b) = (lcm(a, b)) in Z
		g := mathutil.ScaleInt(ctx.Difficulty, 1, 4)
		x := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		y := mathutil.ScaleInt(ctx.Difficulty, 3, 5)
		for mathutil.GCD(x, y) != 1 {
			y++
		}
		a := g * x
		b := g * y
		ans := g * x * y
		return generator.Problem{
			Question: fmt.Sprintf("In the ring of integers \\(\\mathbb{Z}\\), find the positive generator \\(m\\) of the intersection ideal \\((%d) \\cap (%d) = (m)\\).", a, b),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula: \\((a) \\cap (b) = (\\operatorname{lcm}(a, b))\\).\n"+
				"Subgoal 2: Compute \\(\\operatorname{lcm}(%d, %d) = \\frac{%d \\cdot %d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d.", a, b, a, b, g, ans),
		}
	}
}

// ----- 11. quotient -----

type quotientGen struct{}

func (g *quotientGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		// |Z_n / (d)| = gcd(d, n), and d does not divide n so this is not d.
		d := mathutil.ScaleInt(ctx.Difficulty, 3, 9)
		n := mathutil.ScaleInt(ctx.Difficulty, 4, 20)
		for n%d == 0 {
			n++
		}
		ans := mathutil.GCD(d, n)
		return generator.Problem{
			Question: fmt.Sprintf("How many elements are in the quotient ring \\(\\mathbb{Z}_{%d} / (%d)\\)?", n, d),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: For an ideal \\((%d)\\) in \\(\\mathbb{Z}_n\\), the quotient has \\(\\gcd(d, n)\\) elements.\n"+
				"Subgoal 2: Substitute \\(d = %d\\) and \\(n = %d\\); because %d does not divide %d the answer is not %d.\n"+
				"Subgoal 3: Evaluate: \\(\\gcd(%d, %d) = %d\\).", d, d, n, d, n, d, d, n, ans),
		}
	case 1:
		// The characteristic of a product of rings is the lcm of the characteristics.
		m := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		n := m + 1
		for mathutil.GCD(m, n) != 1 {
			n++
		}
		ans := m * n
		return generator.Problem{
			Question: fmt.Sprintf("What is the characteristic of the ring \\(\\mathbb{Z}_{%d} \\times \\mathbb{Z}_{%d}\\)?", m, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The characteristic of a product is the least common multiple of the characteristics of its factors.\n"+
				"Subgoal 2: The factors have characteristic %d and %d, and \\(\\gcd(%d, %d) = 1\\).\n"+
				"Subgoal 3: Evaluate: \\(\\operatorname{lcm}(%d, %d) = %d\\).", m, n, m, n, m, n, ans),
		}
	case 2:
		// Z[x]/(x^2 + 1, p) is F_p[x]/(x^2 + 1), so it has p^2 elements.
		primes := []int{2, 3, 5}
		p := primes[rand.Intn(len(primes))]
		ans := p * p
		return generator.Problem{
			Question: fmt.Sprintf("How many elements belong to the ring \\(\\mathbb{Z}[x] / (x^{2} + 1, %d)\\)?", p),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: Reducing modulo %d sends \\(\\mathbb{Z}[x]\\) to \\(\\mathbb{F}_{%d}[x]\\).\n"+
				"Subgoal 2: The quotient by the monic degree-two polynomial \\(x^2 + 1\\) is a two-dimensional vector space over \\(\\mathbb{F}_{%d}\\).\n"+
				"Subgoal 3: Evaluate: \\(%d^{2} = %d\\) elements.", p, p, p, p, ans),
		}
	case 3:
		// F_p[x]/(f) with deg f = d has p^d elements.
		primes := []int{2, 3, 5}
		p := primes[rand.Intn(len(primes))]
		d := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		ans := 1
		for i := 0; i < d; i++ {
			ans *= p
		}
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(f(x)\\) be an irreducible polynomial of degree %d over \\(\\mathbb{Z}_{%d}\\). How many elements belong to \\(\\mathbb{Z}_{%d}[x] / (f(x))\\)?", d, p, p),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The quotient of \\(\\mathbb{F}_p[x]\\) by an irreducible polynomial of degree \\(d\\) is a field of order \\(p^d\\).\n"+
				"Subgoal 2: Substitute \\(p = %d\\) and \\(d = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(%d^{%d} = %d\\).", p, d, p, d, ans),
		}
	default:
		// The coset m + (d) lies in Z_n/(d) = Z_d, so its additive order is d/gcd(m, d).
		d := mathutil.ScaleInt(ctx.Difficulty, 3, 12)
		m := mathutil.ScaleInt(ctx.Difficulty, 1, 9)
		n := d * mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		ans := d / mathutil.GCD(m, d)
		return generator.Problem{
			Question: fmt.Sprintf("In the quotient ring \\(\\mathbb{Z}_{%d} / (%d)\\), what is the additive order of the coset \\(%d + (%d)\\)?", n, d, m, d),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The quotient is isomorphic to \\(\\mathbb{Z}_{%d}\\), where the coset corresponds to the residue %d.\n"+
				"Subgoal 2: The additive order of a residue is \\(\\frac{d}{\\gcd(m, d)} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d.", d, m, d, mathutil.GCD(m, d), ans),
		}
	}
}

// ----- 12. normal subgroup -----

type normalSubgroupGen struct{}

func (g *normalSubgroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		// |A_n| = n!/2. A_n varies with n, so this is not a constant answer.
		n := 4 + rand.Intn(3)
		fact := 1
		for i := 2; i <= n; i++ {
			fact *= i
		}
		ans := fact / 2
		return generator.Problem{
			Question: fmt.Sprintf("What is the order of the alternating group \\(A_{%d}\\)?", n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: \\(A_n\\) is the kernel of the sign homomorphism, so \\([S_n : A_n] = 2\\).\n"+
				"Subgoal 2: \\(|S_{%d}| = %d! = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(|A_{%d}| = \\frac{%d}{2} = %d\\).", n, n, fact, n, fact, ans),
		}
	case 1:
		// Orbit-stabiliser: the number of conjugates of H is |G| / |N_G(H)|.
		cOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		conj := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		gOrder := cOrder * conj
		return generator.Problem{
			Question: fmt.Sprintf("A subgroup \\(H\\) of a group \\(G\\) of order %d has normalizer \\(N_G(H)\\) of order %d. How many distinct conjugates does \\(H\\) have?", gOrder, cOrder),
			Answer:   fmt.Sprintf("%d", conj),
			Explanation: fmt.Sprintf("Subgoal 1: The conjugates of \\(H\\) are the orbit of \\(H\\) under conjugation, whose stabiliser is \\(N_G(H)\\).\n"+
				"Subgoal 2: By orbit-stabiliser, the number of conjugates is \\(\\frac{|G|}{|N_G(H)|} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d distinct conjugates.", gOrder, cOrder, conj),
		}
	case 2:
		// |N| = |G| / |G/N|
		qOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		nOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := qOrder * nOrder
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(N\\) be a normal subgroup of a group \\(G\\) of order %d. If the quotient group \\(G/N\\) has order %d, find the order of \\(N\\).", gOrder, qOrder),
			Answer:   fmt.Sprintf("%d", nOrder),
			Explanation: fmt.Sprintf("Subgoal 1: State Lagrange's relation for quotient groups: \\(|G/N| = \\frac{|G|}{|N|}\\).\n"+
				"Subgoal 2: Substitute \\(|G| = %d\\) and \\(|G/N| = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(|N| = \\frac{%d}{%d} = %d\\).", gOrder, qOrder, gOrder, qOrder, nOrder),
		}
	case 3:
		// [G : Z(G)] = |G| / |Z(G)|
		zOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		index := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := zOrder * index
		return generator.Problem{
			Question: fmt.Sprintf("A finite group \\(G\\) of order %d has center \\(Z(G)\\) of order %d. Find the index \\([G : Z(G)]\\).", gOrder, zOrder),
			Answer:   fmt.Sprintf("%d", index),
			Explanation: fmt.Sprintf("Subgoal 1: The index formula gives \\([G : Z(G)] = \\frac{|G|}{|Z(G)|}\\).\n"+
				"Subgoal 2: Substitute \\(|G| = %d\\) and \\(|Z(G)| = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", gOrder, zOrder, gOrder, zOrder, index),
		}
	default:
		// The conjugacy class of g has size |G| / |C_G(g)|.
		cOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		cls := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		gOrder := cOrder * cls
		return generator.Problem{
			Question: fmt.Sprintf("An element \\(g\\) of a group of order %d has centralizer of order %d. What is the size of its conjugacy class?", gOrder, cOrder),
			Answer:   fmt.Sprintf("%d", cls),
			Explanation: fmt.Sprintf("Subgoal 1: The conjugacy class of \\(g\\) is its orbit under conjugation, stabilised by \\(C_G(g)\\).\n"+
				"Subgoal 2: By orbit-stabiliser, the class size is \\(\\frac{|G|}{|C_G(g)|} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d.", gOrder, cOrder, cls),
		}
	}
}

// ----- 13. lagrange -----

type lagrangeGen struct{}

func (g *lagrangeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type qa struct {
		orderH string
		orderG string
		valid  string
		reason string
	}
	tableSmall := []qa{
		{"2", "6", "yes", "Lagrange: |H| divides |G|, and 2 divides 6."},
		{"3", "6", "yes", "3 divides 6, so a subgroup of order 3 can occur in a group of order 6."},
		{"3", "7", "no", "3 does not divide 7, so no subgroup of order 3 in a group of order 7."},
		{"4", "6", "no", "4 does not divide 6, violating Lagrange."},
	}
	tableLarge := []qa{
		{"12", "24", "yes", "12 divides 24."},
		{"8", "24", "yes", "8 divides 24."},
		{"5", "24", "no", "5 does not divide 24."},
		{"6", "18", "yes", "6 divides 18."},
	}
	pool := tableSmall
	if scale > 3 {
		pool = append(tableSmall, tableLarge...)
	}
	e := pool[rand.Intn(len(pool))]
	if rand.Intn(3) == 0 {
		// production: subgroup indices from Lagrange (unique answers).
		type idxEntry struct {
			orderG string
			orderH string
			index  string
			reason string
		}
		idxTable := []idxEntry{
			{"12", "4", "3", "Index [G:H] = |G|/|H| = 12/4 = 3."},
			{"12", "3", "4", "Index = 12/3 = 4."},
			{"24", "8", "3", "Index = 24/8 = 3."},
		}
		v := idxTable[rand.Intn(len(idxTable))]
		return generator.Problem{
			Question:    fmt.Sprintf("A group G has order %s and a subgroup H of order %s. What is the index [G:H]? (enter a number)", v.orderG, v.orderH),
			Answer:      v.index,
			Explanation: v.reason,
		}
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Can a group of order %s have a subgroup of order %s? (yes/no)", e.orderG, e.orderH),
		Answer:      e.valid,
		Explanation: e.reason,
	}
}

// ----- 14. quotient group -----

type quotientGroupGen struct{}

func (g *quotientGroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		// |G/N| = |G| / |N|
		nOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		qOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := nOrder * qOrder
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(G\\) be a group of order %d and let \\(N\\) be a normal subgroup of order %d. What is the order of the quotient group \\(G/N\\)?", gOrder, nOrder),
			Answer:   fmt.Sprintf("%d", qOrder),
			Explanation: fmt.Sprintf("Subgoal 1: State Lagrange's Theorem for quotient groups: \\(|G/N| = \\frac{|G|}{|N|}\\).\n"+
				"Subgoal 2: Substitute \\(|G| = %d\\) and \\(|N| = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", gOrder, nOrder, gOrder, nOrder, qOrder),
		}
	case 1:
		// |Z_n / <d>| = gcd(d, n), and d does not divide n so this is not d.
		d := mathutil.ScaleInt(ctx.Difficulty, 3, 9)
		n := mathutil.ScaleInt(ctx.Difficulty, 4, 20)
		for n%d == 0 {
			n++
		}
		ans := mathutil.GCD(d, n)
		return generator.Problem{
			Question: fmt.Sprintf("In the additive cyclic group \\(\\mathbb{Z}_{%d}\\), find the order of the quotient group \\(\\mathbb{Z}_{%d} / \\langle %d \\rangle\\).", n, n, d),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The subgroup has order \\(\\frac{n}{\\gcd(d, n)}\\), so the quotient has order \\(\\gcd(d, n)\\).\n"+
				"Subgoal 2: Substitute \\(n = %d\\) and \\(d = %d\\); note that %d does not divide %d.\n"+
				"Subgoal 3: Evaluate: \\(\\gcd(%d, %d) = %d\\).", n, d, d, n, d, n, ans),
		}
	case 2:
		// The coset m + N has order d / gcd(m, d) in Z_n / <d>.
		// Ensure gcd(m, d) > 1 so ans < d, preventing read-off of d from the stem.
		gVal := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		k := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		d := gVal * k
		m := gVal * (rand.Intn(3) + 1)
		for mathutil.GCD(m, d) != gVal {
			m = gVal * (rand.Intn(3) + 1)
		}
		n := d * mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		ans := d / mathutil.GCD(m, d)
		return generator.Problem{
			Question: fmt.Sprintf("In the quotient group \\(\\mathbb{Z}_{%d} / \\langle %d \\rangle\\), what is the order of the coset \\(%d + \\langle %d \\rangle\\)?", n, d, m, d),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The quotient is cyclic of order %d, and the coset corresponds to the residue %d.\n"+
				"Subgoal 2: The order of that residue is \\(\\frac{%d}{\\gcd(%d, %d)} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d.", d, m, d, m, d, mathutil.GCD(m, d), ans, ans),
		}
	case 3:
		// First Isomorphism Theorem: |G / ker(phi)| = |im(phi)| = |H|
		kOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		hOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := kOrder * hOrder
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(\\phi: G \\to H\\) be a surjective group homomorphism from a group of order %d. If \\(|\\ker(\\phi)| = %d\\), find the order of \\(H\\).", gOrder, kOrder),
			Answer:   fmt.Sprintf("%d", hOrder),
			Explanation: fmt.Sprintf("Subgoal 1: State the First Isomorphism Theorem: \\(G / \\ker(\\phi) \\cong \\operatorname{im}(\\phi) = H\\).\n"+
				"Subgoal 2: Equate orders: \\(|H| = \\frac{|G|}{|\\ker(\\phi)|}\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{%d} = %d\\).", gOrder, kOrder, hOrder),
		}
	default:
		// |A_n| = n!/2, which varies with n.
		n := 4 + rand.Intn(3)
		fact := 1
		for i := 2; i <= n; i++ {
			fact *= i
		}
		ans := fact / 2
		return generator.Problem{
			Question: fmt.Sprintf("The alternating group \\(A_{%d}\\) is a normal subgroup of \\(S_{%d}\\). What is the order of \\(A_{%d}\\)?", n, n, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: \\(|S_{%d}| = %d! = %d\\).\n"+
				"Subgoal 2: \\(A_{%d}\\) is the kernel of the sign homomorphism, so it has half the elements of \\(S_{%d}\\).\n"+
				"Subgoal 3: Evaluate: \\(\\frac{%d}{2} = %d\\).", n, n, fact, n, n, fact, ans),
		}
	}
}

// ----- 15. polynomial ring -----

type polynomialRingGen struct{}

func (g *polynomialRingGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	tableEasy := []entry{
		{"What ring is R[x] / (x) isomorphic to?", "R", "Quotient by the ideal (x) evaluates polynomials at 0, leaving constants R."},
		{"What ring is Z[x] / (x) isomorphic to?", "Z", "Evaluating at 0 kills (x), leaving constants Z."},
		{"What field is R[x] / (x^2+1) isomorphic to?", "C", "x^2+1 is irreducible over R, so the quotient is a field extension C."},
		{"In Q[x]/(x^2-2), what does the class of x square to?", "2", "x^2 = 2 in the quotient, so the class behaves as sqrt(2)."},
		{"Is Q[x] a PID? (yes/no)", "yes", "Polynomial ring over a field is a PID."},
		{"Is Z[x] a PID? (yes/no)", "no", "The ideal (2,x) is not principal, so Z[x] is not a PID."},
	}
	tableHard := []entry{
		{"Is R[x]/(x^2) a field? (yes/no)", "no", "x^2 is not irreducible (has repeated root), quotient has zero divisors, so not a field."},
		{"In R[x]/(x^2), the class of x is nonzero but squares to zero. Name it (the zero divisor).", "x", "The class of x is nonzero (x is not a multiple of x^2) and x*x = x^2 = 0 in the quotient."},
		{"What is the minimal polynomial of sqrt(2) over Q? (type like x^2-2)", "x^2-2", "sqrt(2) satisfies x^2-2=0, irreducible over Q by Eisenstein/rational root test."},
		{"What is Q[x]/(x^2-2) isomorphic to? (type like Q(sqrt(2)))", "Q(sqrt(2))", "Adjoining a root of the irreducible x^2-2 gives Q(sqrt(2))."},
		{"Does Z[x] contain Q? (yes/no)", "no", "Z[x] has integer coefficients only; Q requires rational coefficients."},
		{"Is the ideal (2,x) in Z[x] maximal? (yes/no)", "yes", "Quotient Z[x]/(2,x) ≅ Z_2, a field, so maximal."},
	}
	pool := tableEasy
	if scale > 3 {
		pool = append(tableEasy, tableHard...)
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.question, e.answer, e.reason)
}

// ----- 16. group action -----

type groupActionGen struct{}

func (g *groupActionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Does a group action of G on X partition X into orbits? (yes/no)", "yes", "Orbits are equivalence classes under x~y iff gx=y for some g."},
		{"Does orbit-stabilizer give |G| = |Orb(x)|·|Stab(x)|? (yes/no)", "yes", "The bijection G/Stab(x) ≅ Orb(x) gives the formula."},
		{"Is action of S_n on {1..n} transitive? (yes/no)", "yes", "Any i can be sent to any j by a permutation."},
		{"Does trivial action have all orbits of size 1? (yes/no)", "yes", "gx=x for all g gives orbits {x}."},
		{"Do different orbits of one action always have the same size? (yes/no)", "no", "Orbit sizes divide |G| but vary: fixed points have size 1, others larger."},
	}
	hard := []entry{
		{"Does conjugation action of G on itself have orbits = conjugacy classes? (yes/no)", "yes", "gxg^{-1} orbits are conjugacy classes."},
		{"Is Burnside's lemma |X/G| = (1/|G|)∑_g|Fix(g)| true? (yes/no)", "yes", "Counts orbits by averaging fixed points."},
		{"Does action of D_4 on vertices of square have orbit size 4? (yes/no)", "yes", "All 4 vertices are equivalent under dihedral symmetries."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: orbit-stabilizer arithmetic (unique answers).
		type orbEntry struct {
			question string
			answer   string
			reason   string
		}
		orbTable := []orbEntry{
			{"G acts transitively on 5 points and |Stab(x)|=3. What is |G|? (enter a number)", "15", "Orbit-stabilizer: |G| = 5·3 = 15."},
			{"D_4 acts on the 4 vertices in one orbit. If |Stab|=2, what is |D_4|? (enter a number)", "8", "Orbit-stabilizer: 4·2 = 8."},
			{"S_3 acts on {1,2,3} transitively. The stabilizer of 3 has order 2. What is |S_3|? (enter a number)", "6", "Orbit-stabilizer: 3·2 = 6."},
		}
		o := orbTable[rand.Intn(len(orbTable))]
		return generator.Problem{Question: o.question, Answer: o.answer, Explanation: o.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 17. sylow -----

// ----- 17. sylow -----

type sylowGen struct{}

func (g *sylowGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Does a group of order 12=2^2·3 have a Sylow 2-subgroup of order 4? (yes/no)", "yes", "Sylow existence: subgroup of order p^k where p^k||G|."},
		{"Does n_p ≡ 1 mod p for Sylow p-subgroups? (yes/no)", "yes", "Congruence condition: number of Sylow p-subgroups ≡1 mod p."},
		{"Does Sylow guarantee n_p divides |G|? (yes/no)", "yes", "n_p divides |G|/p^k."},
	}
	hard := []entry{
		{"If |G|=12, must n_3 divide 4 and ≡1 mod 3 so n_3∈{1,4}? (yes/no)", "yes", "n_3 divides 4 and ≡1 mod 3."},
		{"Does a group of order 15=3·5 have a normal Sylow 5-subgroup? (yes/no)", "yes", "n_5 divides 3 and ≡1 mod5 → n_5=1."},
		{"If n_p=1, is the Sylow p-subgroup normal? (yes/no)", "yes", "Unique Sylow is characteristic, hence normal."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: pin down n_p from the divisibility + congruence constraints.
		type sylEntry struct {
			question string
			answer   string
			reason   string
		}
		sylTable := []sylEntry{
			{"|G|=15=3·5. n_5 divides 3 and is 1 mod 5. What is n_5? (enter a number)", "1", "Divisors of 3: 1, 3. Only 1 is 1 mod 5."},
			{"|G|=21=3·7. n_7 divides 3 and is 1 mod 7. What is n_7? (enter a number)", "1", "Divisors of 3: 1, 3. Only 1 is 1 mod 7."},
			{"|G|=12. n_3 divides 4 and is 1 mod 3. How many values can n_3 take? (enter a number)", "2", "n_3 ∈ {1, 4}: two possibilities."},
		}
		s := sylTable[rand.Intn(len(sylTable))]
		return generator.Problem{Question: s.question, Answer: s.answer, Explanation: s.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

// ----- 18. UFD -----

type ufdGen struct{}

func (g *ufdGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is Z a UFD? (yes/no)", "yes", "Integers have unique prime factorization."},
		{"Is Z[√-5] a UFD? (yes/no)", "no", "6=2·3=(1+√-5)(1-√-5) gives non-unique factorization."},
		{"Does PID imply UFD? (yes/no)", "yes", "Every PID is a UFD."},
		{"Is Q[x] a UFD? (yes/no)", "yes", "Polynomial ring over a field is a UFD."},
	}
	hard := []entry{
		{"Is Gauss's lemma used to prove Z[x] is a UFD if Z is? (yes/no)", "yes", "Content and primitive polynomials give UFD lift."},
		{"Does UFD require every irreducible is prime? (yes/no)", "yes", "Irreducible ⇒ prime characterizes UFDs among domains."},
		{"Is R[x] with R a UFD itself a UFD? (yes/no)", "yes", "Polynomial ring over a UFD is a UFD."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: count factors / factorizations (unique answers).
		type ufdEntry struct {
			question string
			answer   string
			reason   string
		}
		ufdTable := []ufdEntry{
			{"How many prime factors (with multiplicity) does 12 have in Z? (enter a number)", "3", "12 = 2·2·3: three prime factors."},
			{"In 6 = 2·3 = (1+√-5)(1-√-5), how many distinct factorizations are shown? (enter a number)", "2", "Two genuinely different factorizations: witnesses non-uniqueness."},
			{"x^2-1 factors over Q. How many linear factors? (enter a number)", "2", "x^2-1 = (x-1)(x+1): two linear factors."},
		}
		u := ufdTable[rand.Intn(len(ufdTable))]
		return groupChoice(u.question, u.answer, u.reason, choicesFor(u.answer)...)
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

// ----- 19. field extension -----

// ----- 20. Galois -----

type localizationGen struct{}

func (g *localizationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is localization S^{-1}R inverting elements of S? (yes/no)", "yes", "Fractions r/s with s∈S."},
		{"Is R localized at prime ideal a local ring? (yes/no)", "yes", "R_p has unique maximal ideal."},
		{"Does localization at S={1,2,4,8,...} give Z[1/2]? (yes/no)", "yes", "Invert powers of 2."},
	}
	hard := []entry{
		{"Is field of fractions of integral domain localization at nonzero elements? (yes/no)", "yes", "S=R\\{0}."},
		{"Does localization preserve exactness? (yes/no)", "yes", "Localization is exact functor."},
		{"Is Z_{(p)} localization at complement of (p) local? (yes/no)", "yes", "Local ring at p."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: fractions and maximal-ideal counts (unique answers).
		type locEntry struct {
			question string
			answer   string
			reason   string
		}
		locTable := []locEntry{
			{"In Z[1/2], what is 1/2+1/4? (enter a fraction like 3/4)", "3/4", "Common denominator 4: 2/4+1/4=3/4."},
			{"How many maximal ideals does the local ring Z_(p) have? (enter a number)", "1", "Local means unique maximal ideal."},
			{"In Z[1/2], what is 2/4 reduced? (enter a fraction like 1/2)", "1/2", "Cancel 2: 2/4 = 1/2."},
		}
		v := locTable[rand.Intn(len(locTable))]
		return groupChoice(v.question, v.answer, v.reason, choicesFor(v.answer)...)
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

type finiteFieldGen struct{}

func (g *finiteFieldGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is F_p field with p prime? (yes/no)", "yes", "Z/pZ is field iff p prime."},
		{"Does finite field F_{p^n} have p^n elements? (yes/no)", "yes", "Extension of degree n over F_p."},
		{"Is multiplicative group of finite field cyclic? (yes/no)", "yes", "F_{p^n}^× cyclic."},
	}
	hard := []entry{
		{"Is there a field with 4 elements? (yes/no)", "yes", "F_4 = F_2[x]/(x^2+x+1)."},
		{"Is there a field with 6 elements? (yes/no)", "no", "Order must be prime power."},
		{"Does Frobenius x→x^p give automorphism of F_{p^n}? (yes/no)", "yes", "Frobenius."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: orders and degrees in finite fields (unique answers).
		type finEntry struct {
			question string
			answer   string
			reason   string
		}
		finTable := []finEntry{
			{"How many elements does F_8 have? (enter a number)", "8", "F_8 is degree 3 over F_2: 2^3 = 8."},
			{"F_4^× is cyclic. What is its order? (enter a number)", "3", "4-1 = 3 nonzero elements."},
			{"A field of order 9 is degree n over F_3. What is n? (enter a number)", "2", "9 = 3^2, so n = 2."},
		}
		v := finTable[rand.Intn(len(finTable))]
		return generator.Problem{Question: v.question, Answer: v.answer, Explanation: v.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

type nilpotentGen struct{}

func (g *nilpotentGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is p-group nilpotent? (yes/no)", "yes", "Finite p-groups are nilpotent."},
		{"Is nilpotent ⇒ solvable? (yes/no)", "yes", "Nilpotent groups are solvable."},
		{"Does direct product of p-groups give nilpotent? (yes/no)", "yes", "Product of its Sylow subgroups."},
	}
	hard := []entry{
		{"Is upper unitriangular group nilpotent? (yes/no)", "yes", "Heisenberg group nilpotent."},
		{"Is S_3 nilpotent? (yes/no)", "no", "S_3 not nilpotent (nor its Sylow structure)."},
		{"Does nilpotent have central series reaching G? (yes/no)", "yes", "Upper central series."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: Sylow counts and nilpotency classes (unique answers).
		type nilEntry struct {
			question string
			answer   string
			reason   string
		}
		nilTable := []nilEntry{
			{"A nilpotent group of order 12=2^2·3 has unique Sylow subgroups. How many Sylow subgroups total? (enter a number)", "2", "One Sylow 2-subgroup and one Sylow 3-subgroup."},
			{"A_3 is abelian, hence nilpotent of class 1. What is its class? (enter a number)", "1", "Abelian groups have trivial commutators: class 1."},
			{"A group of order 8=2^3 is a 2-group. What prime? (enter a number)", "2", "8 = 2^3: it is a 2-group, hence nilpotent."},
		}
		v := nilTable[rand.Intn(len(nilTable))]
		return generator.Problem{Question: v.question, Answer: v.answer, Explanation: v.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

type solvableGen struct{}

func (g *solvableGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is solvable group one with subnormal series with abelian quotients? (yes/no)", "yes", "Derived series reaches 1."},
		{"Is S_3 solvable? (yes/no)", "yes", "S_3 has series 1◃A_3◃S_3 with abelian quotients."},
		{"Is S_5 solvable? (yes/no)", "no", "S_n for n≥5 non-solvable."},
	}
	hard := []entry{
		{"Does Galois solvability correspond to group solvability? (yes/no)", "yes", "Polynomial solvable by radicals iff Galois group solvable."},
		{"Is every p-group solvable? (yes/no)", "yes", "p-groups nilpotent ⇒ solvable."},
		{"Does composition series refine solvable series? (yes/no)", "yes", "Jordan-Hölder."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: orders inside solvable series (unique answers).
		type solEntry struct {
			question string
			answer   string
			reason   string
		}
		solTable := []solEntry{
			{"S_3 has series 1◃A_3◃S_3. What is |A_3|? (enter a number)", "3", "A_3 has the 3 even permutations."},
			{"S_n is non-solvable for n≥5. What is the smallest such n? (enter a number)", "5", "S_5 is the smallest non-solvable symmetric group."},
			{"Derived series of S_3: S_3' = A_3, A_3' = 1. How many steps to reach 1? (enter a number)", "2", "Two derived subgroups: S_3 → A_3 → 1."},
		}
		v := solTable[rand.Intn(len(solTable))]
		return generator.Problem{Question: v.question, Answer: v.answer, Explanation: v.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

type freeGroupGen struct{}

func (g *freeGroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is free group on 1 generator isomorphic to Z? (yes/no)", "yes", "F_1 ≅ Z."},
		{"Does free group have presentation with no relations? (yes/no)", "yes", "Free = no relations."},
		{"Is every group quotient of a free group? (yes/no)", "yes", "Free groups are free objects."},
	}
	hard := []entry{
		{"Is free group on 2 generators non-abelian? (yes/no)", "yes", "F_2 non-abelian."},
		{"Does Nielsen-Schreier state subgroups of free are free? (yes/no)", "yes", "Subgroups of free are free."},
		{"Is free product Z * Z = F_2? (yes/no)", "yes", "Free product of Z's."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: ranks and generator counts (unique answers).
		type freeEntry struct {
			question string
			answer   string
			reason   string
		}
		freeTable := []freeEntry{
			{"F_1 ≅ Z. What is the rank of F_1? (enter a number)", "1", "One generator: rank 1."},
			{"F_n/[F_n,F_n] ≅ Z^n. For n=3, what is the rank? (enter a number)", "3", "Abelianization of F_3 is Z^3."},
			{"Z * Z ≅ F_2. How many generators? (enter a number)", "2", "Free product of two Z's: two generators."},
		}
		v := freeTable[rand.Intn(len(freeTable))]
		return generator.Problem{Question: v.question, Answer: v.answer, Explanation: v.reason}
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

type noetherianGen struct{}

func (g *noetherianGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		q string
		a string
		e string
	}
	easy := []entry{
		{"Is Z Noetherian? (yes/no)", "yes", "PID ⇒ Noetherian, ACC on ideals."},
		{"Does Hilbert basis say R Noetherian ⇒ R[x] Noetherian? (yes/no)", "yes", "Hilbert basis theorem."},
		{"Is field Noetherian? (yes/no)", "yes", "Field has only 0 and itself."},
	}
	hard := []entry{
		{"Is k[x1,x2,...] infinite variables non-Noetherian? (yes/no)", "yes", "Chain (x1) ⊂ (x1,x2) ⊂ ... never stabilizes."},
		{"Are Noetherian rings closed under quotient? (yes/no)", "yes", "Quotient of Noetherian is Noetherian."},
		{"Is localization of Noetherian Noetherian? (yes/no)", "yes", "Localization preserves Noetherian."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	if rand.Intn(3) == 0 {
		// production: ideal counts and principal generators (unique answers).
		type noeEntry struct {
			question string
			answer   string
			reason   string
		}
		noeTable := []noeEntry{
			{"A field has exactly how many ideals? (enter a number)", "2", "Only 0 and the field itself."},
			{"Ideals of Z/12Z correspond to divisors of 12. How many divisors does 12 have? (enter a number)", "6", "1, 2, 3, 4, 6, 12: six divisors, six ideals."},
			{"In the PID Z, the ideal (6,10) equals (d) for d=gcd(6,10). What is d? (enter a number)", "2", "gcd(6,10) = 2, so (6,10) = (2)."},
		}
		v := noeTable[rand.Intn(len(noeTable))]
		return groupChoice(v.question, v.answer, v.reason, choicesFor(v.answer)...)
	}
	e := pool[rand.Intn(len(pool))]
	return groupYesNo(e.q, e.a, e.e)
}

type commutatorGen struct{}

func (g *commutatorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		// [S_n, S_n] = A_n, so the derived subgroup has order n!/2.
		n := 3 + rand.Intn(4) // 3 to 6
		fact := 1
		for i := 2; i <= n; i++ {
			fact *= i
		}
		ans := fact / 2
		return generator.Problem{
			Question: fmt.Sprintf("What is the order of the commutator subgroup \\([S_{%d}, S_{%d}]\\)?", n, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The derived subgroup of \\(S_n\\) is \\(A_n\\) for \\(n \\ge 3\\).\n"+
				"Subgoal 2: Substitute \\(|S_{%d}| = %d! = %d\\).\n"+
				"Subgoal 3: Evaluate: \\(|A_{%d}| = \\frac{%d}{2} = %d\\).", n, n, fact, n, fact, ans),
		}
	case 1:
		// Dihedral group commutator subgroup [D_{2k}, D_{2k}] = <r^2> has order k.
		// Use even n = 2k with k in [3, 9] so the answer k is never printed in the stem.
		k := mathutil.ScaleInt(ctx.Difficulty, 3, 9)
		n := 2 * k
		ans := k
		return generator.Problem{
			Question: fmt.Sprintf("What is the order of the commutator subgroup \\([D_{%d}, D_{%d}]\\) of the dihedral group \\(D_{%d}\\)?", n, n, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: In the dihedral group \\(D_n\\) of symmetries of a regular \\(n\\)-gon, the commutator subgroup is generated by the square of the rotation: \\([D_n, D_n] = \\langle r^2 \\rangle\\).\n"+
				"Subgoal 2: The rotation \\(r\\) has order %d. Since \\(n = 2(%d)\\) is even, the order of \\(r^2\\) is \\(\\frac{%d}{2} = %d\\).\n"+
				"Subgoal 3: Evaluate: %d.", n, k, n, k, ans),
		}
	case 2:
		// Order of the abelianization G / [G, G] given |G| and |[G, G]|
		derOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
		abOrder := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		gOrder := derOrder * abOrder
		return generator.Problem{
			Question: fmt.Sprintf("A finite group \\(G\\) of order %d has commutator subgroup \\([G, G]\\) of order %d. Find the order of the abelianization \\(G / [G, G]\\).", gOrder, derOrder),
			Answer:   fmt.Sprintf("%d", abOrder),
			Explanation: fmt.Sprintf("Subgoal 1: The abelianization is \\(G / [G, G]\\).\n"+
				"Subgoal 2: By Lagrange's Theorem, \\(|G / [G, G]| = \\frac{|G|}{|[G, G]|} = \\frac{%d}{%d}\\).\n"+
				"Subgoal 3: Evaluate: %d.", gOrder, derOrder, abOrder),
		}
	case 3:
		// Matrix group commutator: [GL_2(F_p), GL_2(F_p)] = SL_2(F_p) of order p(p^2 - 1)
		primes := []int{3, 5}
		p := primes[rand.Intn(len(primes))]
		ans := p * (p*p - 1)
		return generator.Problem{
			Question: fmt.Sprintf("For the finite field \\(\\mathbb{F}_{%d}\\), what is the order of the commutator subgroup \\([\\operatorname{GL}_2(\\mathbb{F}_{%d}), \\operatorname{GL}_2(\\mathbb{F}_{%d})]\\)?", p, p, p),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The commutator subgroup of the general linear group \\(\\operatorname{GL}_2(\\mathbb{F}_p)\\) is the special linear group \\(\\operatorname{SL}_2(\\mathbb{F}_p)\\) consisting of matrices with determinant 1.\n"+
				"Subgoal 2: The order of \\(\\operatorname{SL}_2(\\mathbb{F}_p)\\) is \\(p(p^2 - 1)\\).\n"+
				"Subgoal 3: Evaluate for \\(p = %d\\): \\(%d(%d^2 - 1) = %d(%d) = %d\\).", p, p, p, p, p*p-1, ans),
		}
	case 4:
		// Quaternion group Q_8 commutator subgroup has order 2, abelianization has order 4
		sub := rand.Intn(2)
		if sub == 0 {
			return generator.Problem{
				Question: "What is the order of the commutator subgroup \\([Q_8, Q_8]\\) of the quaternion group \\(Q_8\\)?",
				Answer:   "2",
				Explanation: "Subgoal 1: In the quaternion group \\(Q_8\\), the commutator of any two non-commuting elements is \\([i, j] = i j i^{-1} j^{-1} = -1\\).\n" +
					"Subgoal 2: The commutator subgroup is the center \\(Z(Q_8) = \\{1, -1\\}\\).\n" +
					"Subgoal 3: Evaluate: the order is 2.",
			}
		} else {
			return generator.Problem{
				Question: "What is the order of the abelianization \\(Q_8 / [Q_8, Q_8]\\) of the quaternion group \\(Q_8\\)?",
				Answer:   "4",
				Explanation: "Subgoal 1: The abelianization is the quotient \\(Q_8 / [Q_8, Q_8]\\).\n" +
					"Subgoal 2: The quaternion group has order 8 and its commutator subgroup has order 2.\n" +
					"Subgoal 3: Evaluate: \\(\\frac{8}{2} = 4\\) (isomorphic to the Klein four-group \\(V_4\\)).",
			}
		}
	default:
		// (S_n x Z_m)^ab = Z_2 x Z_m, of order 2m.
		m := mathutil.ScaleInt(ctx.Difficulty, 2, 9)
		ans := 2 * m
		return generator.Problem{
			Question: fmt.Sprintf("What is the order of the abelianization of \\(S_3 \\times \\mathbb{Z}_{%d}\\)?", m),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: The abelianization of a product is the product of the abelianizations, and \\(S_3^{ab} \\cong \\mathbb{Z}_2\\).\n"+
				"Subgoal 2: The abelianization is \\(\\mathbb{Z}_2 \\times \\mathbb{Z}_{%d}\\), which has \\(2 \\cdot %d\\) elements.\n"+
				"Subgoal 3: Evaluate: %d.", m, m, ans),
		}
	}
}

// ----- symmetric group -----

type symmetricGroupGen struct{}

func (g *symmetricGroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	tableEasy := []entry{
		{"What is the order of S_3? (enter a number)", "6", "3! = 6 permutations."},
		{"What is the order of S_4? (enter a number)", "24", "4! = 24."},
		{"What is the order of (1 2 3) in S_3? (enter a number)", "3", "3-cycles have order 3."},
		{"Is S_3 abelian? (yes/no)", "no", "σ∘τ ≠ τ∘σ in general: (1 3) vs (2 3)."},
		{"Does every permutation decompose into disjoint cycles? (yes/no)", "yes", "Follow orbits until they close; decomposition is unique up to order."},
	}
	tableHard := []entry{
		{"What is |A_4|? (enter a number)", "12", "Index 2 in S_4: 24/2 = 12."},
		{"How many conjugacy classes does S_4 have? (enter a number)", "5", "Five partitions of 4: identity, transpositions, double transpositions, 3-cycles, 4-cycles."},
		{"Is S_n non-abelian for n≥3? (yes/no)", "yes", "Embed the S_3 pair fixing 4..n."},
	}
	pool := tableEasy
	if scale > 3 {
		pool = append(tableEasy, tableHard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- dihedral group -----

type dihedralGroupGen struct{}

func (g *dihedralGroupGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	tableEasy := []entry{
		{"What is the order of D_3? (enter a number)", "6", "3 rotations + 3 reflections; D_3 ≅ S_3."},
		{"What is the order of D_4? (enter a number)", "8", "4 rotations + 4 reflections."},
		{"In D_5, what is the order of the rotation r? (enter a number)", "5", "Rotations form a cyclic subgroup of order n."},
		{"Is D_3 isomorphic to S_3? (yes/no)", "yes", "Both order 6 with the same structure: rotations as 3-cycles, reflections as transpositions."},
		{"Do all reflections in D_n have order 2? (yes/no)", "yes", "Each reflection is an involution."},
		{"Is D_4 abelian? (yes/no)", "no", "Rotations and reflections do not commute: sr = r^-1 s."},
	}
	tableHard := []entry{
		{"How many central elements does D_4 have? (enter a number)", "2", "Identity plus the half-turn r^2."},
		{"How many subgroups does D_6 have? (enter a number)", "16", "τ(6) + σ(6) = 4 + 12."},
		{"In D_n, srs equals what power expression? (type like r^-1)", "r^-1", "Moving a rotation across a reflection inverts it."},
	}
	pool := tableEasy
	if scale > 3 {
		pool = append(tableEasy, tableHard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}
