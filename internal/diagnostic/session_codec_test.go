package diagnostic

import (
	"encoding/json"
	"testing"
)

func TestSession_EncodeDecode_RoundTrip(t *testing.T) {
	d := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(d, reg)
	session := eng.Start()
	session.ID = "session-roundtrip-1"
	session.StudentID = "student-roundtrip-1"

	// Ask and answer two questions.
	p1, cid1, err := eng.NextQuestion(session)
	if err != nil || cid1 == "" || p1 == nil {
		t.Fatalf("expected question, got cid=%q, err=%v", cid1, err)
	}
	eng.RecordAnswer(session, cid1, true, true)

	p2, cid2, err := eng.NextQuestion(session)
	if err != nil || cid2 == "" || p2 == nil {
		t.Fatalf("expected second question, got cid=%q, err=%v", cid2, err)
	}
	eng.RecordAnswerTimed(session, cid2, false, 8.5, 10.0)

	// Encode session.
	data, err := session.Encode()
	if err != nil {
		t.Fatalf("encode session: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("encoded session is empty")
	}

	// Decode session.
	restored, err := eng.Decode(data)
	if err != nil {
		t.Fatalf("decode session: %v", err)
	}

	// Verify restored invariants.
	if restored.ID != session.ID {
		t.Errorf("restored ID = %q, want %q", restored.ID, session.ID)
	}
	if restored.StudentID != session.StudentID {
		t.Errorf("restored StudentID = %q, want %q", restored.StudentID, session.StudentID)
	}
	if restored.State != session.State {
		t.Errorf("restored State = %q, want %q", restored.State, session.State)
	}
	if len(restored.Attempts) != len(session.Attempts) {
		t.Fatalf("restored attempts count = %d, want %d", len(restored.Attempts), len(session.Attempts))
	}
	for i, att := range restored.Attempts {
		if att.ConceptID != session.Attempts[i].ConceptID {
			t.Errorf("attempt %d ConceptID = %q, want %q", i, att.ConceptID, session.Attempts[i].ConceptID)
		}
		if att.Correct != session.Attempts[i].Correct {
			t.Errorf("attempt %d Correct = %v, want %v", i, att.Correct, session.Attempts[i].Correct)
		}
	}
	if len(restored.order) != len(session.order) {
		t.Errorf("restored order len = %d, want %d", len(restored.order), len(session.order))
	}
	if restored.beliefs[cid1] != session.beliefs[cid1] {
		t.Errorf("restored belief[%s] = %v, want %v", cid1, restored.beliefs[cid1], session.beliefs[cid1])
	}
	if restored.confidence[cid1] != session.confidence[cid1] {
		t.Errorf("restored confidence[%s] = %v, want %v", cid1, restored.confidence[cid1], session.confidence[cid1])
	}
	if restored.doneSet[cid1] != session.doneSet[cid1] {
		t.Errorf("restored doneSet[%s] = %v, want %v", cid1, restored.doneSet[cid1], session.doneSet[cid1])
	}
}

func TestSession_Decode_EdgeCases(t *testing.T) {
	d := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(d, reg)

	// 1. Invalid JSON.
	if _, err := eng.Decode([]byte(`{not valid json}`)); err == nil {
		t.Error("expected error decoding malformed JSON")
	}

	// 2. Missing ID.
	if _, err := eng.Decode([]byte(`{"order":["a","b"]}`)); err == nil {
		t.Error("expected error decoding snapshot without ID")
	}

	// 3. Snapshot with unknown concepts only.
	unknownSnap := Snapshot{
		ID:    "session-unknown",
		Order: []string{"unknown_x", "unknown_y"},
	}
	rawUnknown, _ := json.Marshal(unknownSnap)
	if _, err := eng.Decode(rawUnknown); err == nil {
		t.Error("expected error when no concepts survive DAG lookup")
	}

	// 4. Missing/nil maps populate safe non-nil maps.
	nilMapsSnap := Snapshot{
		ID:    "session-nil-maps",
		Order: []string{"a", "b"},
	}
	rawNil, _ := json.Marshal(nilMapsSnap)
	restored, err := eng.Decode(rawNil)
	if err != nil {
		t.Fatalf("decode nil maps snapshot: %v", err)
	}
	if restored.beliefs == nil || restored.confidence == nil || restored.probeCounts == nil || restored.doneSet == nil {
		t.Error("expected nil maps to be initialized to safe empty maps")
	}
	if restored.State != StateProbing {
		t.Errorf("restored state with empty state = %q, want %q", restored.State, StateProbing)
	}
}

func TestEngine_RecordDontKnow(t *testing.T) {
	d := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(d, reg)
	session := eng.Start()

	initialBelief := session.beliefs["a"]
	eng.RecordDontKnow(session, "a", 2.0, 10.0)

	if len(session.Attempts) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(session.Attempts))
	}
	att := session.Attempts[0]
	if att.Correct {
		t.Errorf("admitted dont-know was marked correct")
	}
	if att.Fast {
		t.Errorf("admitted dont-know should never be fast")
	}
	if session.beliefs["a"] >= initialBelief {
		t.Errorf("expected belief to drop after dont-know, got initial=%v, current=%v",
			initialBelief, session.beliefs["a"])
	}
}

func TestEngine_SettleCurrent(t *testing.T) {
	d := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(d, reg)
	session := eng.Start()

	// Before any question is served, nothing to settle.
	settled := eng.SettleCurrent(session)
	if settled != "" {
		t.Errorf("expected empty string when no question pending, got %q", settled)
	}

	// Serve question.
	p, cid, err := eng.NextQuestion(session)
	if err != nil || cid == "" || p == nil {
		t.Fatal("expected question to be served")
	}

	// Settle current pending question.
	settled = eng.SettleCurrent(session)
	if settled != cid {
		t.Errorf("settled = %q, want %q", settled, cid)
	}
	if !session.doneSet[cid] {
		t.Errorf("concept %q was not marked in doneSet", cid)
	}
}

func TestEngine_FrontierAndKnowledgeConfidence(t *testing.T) {
	d := testDAG(t)
	reg := mustRegistry(t)
	eng := NewEngine(d, reg)
	session := eng.Start()

	// Initially all beliefs are unasserted/initial prior (0.1), so frontier is -1.
	if frontier := eng.FrontierEstimate(session); frontier != -1 {
		t.Errorf("initial frontier = %d, want -1", frontier)
	}

	// Manually boost belief of "a" (index 0) and "b" (index 1) past threshold (0.75).
	session.beliefs["a"] = 0.85
	session.beliefs["b"] = 0.90
	if frontier := eng.FrontierEstimate(session); frontier != 1 {
		t.Errorf("frontier after boosting 'b' = %d, want 1", frontier)
	}

	// Check KnowledgeConfidence returns safe copy.
	conf := eng.KnowledgeConfidence(session)
	if conf == nil {
		t.Fatal("expected non-nil confidence map")
	}
	// Modifying the returned map must not alter the session's internal state.
	conf["test_mutation"] = 1.0
	if session.confidence["test_mutation"] != 0.0 {
		t.Error("modifying returned confidence map mutated internal session state")
	}
}
