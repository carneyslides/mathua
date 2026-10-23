package realanalysis

import (
	"math"
	"strconv"
	"testing"
)

func TestConvertedMetricSpacesRecomputesFromFacts(t *testing.T) {
	reg := raRegistry(t)

	for i := 0; i < 200; i++ {
		p := raSample(t, reg, "ra.metric.spaces", i)
		metric := p.Facts["metric"]
		ans := raFactInt(t, p, "answer")
		if p.Answer != strconv.Itoa(ans) {
			t.Fatalf("sample %d: p.Answer = %q, fact answer = %d", i, p.Answer, ans)
		}

		switch metric {
		case "d_2":
			px := raFactInt(t, p, "px")
			py := raFactInt(t, p, "py")
			qx := raFactInt(t, p, "qx")
			qy := raFactInt(t, p, "qy")
			dx := px - qx
			dy := py - qy
			expectedSq := dx*dx + dy*dy
			expected := int(math.Round(math.Sqrt(float64(expectedSq))))
			if expected*expected != expectedSq {
				t.Fatalf("sample %d: euclidean distance squared %d is not a perfect square", i, expectedSq)
			}
			if ans != expected {
				t.Fatalf("sample %d: euclidean distance recomputed %d != ans %d", i, expected, ans)
			}

		case "d_1":
			px := raFactInt(t, p, "px")
			py := raFactInt(t, p, "py")
			pz := raFactInt(t, p, "pz")
			qx := raFactInt(t, p, "qx")
			qy := raFactInt(t, p, "qy")
			qz := raFactInt(t, p, "qz")
			expected := intAbs(px-qx) + intAbs(py-qy) + intAbs(pz-qz)
			if ans != expected {
				t.Fatalf("sample %d: d_1 recomputed %d != ans %d", i, expected, ans)
			}

		case "d_infty":
			px := raFactInt(t, p, "px")
			py := raFactInt(t, p, "py")
			pz := raFactInt(t, p, "pz")
			qx := raFactInt(t, p, "qx")
			qy := raFactInt(t, p, "qy")
			qz := raFactInt(t, p, "qz")
			expected := intMax(intAbs(px-qx), intMax(intAbs(py-qy), intAbs(pz-qz)))
			if ans != expected {
				t.Fatalf("sample %d: d_infty recomputed %d != ans %d", i, expected, ans)
			}

		case "diameter_d1":
			p1x := raFactInt(t, p, "p1x")
			p1y := raFactInt(t, p, "p1y")
			p2x := raFactInt(t, p, "p2x")
			p2y := raFactInt(t, p, "p2y")
			p3x := raFactInt(t, p, "p3x")
			p3y := raFactInt(t, p, "p3y")
			d12 := intAbs(p1x-p2x) + intAbs(p1y-p2y)
			d13 := intAbs(p1x-p3x) + intAbs(p1y-p3y)
			d23 := intAbs(p2x-p3x) + intAbs(p2y-p3y)
			expected := intMax(d12, intMax(d13, d23))
			if ans != expected {
				t.Fatalf("sample %d: diam recomputed %d != ans %d", i, expected, ans)
			}

		case "discrete_ball":
			radius := raFactInt(t, p, "radius")
			setSize := raFactInt(t, p, "set_size")
			var expected int
			if radius == 1 {
				expected = 1
			} else if radius == 2 {
				expected = setSize
			} else {
				t.Fatalf("sample %d: unexpected radius %d", i, radius)
			}
			if ans != expected {
				t.Fatalf("sample %d: discrete ball recomputed %d != ans %d", i, expected, ans)
			}

		default:
			t.Fatalf("sample %d: unknown metric mode %q", i, metric)
		}
	}
}
