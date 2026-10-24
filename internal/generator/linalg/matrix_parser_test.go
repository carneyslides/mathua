package linalg

import (
	"reflect"
	"testing"
)

func TestParseMatrixFormats(t *testing.T) {
	tests := []struct {
		input string
		want  [][]int
	}{
		{
			input: "[[1, 2], [3, 4]]",
			want:  [][]int{{1, 2}, {3, 4}},
		},
		{
			input: "[[1,2],[3,4]]",
			want:  [][]int{{1, 2}, {3, 4}},
		},
		{
			input: "[[1, 2, 3], [4, 5, 6], [7, 8, 9]]",
			want:  [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
		},
		{
			input: `\begin{pmatrix} 1 & 2 \\ 3 & 4 \end{pmatrix}`,
			want:  [][]int{{1, 2}, {3, 4}},
		},
		{
			input: `\begin{bmatrix} 1 & -2 \\ 0 & 5 \end{bmatrix}`,
			want:  [][]int{{1, -2}, {0, 5}},
		},
		{
			input: `\begin{matrix} 7 & 8 \\ 9 & 10 \end{matrix}`,
			want:  [][]int{{7, 8}, {9, 10}},
		},
		{
			input: `\left[ \begin{pmatrix} 1 & -2 \\ 3 & -4 \end{pmatrix} \right]`,
			want:  [][]int{{1, -2}, {3, -4}},
		},
		{
			input: `[[1, −2], [3, −4]]`,
			want:  [][]int{{1, -2}, {3, -4}},
		},
		{
			input: `1 & 2 \\ 3 & 4`,
			want:  [][]int{{1, 2}, {3, 4}},
		},
	}

	for _, tc := range tests {
		got, ok := parseMatrix(tc.input)
		if !ok {
			t.Fatalf("parseMatrix(%q) returned false, want %v", tc.input, tc.want)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("parseMatrix(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestGradeMatrixEquivalence(t *testing.T) {
	expected := "[[1,2],[3,4]]"
	latexInput := `\begin{pmatrix} 1 & 2 \\ 3 & 4 \end{pmatrix}`

	res := gradeMatrix(expected, latexInput)
	if !res.Correct {
		t.Fatalf("gradeMatrix(%q, %q) returned not correct: %v", expected, latexInput, res.Feedback)
	}

	resBracket := gradeMatrix(expected, "[[1, 2], [3, 4]]")
	if !resBracket.Correct {
		t.Fatalf("gradeMatrix bracket format failed: %v", resBracket.Feedback)
	}

	wrong := `\begin{pmatrix} 1 & 2 \\ 3 & 5 \end{pmatrix}`
	resWrong := gradeMatrix(expected, wrong)
	if resWrong.Correct {
		t.Fatalf("gradeMatrix expected incorrect for wrong matrix")
	}
}
