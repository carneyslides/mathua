package diagnostic

import (
	"math"
	"testing"

	"github.com/chuma-beep/mathua/internal/concepts"
)

func Test2PL_ItemCharacteristicCurve(t *testing.T) {
	a := 1.5
	b := 0.5

	// At theta == b, P(theta) must equal 0.5
	pMid := Prob2PL(b, a, b)
	if math.Abs(pMid-0.5) > 1e-9 {
		t.Fatalf("Prob2PL at b=%f got %f, want 0.5", b, pMid)
	}

	// P(theta) must strictly increase monotonically with theta
	thetas := []float64{-2.0, -1.0, 0.0, 0.5, 1.0, 2.0}
	prev := -1.0
	for _, th := range thetas {
		p := Prob2PL(th, a, b)
		if p <= prev {
			t.Fatalf("Prob2PL non-monotonic at theta=%f: prev=%f cur=%f", th, prev, p)
		}
		prev = p
	}
}

func Test2PL_FisherInformationPeak(t *testing.T) {
	a := 1.8
	b := 0.2

	// Fisher Information must peak at theta == b
	infoPeak := FisherInformation(b, a, b)
	expectedPeak := a * a * 0.25
	if math.Abs(infoPeak-expectedPeak) > 1e-9 {
		t.Fatalf("FisherInformation peak at theta=b got %f, want %f", infoPeak, expectedPeak)
	}

	// Any theta != b must yield strictly less information
	for _, offset := range []float64{-1.5, -0.5, 0.5, 1.5} {
		info := FisherInformation(b+offset, a, b)
		if info >= infoPeak {
			t.Fatalf("FisherInformation at offset %f got %f >= peak %f", offset, info, infoPeak)
		}
	}
}

func Test2PL_AbilityEstimationMonotonicityAndPrecision(t *testing.T) {
	dag := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(dag, reg)

	// Session 1: answers all correct
	sHigh := eng.Start()
	for i := 0; i < 5; i++ {
		_, cid, err := eng.NextQuestion(sHigh)
		if err != nil || cid == "" {
			break
		}
		eng.RecordAnswer(sHigh, cid, true, true)
	}
	thHigh, seHigh := eng.AbilityEstimate(sHigh)

	// Session 2: answers all incorrect
	sLow := eng.Start()
	for i := 0; i < 5; i++ {
		_, cid, err := eng.NextQuestion(sLow)
		if err != nil || cid == "" {
			break
		}
		eng.RecordAnswer(sLow, cid, false, false)
	}
	thLow, seLow := eng.AbilityEstimate(sLow)

	if thHigh <= thLow {
		t.Fatalf("expected strong ability > weak ability: strong=%f weak=%f", thHigh, thLow)
	}
	if thHigh <= 0.0 {
		t.Fatalf("expected positive ability for strong responses, got %f", thHigh)
	}
	if thLow >= 0.0 {
		t.Fatalf("expected negative ability for weak responses, got %f", thLow)
	}
	if seHigh >= 1.0 || seLow >= 1.0 {
		t.Fatalf("expected measurement standard error to shrink below prior variance (1.0), got high=%f low=%f", seHigh, seLow)
	}
}

func Test2PL_ItemParametersCalibration(t *testing.T) {
	dag := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(dag, reg)
	s := eng.Start()

	c := &concepts.Concept{
		ID:               "calc.test.derivative",
		Prerequisites:    []string{"a", "b", "c"},
		KeyPrerequisites: []string{"a"},
		Encompasses:      []string{"b"},
		MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 45},
	}

	a, b := eng.ItemParameters(s, c)
	// Base (1.0) + connectivity + key prereq bonus + encompasses bonus
	if a <= 1.2 {
		t.Fatalf("expected composite node discrimination > 1.2, got %f", a)
	}
	if b < -3.0 || b > 3.0 {
		t.Fatalf("expected difficulty calibrated in [-3.0, 3.0], got %f", b)
	}

	// Verify monotonic depth calibration on DAG
	root := dag.Concept("a")
	capstone := dag.Concept("j")
	if root == nil || capstone == nil {
		t.Fatalf("expected concepts a and j in testDAG")
	}
	_, bRoot := eng.ItemParameters(s, root)
	_, bCap := eng.ItemParameters(s, capstone)
	if bRoot >= bCap {
		t.Fatalf("expected root difficulty (%f) < capstone difficulty (%f)", bRoot, bCap)
	}
	if bCap < 1.5 {
		t.Fatalf("expected capstone difficulty >= 1.5, got %f", bCap)
	}
}
