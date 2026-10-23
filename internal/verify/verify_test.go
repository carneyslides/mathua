package verify

import (
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"2+2", 4},
		{"3*4", 12},
		{"10-3", 7},
		{"20/4", 5},
		{"2**3", 8},
		{"3+4*2", 11},
		{"(3+4)*2", 14},
		{"25/100*80", 20},
		{"100-35", 65},
	}
	for _, tt := range tests {
		got, err := Eval(tt.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tt.expr, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("Eval(%q) = %g, want %g", tt.expr, got, tt.expected)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"syntax parse error", "2 + * 3"},
		{"unbalanced parenthesis", "(2 + 3"},
		{"invalid operator type evaluation error", "'hello' - 'world'"},
		{"boolean expression type", "2 == 2"},
		{"string comparison type", "'a' == 'a'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(tt.expr)
			if err == nil {
				t.Fatalf("Eval(%q) expected error, got %v", tt.expr, got)
			}
		})
	}
}

func TestCheckExpr(t *testing.T) {
	if err := CheckExpr("2+2", "4"); err != nil {
		t.Errorf("CheckExpr(2+2, 4) failed: %v", err)
	}
	if err := CheckExpr("3*4", "12"); err != nil {
		t.Errorf("CheckExpr(3*4, 12) failed: %v", err)
	}
	if err := CheckExpr("2+2", "5"); err == nil {
		t.Error("CheckExpr(2+2, 5) should have failed")
	}
	if err := CheckExpr("2 + * 3", "4"); err == nil {
		t.Error("CheckExpr with invalid syntax should have failed")
	}
	if err := CheckExpr("2+2", "not-a-number"); err == nil {
		t.Error("CheckExpr with unparseable expected value should have failed")
	}

	// Precision tolerance check (within 1e-9 passes, outside 1e-9 fails)
	if err := CheckExpr("1.00000000005", "1.0"); err != nil {
		t.Errorf("CheckExpr within 1e-9 tolerance failed: %v", err)
	}
	if err := CheckExpr("1.00000001", "1.0"); err == nil {
		t.Error("CheckExpr exceeding 1e-9 tolerance should have failed")
	}
}

func TestExtractExpr(t *testing.T) {
	tests := []struct {
		question string
		expected string
		ok       bool
	}{
		{"5 + 3 = ?", "5+3", true},
		{"5+3=?", "5+3", true},
		{"5 + 3 = ____", "5+3", true},
		{"-4 + -8 = ?", "-4+(0-8)", true},
		{"12 - 7 = ?", "12-(7)", true},
		{"-6 - -2 = ?", "-6-((0-2))", true},
		{"4 x 3 = ?", "4*3", true},
		{"4 × 3 = ?", "4*3", true},
		{"8 * 2 = ?", "8*2", true},
		{"5 * -3 = ?", "5*(0-3)", true},
		{"What is 25% of 80?", "25/100*80", true},
		{"What is 10 + 25 = ?", "10+25", true},
		{"What is 14 + 19", "14+19", true},
		{"count the blue triangles", "", false},
		{"random text with no pattern", "", false},
	}
	for _, tt := range tests {
		got, ok := ExtractExpr(tt.question)
		if ok != tt.ok {
			t.Errorf("ExtractExpr(%q) ok=%v, want %v", tt.question, ok, tt.ok)
			continue
		}
		if ok && got != tt.expected {
			t.Errorf("ExtractExpr(%q) = %q, want %q", tt.question, got, tt.expected)
		}
	}
}
