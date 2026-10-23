package calculus

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerVector3D(reg *generator.Registry) {
	reg.Register("calc.vec.divergence_theorem", &divergenceTheoremGen{})
	reg.Register("calc.vec.flux_uniform", &fluxUniformGen{})
	reg.Register("calc.vec.potential_value", &potentialValueGen{})
	reg.Register("calc.multi.coordinates_jacobian", &coordinatesJacobianGen{})
}

// divergenceTheoremGen turns a flux through a box into a volume integral of the divergence, which
// for a linear field is constant.
type divergenceTheoremGen struct{}

func (g *divergenceTheoremGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, p, q, r, divVal, vol, flux int
	var qText, ex string
	for {
		maxBound := 2 + tier*2
		a = small(1, maxBound)
		b = small(1, maxBound)
		c = small(1, maxBound)
		p = small(2, maxBound+1)
		q = small(2, maxBound+1)
		r = small(2, maxBound+1)
		divVal = a + b + c
		vol = p * q * r
		flux = divVal * vol
		ansStr := strconv.Itoa(flux)
		qText = fmt.Sprintf("Use the divergence theorem to find the outward flux of \\(\\mathbf{F} = (%dx,\\ %dy,\\ %dz)\\) through the boundary of the box \\(0 \\le x \\le %d\\), \\(0 \\le y \\le %d\\), \\(0 \\le z \\le %d\\).", a, b, c, p, q, r)
		if flux == a || flux == b || flux == c || flux == p || flux == q || flux == r || containsWord(qText, ansStr) {
			continue
		}
		ex = fmt.Sprintf("By the Divergence Theorem, the outward flux equals the volume integral of the divergence: \\(\\iint_{\\partial V} \\mathbf{F} \\cdot d\\mathbf{S} = \\iiint_V (\\nabla \\cdot \\mathbf{F})\\,dV\\). Calculate the divergence: \\(\\nabla \\cdot \\mathbf{F} = \\frac{\\partial}{\\partial x}(%dx) + \\frac{\\partial}{\\partial y}(%dy) + \\frac{\\partial}{\\partial z}(%dz) = %d + %d + %d = %d\\). Multiply constant divergence by the box volume \\(V = %d \\times %d \\times %d = %d\\) to obtain \\(\\text{Flux} = %d \\times %d = %d\\).",
			a, b, c, a, b, c, divVal, p, q, r, vol, divVal, vol, flux)
		break
	}
	return generator.Problem{
		Question:    qText,
		Answer:      strconv.Itoa(flux),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"c":      strconv.Itoa(c),
			"p":      strconv.Itoa(p),
			"q":      strconv.Itoa(q),
			"r":      strconv.Itoa(r),
			"flux":   strconv.Itoa(flux),
			"answer": strconv.Itoa(flux),
		},
	}
}

// fluxUniformGen is the same flux without the theorem, for a field that only pushes through one
// pair of faces, so the answer is the field strength times the face area.
type fluxUniformGen struct{}

func (g *fluxUniformGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var c, p, q, areaVal, flux int
	var qText, ex string
	for {
		maxBound := 3 + tier*2
		c = small(2, maxBound+1)
		p = small(2, maxBound)
		q = small(2, maxBound)
		areaVal = p * q
		flux = c * areaVal
		ansStr := strconv.Itoa(flux)
		qText = fmt.Sprintf("Find the flux of the constant field \\(\\mathbf{F} = (0, 0, %d)\\) through the rectangle \\(0 \\le x \\le %d\\), \\(0 \\le y \\le %d\\) in the plane \\(z = 0\\), oriented upward.", c, p, q)
		if flux == c || flux == p || flux == q || containsWord(qText, ansStr) {
			continue
		}
		ex = fmt.Sprintf("Surface flux is defined by the surface integral \\(\\iint_S \\mathbf{F} \\cdot d\\mathbf{S} = \\iint_S (\\mathbf{F} \\cdot \\mathbf{n})\\,dA\\). For the upward-oriented plane \\(z = 0\\), the unit normal is \\(\\mathbf{n} = (0, 0, 1)\\), yielding integrand \\(\\mathbf{F} \\cdot \\mathbf{n} = (0, 0, %d) \\cdot (0, 0, 1) = %d\\). Multiply constant normal flux by the rectangle area \\(\\text{Area} = %d \\times %d = %d\\) to find \\(\\text{Flux} = %d \\times %d = %d\\).",
			c, c, p, q, areaVal, c, areaVal, flux)
		break
	}
	return generator.Problem{
		Question:    qText,
		Answer:      strconv.Itoa(flux),
		Explanation: ex,
		Facts: map[string]string{
			"c":      strconv.Itoa(c),
			"p":      strconv.Itoa(p),
			"q":      strconv.Itoa(q),
			"flux":   strconv.Itoa(flux),
			"answer": strconv.Itoa(flux),
		},
	}
}

// potentialValueGen recovers a potential and evaluates it, which avoids asking for a symbolic
// antiderivative while still testing that the potential is the thing computed.
type potentialValueGen struct{}

func (g *potentialValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, p, q, val int
	var qText, ex string
	for {
		maxBound := 3 + tier*2
		a = small(1, maxBound)
		p = small(2, maxBound+1)
		q = small(2, maxBound+1)
		val = a * (p*p + q*q)
		ansStr := strconv.Itoa(val)
		qText = fmt.Sprintf("The field \\(\\mathbf{F} = (%dx,\\ %dy)\\) is conservative. Its potential \\(f\\) satisfies \\(f(0,0) = 0\\); evaluate \\(f(%d, %d)\\).", 2*a, 2*a, p, q)
		if val == 2*a || val == p || val == q || containsWord(qText, ansStr) {
			continue
		}
		ex = fmt.Sprintf("For a conservative field, \\(\\nabla f = (\\partial f/\\partial x, \\partial f/\\partial y) = (%dx, %dy)\\). Integrate partial derivatives: \\(f(x,y) = \\int %dx\\,dx = %dx^2 + g(y)\\) and \\(\\partial f/\\partial y = %dy\\) confirms \\(f(x,y) = %dx^2 + %dy^2 + C\\). Using \\(f(0,0) = 0\\) sets \\(C = 0\\); evaluate at \\((%d, %d)\\) to obtain \\(f(%d, %d) = %d(%d)^2 + %d(%d)^2 = %d(%d) + %d(%d) = %d\\).",
			2*a, 2*a, 2*a, a, 2*a, a, a, p, q, p, q, a, p, a, q, a, p*p, a, q*q, val)
		break
	}
	return generator.Problem{
		Question:    qText,
		Answer:      strconv.Itoa(val),
		Explanation: ex,
		Facts: map[string]string{
			"coeff":  strconv.Itoa(2 * a),
			"p":      strconv.Itoa(p),
			"q":      strconv.Itoa(q),
			"value":  strconv.Itoa(val),
			"answer": strconv.Itoa(val),
		},
	}
}

// coordinatesJacobianGen evaluates the volume element of a three-dimensional coordinate system.
type coordinatesJacobianGen struct{}

func (g *coordinatesJacobianGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var factor, rho, a, r int
	var phi, qText, ex string
	for {
		if rand.Intn(2) == 0 {
			// Scaled cylindrical coordinates: (x, y, z) = (a r cos theta, a r sin theta, z)
			// Jacobian determinant is a^2 r
			a = small(2, 2+tier)
			r = small(2, 3+tier*2)
			factor = a * a * r
			ansStr := strconv.Itoa(factor)
			qText = fmt.Sprintf("For scaled cylindrical coordinates \\(x = %dr\\cos\\theta,\\ y = %dr\\sin\\theta,\\ z = z\\), evaluate the Jacobian determinant \\(\\frac{\\partial(x,y,z)}{\\partial(r,\\theta,z)}\\) at \\(r = %d\\).", a, a, r)
			if factor == a || factor == r || containsWord(qText, ansStr) {
				continue
			}
			ex = fmt.Sprintf("The Jacobian matrix for cylindrical transformation has determinant \\(\\frac{\\partial(x,y,z)}{\\partial(r,\\theta,z)} = %d^2 r = %d r\\). Substitute the evaluation radius \\(r = %d\\) into the Jacobian expression: \\(%d \\times %d = %d\\). Evaluate the product to find the volume scaling factor %d.",
				a, a*a, r, a*a, r, factor, factor)
			return generator.Problem{
				Question:    qText,
				Answer:      strconv.Itoa(factor),
				Explanation: ex,
				Facts: map[string]string{
					"a":      strconv.Itoa(a),
					"r":      strconv.Itoa(r),
					"factor": strconv.Itoa(factor),
					"answer": strconv.Itoa(factor),
				},
			}
		}

		// Spherical coordinates: Jacobian determinant is rho^2 sin(phi)
		rho = small(3, 3+tier*3)
		phi = "\\pi/2"
		factor = rho * rho
		if rand.Intn(2) == 0 && rho%2 == 0 {
			phi = "\\pi/6"
			factor = rho * rho / 2
		}
		ansStr := strconv.Itoa(factor)
		qText = fmt.Sprintf("For spherical coordinates \\((\\rho, \\theta, \\varphi)\\), evaluate the volume element factor \\(\\rho^2 \\sin\\varphi\\) at \\(\\rho = %d\\) and \\(\\varphi = %s\\).", rho, phi)
		if factor == rho || containsWord(qText, ansStr) {
			continue
		}
		sinValStr := "1"
		if phi == "\\pi/6" {
			sinValStr = "\\frac{1}{2}"
		}
		ex = fmt.Sprintf("The spherical coordinate volume element carries the Jacobian factor \\(dV = \\rho^2 \\sin\\varphi\\,d\\rho\\,d\\theta\\,d\\varphi\\). Substitute radial coordinate \\(\\rho = %d\\) and polar angle \\(\\varphi = %s\\), where \\(\\sin(%s) = %s\\). Evaluate the product \\((%d)^2 \\times %s = %d \\times %s = %d\\).",
			rho, phi, phi, sinValStr, rho, sinValStr, rho*rho, sinValStr, factor)
		return generator.Problem{
			Question:    qText,
			Answer:      strconv.Itoa(factor),
			Explanation: ex,
			Facts: map[string]string{
				"rho":    strconv.Itoa(rho),
				"factor": strconv.Itoa(factor),
				"answer": strconv.Itoa(factor),
			},
		}
	}
}
