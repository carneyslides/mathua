package topology

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func assertNoReadOffAndDiverse(t *testing.T, name string, gen generator.Generator) {
	t.Helper()
	seenQuestions := make(map[string]bool)
	seenAnswers := make(map[string]bool)

	for i := 0; i < 500; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d, Seed: int64(i + 1)})
		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("%s draw %d produced empty field", name, i)
		}

		stem := p.Question
		if idx := strings.Index(stem, "\n\nA)"); idx != -1 {
			stem = stem[:idx]
		}

		if strings.Contains(stem, p.Answer) {
			t.Fatalf("%s draw %d leaked answer %q in question stem: %q", name, i, p.Answer, stem)
		}

		seenQuestions[p.Question] = true
		seenAnswers[p.Answer] = true
	}

	if len(seenQuestions) < 5 {
		t.Fatalf("%s produced only %d distinct questions across 500 draws, want >= 5", name, len(seenQuestions))
	}
	if len(seenAnswers) < 2 {
		t.Fatalf("%s produced only %d distinct answers across 500 draws, want >= 2", name, len(seenAnswers))
	}
}

func TestMutationAntiReadOffTopology(t *testing.T) {
	t.Run("coveringGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "coveringGen", &coveringGen{})
	})
	t.Run("vanKampenGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "vanKampenGen", &vanKampenGen{})
	})
	t.Run("fiberBundleGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "fiberBundleGen", &fiberBundleGen{})
	})
	t.Run("metricGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "metricGen", &metricGen{})
	})
	t.Run("cellularHomologyGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "cellularHomologyGen", &cellularHomologyGen{})
	})
	t.Run("deRhamGen", func(t *testing.T) {
		assertNoReadOffAndDiverse(t, "deRhamGen", &deRhamGen{})
	})
}
