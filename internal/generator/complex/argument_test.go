package complex

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/latex"
)

func TestArgumentGen_Properties(t *testing.T) {
	gen := &argumentGen{}
	distinctAnswers := make(map[string]bool)

	for i := 0; i < 200; i++ {
		diff := float64(i%10) / 10.0
		p := gen.Generate(generator.GeneratorContext{
			Seed:       int64(i * 17),
			Difficulty: diff,
		})

		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty field at iter %d: %+v", i, p)
		}

		deg, err := strconv.Atoi(p.Answer)
		if err != nil {
			t.Fatalf("answer %q not an integer at iter %d", p.Answer, i)
		}
		if deg < 0 || deg >= 360 {
			t.Errorf("degree %d out of bounds [0, 360) at iter %d", deg, i)
		}

		distinctAnswers[p.Answer] = true

		if p.Facts == nil {
			t.Fatalf("missing facts at iter %d", i)
		}
		factDegStr := p.Facts["arg_deg"]
		factDeg, err := strconv.Atoi(factDegStr)
		if err != nil || factDeg != deg {
			t.Errorf("fact arg_deg %v does not match answer %d at iter %d", factDegStr, deg, i)
		}
		zStr := p.Facts["z"]
		if zStr == "" {
			t.Errorf("fact z empty at iter %d", i)
		}
		quad := p.Facts["quadrant"]
		if quad == "" {
			t.Errorf("fact quadrant empty at iter %d", i)
		}

		if errs := latex.Validate(p.Question, latex.Generators); len(errs) > 0 {
			t.Errorf("question validation failed at iter %d: %v", i, errs)
		}
		if errs := latex.Validate(p.Explanation, latex.Generators); len(errs) > 0 {
			t.Errorf("explanation validation failed at iter %d: %v", i, errs)
		}

		switch deg {
		case 0:
			if !strings.Contains(quad, "positive real") {
				t.Errorf("deg 0 expected positive real axis, got %s", quad)
			}
		case 90:
			if !strings.Contains(quad, "positive imaginary") {
				t.Errorf("deg 90 expected positive imaginary axis, got %s", quad)
			}
		case 180:
			if !strings.Contains(quad, "negative real") {
				t.Errorf("deg 180 expected negative real axis, got %s", quad)
			}
		case 270:
			if !strings.Contains(quad, "negative imaginary") {
				t.Errorf("deg 270 expected negative imaginary axis, got %s", quad)
			}
		case 30, 45, 60:
			if !strings.Contains(quad, "quadrant 1") {
				t.Errorf("deg %d expected quadrant 1, got %s", deg, quad)
			}
		case 120, 135, 150:
			if !strings.Contains(quad, "quadrant 2") {
				t.Errorf("deg %d expected quadrant 2, got %s", deg, quad)
			}
		case 210, 225, 240:
			if !strings.Contains(quad, "quadrant 3") {
				t.Errorf("deg %d expected quadrant 3, got %s", deg, quad)
			}
		case 300, 315, 330:
			if !strings.Contains(quad, "quadrant 4") {
				t.Errorf("deg %d expected quadrant 4, got %s", deg, quad)
			}
		default:
			t.Errorf("unexpected angle degree %d", deg)
		}
	}

	if len(distinctAnswers) < 8 {
		t.Errorf("expected at least 8 distinct answers, got %d: %v", len(distinctAnswers), distinctAnswers)
	}
}

func TestArgumentGen_Registry(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	gen, err := reg.Get("complex.ops.argument")
	if err != nil || gen == nil {
		t.Fatalf("complex.ops.argument not registered: %v", err)
	}

	for i := 0; i < 50; i++ {
		p := gen.Generate(generator.GeneratorContext{
			Seed:       int64(i + 1),
			Difficulty: rand.Float64(),
		})
		if p.Answer == "" {
			t.Fatalf("empty answer for seed %d", i+1)
		}
		if _, err := strconv.Atoi(p.Answer); err != nil {
			t.Fatalf("expected integer answer, got %q", p.Answer)
		}
	}
}
