package mathutil

import (
	"testing"
)

func TestIntMatrixBasic(t *testing.T) {
	m1 := NewIntMatrix(2, 2)
	m1.Data[0][0] = 1
	m1.Data[0][1] = 2
	m1.Data[1][0] = 3
	m1.Data[1][1] = 4

	m2 := NewIntMatrix(2, 2)
	m2.Data[0][0] = 2
	m2.Data[0][1] = 0
	m2.Data[1][0] = 1
	m2.Data[1][1] = 2

	prod, err := m1.Multiply(m2)
	if err != nil {
		t.Fatalf("Multiply failed: %v", err)
	}

	// [1 2] * [2 0] = [1*2 + 2*1, 1*0 + 2*2] = [4, 4]
	// [3 4]   [1 2]   [3*2 + 4*1, 3*0 + 4*2] = [10, 8]
	if prod.Data[0][0] != 4 || prod.Data[0][1] != 4 || prod.Data[1][0] != 10 || prod.Data[1][1] != 8 {
		t.Fatalf("Unexpected product matrix: %v", prod.Data)
	}
}

func TestSmithNormalFormRank(t *testing.T) {
	// Matrix of rank 1: [1 2; 2 4]
	m := NewIntMatrix(2, 2)
	m.Data[0][0] = 1
	m.Data[0][1] = 2
	m.Data[1][0] = 2
	m.Data[1][1] = 4

	inv, rank := m.SmithNormalForm()
	if rank != 1 {
		t.Fatalf("Expected rank 1, got %d", rank)
	}
	if len(inv) != 1 || inv[0] != 1 {
		t.Fatalf("Expected invariants [1], got %v", inv)
	}
}

func TestCellularBettiNumbersTorusAndKlein(t *testing.T) {
	// Torus CW complex: c0=1, c1=2, c2=1. d1=0, d2=0.
	c0, c1, c2 := 1, 2, 1
	d1 := NewIntMatrix(1, 2)
	d2 := NewIntMatrix(2, 1)

	b0, b1, b2 := CellularBettiNumbers(c0, c1, c2, d1, d2)
	if b0 != 1 || b1 != 2 || b2 != 1 {
		t.Fatalf("Torus Betti numbers expected (1, 2, 1), got (%d, %d, %d)", b0, b1, b2)
	}

	// Klein bottle CW complex: c0=1, c1=2, c2=1. d1=0, d2=[0; 2].
	d2Klein := NewIntMatrix(2, 1)
	d2Klein.Data[1][0] = 2

	kb0, kb1, kb2 := CellularBettiNumbers(c0, c1, c2, d1, d2Klein)
	if kb0 != 1 || kb1 != 1 || kb2 != 0 {
		t.Fatalf("Klein bottle Betti numbers expected (1, 1, 0), got (%d, %d, %d)", kb0, kb1, kb2)
	}
}
