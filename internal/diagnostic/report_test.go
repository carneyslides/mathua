package diagnostic

import (
	"testing"

	"github.com/chuma-beep/mathua/internal/concepts"
)

func TestDomainToCourse_AllDomains(t *testing.T) {
	cases := []struct {
		domain, subdomain, expected string
	}{
		// Foundational domains
		{"arithmetic", "arithmetic.basics", "4"},
		{"fractions", "fractions.add", "5"},
		{"prealgebra", "prealgebra.expressions", "pa"},
		{"algebra", "algebra.linear", "a1"},
		{"trigonometry", "trigonometry.unit_circle", "a2"},
		{"precalculus", "precalculus.functions", "pc"},

		// Calculus and advanced calculus
		{"calculus", "calculus.limits", "calc1"},
		{"calculus", "calculus.derivatives", "calc1"},
		{"calculus", "calculus.integrals", "calc1"},
		{"calculus", "calculus.series", "calc2"},
		{"calculus", "calculus.sequences", "calc2"},

		// Statistics and probability
		{"statistics", "statistics.descriptive", "probstat"},
		{"probability_theory", "probability_theory.basics", "prob"},

		// Linear algebra and geometry
		{"linear_algebra", "linear_algebra.matrices", "linalg"},
		{"geometry", "geometry.triangles", "geo"},

		// Discrete math
		{"discrete_math", "discrete_math.logic", "discrete"},
		{"discrete_math", "discrete_math.sets", "discrete"},
		{"discrete_math", "discrete_math.combinatorics", "discrete"},
		{"discrete_math", "discrete_math.graph_theory", "discrete2"},
		{"discrete_math", "discrete_math.proof", "discrete2"},
		{"discrete_math", "discrete_math.recurrence", "discrete2"},

		// Number theory, complex and ODEs
		{"number_theory", "number_theory.basics", "nt"},
		{"complex", "complex.arithmetic", "a2"},
		{"differential_equations", "differential_equations.first_order", "diffeq"},

		// Abstract algebra
		{"abstract_algebra", "abstract_algebra.groups", "abstract"},
		{"abstract_algebra", "abstract_algebra.rings", "abstract2"},
		{"abstract_algebra", "abstract_algebra.fields", "abstract2"},
		{"abstract_algebra", "abstract_algebra.structures", "abstract2"},

		// Advanced collegiate mathematics
		{"topology", "topology.open_sets", "topo"},
		{"machine_learning", "machine_learning.gradient", "ml"},
		{"real_analysis", "real_analysis.sequences", "ra"},
		{"graph_theory", "graph_theory.basics", "graph"},
		{"numerical_analysis", "numerical_analysis.methods", "numa"},

		// Unknown fallback
		{"unknown_domain", "unknown_subdomain", ""},
	}
	for _, tc := range cases {
		c := &concepts.Concept{Domain: tc.domain, Subdomain: tc.subdomain}
		if got := domainToCourse(c); got != tc.expected {
			t.Errorf("%s/%s: expected %q, got %q", tc.domain, tc.subdomain, tc.expected, got)
		}
	}
}
