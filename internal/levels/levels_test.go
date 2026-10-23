package levels

import (
	"testing"
)

func TestComputeTiers(t *testing.T) {
	cases := []struct {
		mastered int
		want     string
	}{
		{-10, "Novice"},
		{-1, "Novice"},
		{0, "Novice"},
		{1, "Novice"},
		{31, "Novice"},
		{32, "Apprentice"},
		{33, "Apprentice"},
		{63, "Apprentice"},
		{64, "Student"},
		{95, "Student"},
		{96, "Scholar"},
		{127, "Scholar"},
		{128, "Adept"},
		{159, "Adept"},
		{160, "Expert"},
		{191, "Expert"},
		{192, "Master"},
		{223, "Master"},
		{224, "Grandmaster"},
		{255, "Grandmaster"},
		{256, "Math Architect"},
		{300, "Math Architect"},
		{1250, "Math Architect"},
	}
	for _, tc := range cases {
		if got := Compute(tc.mastered); got != tc.want {
			t.Errorf("Compute(%d) = %q, want %q", tc.mastered, got, tc.want)
		}
	}
}

func TestAllTiersSortedAndNonEmpty(t *testing.T) {
	if len(All) == 0 {
		t.Fatal("levels.All must not be empty")
	}
	prevMin := -1
	for i, l := range All {
		if l.Title == "" {
			t.Errorf("level %d has empty title", i)
		}
		if l.Min <= prevMin {
			t.Errorf("level %d Min (%d) must be strictly greater than previous (%d)", i, l.Min, prevMin)
		}
		prevMin = l.Min
	}
}
