package lessons

import (
	"strings"
	"testing"
)

func TestRenderSlice(t *testing.T) {
	s := Slice{
		Rule: "The derivative of a constant is zero.",
		Examples: []Example{
			{Item: "$d/dx(5) = 0$", Says: "constant"},
			{Item: "   ", Says: "ignored empty item"},
			{Item: "$d/dx(\\pi) = 0$", Says: ""},
		},
		Note: "Do not confuse pi with a variable.",
	}

	rendered := renderSlice("Constant Rule", s)

	if !strings.Contains(rendered, "**Constant Rule**") {
		t.Errorf("expected rendered slice to include label header, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "The derivative of a constant is zero.") {
		t.Errorf("expected rule in output, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "- $d/dx(5) = 0$ — constant") {
		t.Errorf("expected example with explanation, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "- $d/dx(\\pi) = 0$\n") {
		t.Errorf("expected example without explanation, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "ignored empty item") {
		t.Errorf("expected empty item to be skipped, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Watch out: Do not confuse pi with a variable.") {
		t.Errorf("expected note in output, got:\n%s", rendered)
	}

	// Minimal slice with no label, no examples and no note
	emptySlice := Slice{Rule: "Just a rule."}
	renderedEmpty := renderSlice("", emptySlice)
	if strings.Contains(renderedEmpty, "**") {
		t.Errorf("unexpected header for empty label, got:\n%s", renderedEmpty)
	}
	if strings.Contains(renderedEmpty, "Examples") {
		t.Errorf("unexpected Examples section for slice without examples, got:\n%s", renderedEmpty)
	}
}

func TestRenderMinimalSlice(t *testing.T) {
	withLabel := renderMinimalSlice("Core Concept", nil)
	if !strings.Contains(withLabel, "**Core Concept**\n\n") {
		t.Errorf("expected label header, got:\n%s", withLabel)
	}
	if !strings.Contains(withLabel, "There isn't a written explanation for this step yet.") {
		t.Errorf("expected placeholder text, got:\n%s", withLabel)
	}

	withoutLabel := renderMinimalSlice("", nil)
	if strings.HasPrefix(withoutLabel, "**") {
		t.Errorf("expected no label header when empty, got:\n%s", withoutLabel)
	}
	if !strings.Contains(withoutLabel, "There isn't a written explanation for this step yet.") {
		t.Errorf("expected placeholder text, got:\n%s", withoutLabel)
	}
}

func TestTeachingSlice(t *testing.T) {
	l := &Loader{
		concepts: map[string]*Lesson{
			"calc.derivatives.power": {
				Title: "Power Rule",
				Body:  "## Power Rule Section\n\nMultiply by the exponent and subtract one.\n\n## Next\n\nOther content.",
			},
			"calc.derivatives.empty": {
				Title: "Empty Lesson",
				Body:  "Unstructured body without headings.",
			},
		},
		kps: map[string][]KP{},
	}

	// 1. Resolved section
	kpSection := KP{
		Label:   "Apply Exponent",
		Section: "Power Rule Section",
	}
	body, ok := l.TeachingSlice("calc.derivatives.power", kpSection)
	if !ok || !strings.Contains(body, "Multiply by the exponent") {
		t.Fatalf("expected resolved section body, got ok=%v body=%q", ok, body)
	}

	// 2. Authored slice when section does not resolve
	kpAuthored := KP{
		Label:   "Special Product",
		Section: "Nonexistent Section",
		Slice: &Slice{
			Rule: "Square the first term, add twice the product and square the last term.",
		},
	}
	body, ok = l.TeachingSlice("calc.derivatives.power", kpAuthored)
	if !ok || !strings.Contains(body, "**Special Product**") || !strings.Contains(body, "Square the first term") {
		t.Fatalf("expected authored slice rendered, got ok=%v body=%q", ok, body)
	}

	// 3. Fallback to minimal slice when no section matches and no authored slice exists
	kpMinimal := KP{
		Label: "Unauthored Step",
	}
	body, ok = l.TeachingSlice("calc.derivatives.empty", kpMinimal)
	if !ok || !strings.Contains(body, "There isn't a written explanation for this step yet.") {
		t.Fatalf("expected minimal slice placeholder, got ok=%v body=%q", ok, body)
	}
}

func TestSliceAudit(t *testing.T) {
	loader := &Loader{
		concepts: map[string]*Lesson{
			"c.first": {
				Title: "First Lesson",
				Body:  "## First Section\n\nShort section.\n\n## Next\n\nEnd.",
			},
			"c.second": {
				Title: "Second Lesson",
				Body:  "Word one word two word three.",
			},
		},
		kps: map[string][]KP{
			"c.second": {
				{
					Label: "KP B Minimal",
				},
				{
					Label: "KP A Authored",
					Slice: &Slice{Rule: "Authored rule explanation."},
				},
			},
			"c.first": {
				{
					Label:   "KP Section",
					Section: "First Section",
				},
			},
		},
	}

	findings := loader.SliceAudit()
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	// Findings must be sorted by ConceptID and then KPLabel
	if findings[0].ConceptID != "c.first" || findings[0].KPLabel != "KP Section" {
		t.Errorf("unexpected item 0: %+v", findings[0])
	}
	if findings[0].Level != SliceLevelSection || !findings[0].FirstCard || findings[0].FullBodyWords != 0 {
		t.Errorf("unexpected properties for section finding: %+v", findings[0])
	}

	if findings[1].ConceptID != "c.second" || findings[1].KPLabel != "KP A Authored" {
		t.Errorf("unexpected item 1: %+v", findings[1])
	}
	if findings[1].Level != SliceLevelAuthored || findings[1].FirstCard || findings[1].FullBodyWords != 0 {
		t.Errorf("unexpected properties for authored finding: %+v", findings[1])
	}

	if findings[2].ConceptID != "c.second" || findings[2].KPLabel != "KP B Minimal" {
		t.Errorf("unexpected item 2: %+v", findings[2])
	}
	if findings[2].Level != SliceLevelMinimal || !findings[2].FirstCard || findings[2].FullBodyWords != 6 {
		t.Errorf("unexpected properties for minimal finding: %+v", findings[2])
	}
}
