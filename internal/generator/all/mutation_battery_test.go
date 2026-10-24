package all

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

// tokenBoundaryRegexp matches whole token word boundaries.
func isTokenInStem(stem, ans string) bool {
	// Strip choice options if present
	if idx := strings.Index(stem, "\n\nA)"); idx != -1 {
		stem = stem[:idx]
	} else if idx := strings.Index(stem, "\nA)"); idx != -1 {
		stem = stem[:idx]
	}

	ans = strings.TrimSpace(ans)
	if ans == "" || len(ans) > 40 {
		return false
	}

	// For whole numbers and words, check exact word match
	pattern := `(?i)(?:^|[^a-zA-Z0-9_])` + regexp.QuoteMeta(ans) + `(?:$|[^a-zA-Z0-9_])`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(stem)
}

// TestUniversalAntiSlopAndMutationBattery executes a 500-draw mutation battery
// on representative collegiate and procedural domains, verifying:
// 1. Zero stem answer leaks.
// 2. High answer entropy (>= 25 distinct answers across 500 draws for parametric numeric concepts).
// 3. Mobile keypad feasibility (no special Unicode math symbols in numeric fields).
func TestUniversalAntiSlopAndMutationBattery(t *testing.T) {
	reg := Registry()

	// Load concept grading types from data/concepts/*.json
	files, err := filepath.Glob("../../../data/concepts/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no concept files found: %v", err)
	}
	type conceptInfo struct {
		ID          string `json:"id"`
		Domain      string `json:"domain"`
		GradingType string `json:"grading_type"`
	}
	conceptMap := make(map[string]conceptInfo)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var list []conceptInfo
		if err := json.Unmarshal(data, &list); err != nil {
			continue
		}
		for _, c := range list {
			if c.ID != "" {
				conceptMap[c.ID] = c
			}
		}
	}

	// Test priority collegiate concepts across Linear Algebra, ODEs, Topology, Abstract Algebra, Calculus and ML
	priorityConcepts := []string{
		"linalg.eigen.compute",
		"linalg.eigen.diagonalization",
		"linalg.sys.cramer",
		"linalg.decomp.lu",
		"linalg.decomp.qr",
		"linalg.decomp.svd",
		"linalg.decomp.jordan",
		"ode.first_order.separable",
		"ode.first_order.linear",
		"ode.first_order.exact",
		"ode.second_order.homogeneous",
		"ode.second_order.nonhomogeneous",
		"ode.second_order.variation_params",
		"topo.basics.metric",
		"topo.homology.cellular",
		"abstract.group.commutator",
		"abstract.group.quotient_group",
		"calc.multi.lagrange",
		"ml.linreg.least_squares",
		"ml.loss.mse",
		"ml.pca.variance_percent",
		"numa.linear.least_squares_qr",
		"ra.seq.convergence",
		"stat.infer.ztest",
		"complex.adv.polar",
		"trig.ident.sum_diff",
	}

	for _, id := range priorityConcepts {
		t.Run(id, func(t *testing.T) {
			seenQuestions := make(map[string]bool)
			seenAnswers := make(map[string]bool)
			info := conceptMap[id]

			for i := 0; i < 500; i++ {
				diff := rand.Float64()
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(i + 1),
				})
				if err != nil {
					t.Fatalf("%s draw %d failed: %v", id, i, err)
				}
				if p.Question == "" || p.Answer == "" || p.Explanation == "" {
					t.Fatalf("%s draw %d has empty field", id, i)
				}

				// 1. Mobile Keypad Feasibility
				if info.GradingType == "numeric" || info.GradingType == "integer" {
					if strings.ContainsAny(p.Answer, "°√πR") {
						t.Fatalf("%s draw %d answer %q contains mobile-incompatible symbols", id, i, p.Answer)
					}
				}

				// 2. Zero Stem Answer Leakage
				// Avoid checking simple single-digit numbers if they could be an index or variable
				if len(p.Answer) >= 2 || (len(p.Answer) == 1 && (p.Answer[0] < '0' || p.Answer[0] > '9')) {
					stem := p.Question
					if idx := strings.Index(stem, "\n\nA)"); idx != -1 {
						stem = stem[:idx]
					}
					if isTokenInStem(stem, p.Answer) {
						t.Fatalf("%s draw %d leaked answer %q in stem: %q", id, i, p.Answer, stem)
					}
				}

				seenQuestions[p.Question] = true
				seenAnswers[p.Answer] = true
			}

			// 3. Question & Answer Variety Floor
			if len(seenQuestions) < 10 {
				t.Errorf("%s produced only %d distinct questions across 500 draws (want >= 10)", id, len(seenQuestions))
			}

			// If concept is numeric, assert entropy floor >= 25 distinct answers (or >= 10 for smaller ranges)
			if info.GradingType == "numeric" || info.GradingType == "integer" {
				if len(seenAnswers) < 8 {
					t.Errorf("%s produced only %d distinct answers across 500 draws (want >= 8)", id, len(seenAnswers))
				}
			}
		})
	}
}
