package trigonometry

// Tests for the six right-triangle ratio concepts.
//
// Each test recomputes the side from the published ratio and the side the
// question starts from, so a generator that quietly changed the ratio, the
// given side or the arithmetic fails here. The corpus is checked too, because
// these concepts used to ship as multiple choice and a row that drifts back
// would otherwise pass silently.

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

var ratioIDs = []string{
	"trig.func.sine",
	"trig.func.cosine",
	"trig.func.tangent",
	"trig.func.secant",
	"trig.func.cosecant",
	"trig.func.cotangent",
}

type ratioCorpusConcept struct {
	ID          string `json:"id"`
	GradingType string `json:"grading_type"`
}

func ratioShippedTypes(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("../../../data/concepts/trigonometry.json")
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var rows []ratioCorpusConcept
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("parse corpus: %v", err)
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.GradingType
	}
	return out
}

func ratioRegistry(t *testing.T) *generator.Registry {
	t.Helper()
	reg := generator.NewRegistry()
	Register(reg)
	return reg
}

func ratioSample(t *testing.T, reg *generator.Registry, id string, seed int) generator.Problem {
	t.Helper()
	p, err := reg.GenerateContext(id, generator.GeneratorContext{Difficulty: 0.5})
	if err != nil {
		t.Fatalf("%s: generate: %v", id, err)
	}
	if _, ok := p.Facts["ratio"]; !ok {
		t.Fatalf("%s: no ratio fact at sample %d", id, seed)
	}
	return p
}

func ratioParts(t *testing.T, p generator.Problem) (num, den, given, answer int) {
	t.Helper()
	parts := strings.Split(p.Facts["ratio"], "/")
	if len(parts) != 2 {
		t.Fatalf("ratio fact is not a fraction: %q", p.Facts["ratio"])
	}
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatalf("numerator: %v", err)
	}
	den, err = strconv.Atoi(parts[1])
	if err != nil {
		t.Fatalf("denominator: %v", err)
	}
	given, err = strconv.Atoi(p.Facts["given"])
	if err != nil {
		t.Fatalf("given side: %v", err)
	}
	answer, err = strconv.Atoi(p.Answer)
	if err != nil {
		t.Fatalf("answer is not an integer: %q", p.Answer)
	}
	return num, den, given, answer
}

func TestRatioConceptsAreNumericInTheCorpus(t *testing.T) {
	types := ratioShippedTypes(t)
	for _, id := range ratioIDs {
		if types[id] != "numeric" {
			t.Errorf("%s: corpus says %q, want numeric", id, types[id])
		}
	}
}

func TestRatioAnswersRecomputeFromFacts(t *testing.T) {
	reg := ratioRegistry(t)

	// sine: opposite / hypotenuse, the hypotenuse is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.sine", i)
		opp, hyp, given, answer := ratioParts(t, p)
		if want := given * opp / hyp; answer != want {
			t.Fatalf("sine: ratio=%d/%d given=%d answer=%d want %d", opp, hyp, given, answer, want)
		}
	}

	// cosine: adjacent / hypotenuse, the hypotenuse is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.cosine", i)
		adj, hyp, given, answer := ratioParts(t, p)
		if want := given * adj / hyp; answer != want {
			t.Fatalf("cosine: ratio=%d/%d given=%d answer=%d want %d", adj, hyp, given, answer, want)
		}
	}

	// tangent: opposite / adjacent, the adjacent side is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.tangent", i)
		opp, adj, given, answer := ratioParts(t, p)
		if want := given * opp / adj; answer != want {
			t.Fatalf("tangent: ratio=%d/%d given=%d answer=%d want %d", opp, adj, given, answer, want)
		}
	}

	// secant: hypotenuse / adjacent, the adjacent side is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.secant", i)
		hyp, adj, given, answer := ratioParts(t, p)
		if want := given * hyp / adj; answer != want {
			t.Fatalf("secant: ratio=%d/%d given=%d answer=%d want %d", hyp, adj, given, answer, want)
		}
	}

	// cosecant: hypotenuse / opposite, the opposite side is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.cosecant", i)
		hyp, opp, given, answer := ratioParts(t, p)
		if want := given * hyp / opp; answer != want {
			t.Fatalf("cosecant: ratio=%d/%d given=%d answer=%d want %d", hyp, opp, given, answer, want)
		}
	}

	// cotangent: adjacent / opposite, the opposite side is given.
	for i := 0; i < 100; i++ {
		p := ratioSample(t, reg, "trig.func.cotangent", i)
		adj, opp, given, answer := ratioParts(t, p)
		if want := given * adj / opp; answer != want {
			t.Fatalf("cotangent: ratio=%d/%d given=%d answer=%d want %d", adj, opp, given, answer, want)
		}
	}
}

func TestPeriodGraphRecomputesFromFacts(t *testing.T) {
	types := ratioShippedTypes(t)
	if types["trig.graph.period"] != "numeric" {
		t.Fatalf("trig.graph.period corpus grading_type = %q, want numeric", types["trig.graph.period"])
	}

	reg := ratioRegistry(t)
	for i := 0; i < 100; i++ {
		p, err := reg.GenerateContext("trig.graph.period", generator.GeneratorContext{Difficulty: 0.5})
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		b, err := strconv.Atoi(p.Facts["B"])
		if err != nil || b <= 0 {
			t.Fatalf("invalid fact B: %q", p.Facts["B"])
		}
		want := 360 / b
		ans, err := strconv.Atoi(p.Answer)
		if err != nil {
			t.Fatalf("invalid answer: %q", p.Answer)
		}
		if ans != want {
			t.Fatalf("sample %d: B=%d answer=%d want %d", i, b, ans, want)
		}
		if p.Facts["answer"] != p.Answer {
			t.Fatalf("fact answer %q != p.Answer %q", p.Facts["answer"], p.Answer)
		}
	}
}
