package topology

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func fuzzGen(t *testing.T, gen generator.Generator) {
	t.Helper()
	for i := 0; i < 100; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})
		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Errorf("empty field at difficulty=%.2f: q=%q a=%q e=%q", d, p.Question, p.Answer, p.Explanation)
		}
	}
}

func TestMetricGen(t *testing.T)     { fuzzGen(t, &metricGen{}) }
func TestOpenClosedGen(t *testing.T) { fuzzGen(t, &openClosedGen{}) }
func TestContinuousGen(t *testing.T) { fuzzGen(t, &continuousGen{}) }

// dimensionGen is the one topology concept that produces a number rather than
// recognising a statement, so it is checked for the shape a produced answer
// needs: a numeric answer and the facts an authored schema rests on.
func TestDimensionGen(t *testing.T) {
	gen := &dimensionGen{}
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		p := gen.Generate(generator.GeneratorContext{Difficulty: rand.Float64()})
		if _, err := strconv.Atoi(p.Answer); err != nil {
			t.Fatalf("answer %q is not a number", p.Answer)
		}
		for _, k := range []string{"space", "dim", "rule"} {
			if p.Facts[k] == "" {
				t.Fatalf("fact %q not published on %+v", k, p.Facts)
			}
		}
		if p.Facts["dim"] != p.Answer {
			t.Fatalf("dim fact %q disagrees with answer %q", p.Facts["dim"], p.Answer)
		}
		seen[p.Answer] = true
	}
	if len(seen) < 2 {
		t.Fatalf("only one distinct answer %v; the pool does not discriminate", seen)
	}
}

func TestFundamentalGen(t *testing.T) {
	gen := &fundamentalGen{}
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		p := gen.Generate(generator.GeneratorContext{Difficulty: rand.Float64()})
		ans, err := strconv.Atoi(p.Answer)
		if err != nil {
			t.Fatalf("answer %q is not an integer", p.Answer)
		}
		if p.Facts["answer"] != p.Answer {
			t.Fatalf("fact answer %q != p.Answer %q", p.Facts["answer"], p.Answer)
		}

		kind := p.Facts["kind"]
		switch kind {
		case "torus":
			n, _ := strconv.Atoi(p.Facts["n"])
			if ans != n {
				t.Fatalf("torus: n=%d answer=%d", n, ans)
			}
		case "wedge_circles":
			k, _ := strconv.Atoi(p.Facts["k"])
			if ans != k {
				t.Fatalf("wedge: k=%d answer=%d", k, ans)
			}
		case "punctured_plane":
			k, _ := strconv.Atoi(p.Facts["points"])
			if ans != k {
				t.Fatalf("punctured: points=%d answer=%d", k, ans)
			}
		case "lens_space":
			prime, _ := strconv.Atoi(p.Facts["p"])
			if ans != prime {
				t.Fatalf("lens: p=%d answer=%d", prime, ans)
			}
		case "surface_genus":
			g, _ := strconv.Atoi(p.Facts["genus"])
			if ans != 2*g {
				t.Fatalf("genus: g=%d answer=%d want %d", g, ans, 2*g)
			}
		case "sphere":
			if ans != 1 {
				t.Fatalf("sphere: answer=%d want 1", ans)
			}
		default:
			t.Fatalf("unknown kind %q", kind)
		}
		seen[p.Answer] = true
	}
	if len(seen) < 4 {
		t.Fatalf("only %d distinct answers %v; fundamental pool lacks diversity", len(seen), seen)
	}
}

func TestFuzz(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	for _, id := range reg.Concepts() {
		gen, _ := reg.Get(id)
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				d := rand.Float64()
				p := gen.Generate(generator.GeneratorContext{Difficulty: d})
				if p.Question == "" || p.Answer == "" || p.Explanation == "" {
					t.Fatalf("empty field at difficulty=%.2f for %s: q=%q a=%q", d, id, p.Question, p.Answer)
				}
			}
		})
	}
}
