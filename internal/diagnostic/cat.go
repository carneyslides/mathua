package diagnostic

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
)

type State string

const (
	StateProbing State = "probing"
	StateDone    State = "done"
)

const (
	// minTotalQuestions is the point at which convergence is *considered*, not a
	// floor it must reach. shouldStop below requires the cover to be finished AND
	// this count, but NextQuestion ends the session outright when the covering set
	// runs out, and that path does not consult this constant. Every cover concept
	// needs two probes and no concept is asked twice, so a scope whose closure is
	// smaller than 13 concepts simply cannot reach 25 and nothing pads the gap,
	// because repeating settled concepts would invent evidence.
	//
	// It was described as "the floor" here and in docs/architecture.md until a
	// five-concept scope was found asking ten questions while reporting
	// min_total: 25. Both now say what the code does.
	minTotalQuestions = 25
	// maxTotalQuestions hard caps an adaptive session (supplemental included).
	maxTotalQuestions = 45
	// coverSize is the target size of the compressed covering set.
	coverSize = 30
	// beliefThreshold is the frontier cutoff: belief >= threshold is "known".
	beliefThreshold = 0.6
	// confidenceThreshold gates supplemental re-probes: cover concepts
	// settled below this confidence get one extra probe each.
	confidenceThreshold = 0.7
	// Bayes evidence weights (conflict resolution): a correct answer is
	// +evidenceCorrectWeight toward prerequisites, an incorrect answer is
	// −evidenceIncorrectWeight toward dependents; conflicting prior
	// evidence discounts both directions by conflictDiscount.
	evidenceCorrectWeight   = 0.3
	evidenceIncorrectWeight = 0.3
	conflictDiscount        = 0.5
	// conditionalStripFactor scales back "conditionally completed"
	// prerequisites (belief in [threshold, threshold+0.15)) on a miss,
	// the frontier falls backwards immediately on struggle.
	conditionalStripFactor = 0.8
	// maxSupplementalProbes caps extra re-probes per concept.
	maxSupplementalProbes = 1
)

type Session struct {
	sync.Mutex

	ID        string
	StudentID string
	State     State
	Attempts  []Attempt
	order     []*concepts.Concept

	// LastProblem stores the current problem the student saw.
	LastProblem     *generator.Problem
	LastConceptID   string
	LastConceptName string
	// prevProblem tracks the superseded pending question so a "silly
	// mistake" retry can take the student back to it (answer responses
	// always carry the next staged question; the voidable attempt belongs
	// to the previous one).
	prevProblem     *generator.Problem
	prevConceptID   string
	prevConceptName string

	// Per-concept student model (belief 0-1 + confidence 0-1).
	beliefs      map[string]float64
	confidence   map[string]float64
	probeCounts  map[string]int
	correctCount map[string]int
	totalCount   map[string]int
	totalAsked   int
	doneSet      map[string]bool
	// supplementalCount tracks extra re-probes per concept (cap enforced).
	supplementalCount map[string]int
	// focus marks the concepts the learner actually chose. The whole
	// prerequisite path is still assessed; the mark only decides which covered
	// concepts are probed first, so a run started from abstract algebra opens on
	// abstract algebra instead of on the arithmetic beneath it.
	focus map[string]bool
	// retryUsed marks that the pending question's "silly mistake" retry was
	// consumed; reset whenever NextQuestion serves a new problem.
	retryUsed bool

	// 2PL Item Response Theory latent ability and measurement precision.
	theta float64
	se    float64
}

type Attempt struct {
	ConceptID      string
	Correct        bool
	Fast           bool
	ElapsedSeconds float64
	Timestamp      time.Time
	// DontKnow marks an admitted unknown ("I don't know" button): clean
	// negative evidence, as opposed to a failed attempt which may be a slip.
	DontKnow bool
	// BeliefSnapshot captures concept beliefs before this attempt's update
	// so a voided "silly mistake" retry can restore exact prior state.
	BeliefSnapshot map[string]float64
}

// Progress is the backend truth for the progress bar. The bar is Answered over
// ProbeBudget. The budget is two probes for each cover concept, and
// maxTotalQuestions caps it. The bar used to count settled concepts over the
// whole cover, so it moved only when a concept reached its second probe.
// EstimatedTotal is a legacy hint, not a promise.
type Progress struct {
	Answered       int  `json:"answered"`
	EstimatedTotal int  `json:"estimated_total"`
	MinTotal       int  `json:"min_total"`
	MaxTotal       int  `json:"max_total"`
	ProbeBudget    int  `json:"probe_budget"`
	Done           bool `json:"done"`
}

type Engine struct {
	dag      *concepts.DAG
	registry *generator.Registry
}

func NewEngine(dag *concepts.DAG, reg *generator.Registry) *Engine {
	return &Engine{dag: dag, registry: reg}
}

// Snapshot is the durable form of a Session.
//
// It exists because the adaptive model is private: encoding a Session directly
// would persist its id and attempts and silently drop every unexported field,
// so a restored run would believe all concepts sit at the 0.5 prior and would
// re-ask questions the learner had already answered. The snapshot names each
// field, and Decode rebuilds the path through the DAG so the session keeps
// pointing at real concepts rather than at strings.
type Snapshot struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	State     State     `json:"state"`
	Attempts  []Attempt `json:"attempts"`

	// Order is the session's frozen path, by concept id.
	Order []string `json:"order"`

	LastProblem     *generator.Problem `json:"last_problem,omitempty"`
	LastConceptID   string             `json:"last_concept_id"`
	LastConceptName string             `json:"last_concept_name"`
	PrevProblem     *generator.Problem `json:"prev_problem,omitempty"`
	PrevConceptID   string             `json:"prev_concept_id"`
	PrevConceptName string             `json:"prev_concept_name"`

	Beliefs           map[string]float64 `json:"beliefs"`
	Confidence        map[string]float64 `json:"confidence"`
	ProbeCounts       map[string]int     `json:"probe_counts"`
	CorrectCount      map[string]int     `json:"correct_count"`
	TotalCount        map[string]int     `json:"total_count"`
	DoneSet           map[string]bool    `json:"done_set"`
	SupplementalCount map[string]int     `json:"supplemental_count"`
	// Focus survives a restart with the rest of the model. Without it a resumed
	// run would fall back to probing the foundations first, which is the
	// behaviour the focus exists to avoid.
	Focus      map[string]bool `json:"focus,omitempty"`
	TotalAsked int             `json:"total_asked"`
	RetryUsed  bool            `json:"retry_used"`
	Theta      float64         `json:"theta,omitempty"`
	SE         float64         `json:"se,omitempty"`
}

// Encode renders the session for storage. The lock is held across marshaling,
// not merely while copying: an answer arriving mid-encode must not produce a
// snapshot that pairs a new attempt count with old beliefs.
//
// Only the newest attempt keeps its belief snapshot. RetryServe restores the
// last miss, and nothing reads an older one, so carrying one per attempt would
// write a full belief map for every question answered, a megabyte or more by
// the end of a run, rewritten on every answer.
func (s *Session) Encode() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	order := make([]string, 0, len(s.order))
	for _, c := range s.order {
		order = append(order, c.ID)
	}
	attempts := make([]Attempt, len(s.Attempts))
	copy(attempts, s.Attempts)
	for i := 0; i < len(attempts)-1; i++ {
		attempts[i].BeliefSnapshot = nil
	}
	snap := Snapshot{
		ID:                s.ID,
		StudentID:         s.StudentID,
		State:             s.State,
		Attempts:          attempts,
		Order:             order,
		LastProblem:       s.LastProblem,
		LastConceptID:     s.LastConceptID,
		LastConceptName:   s.LastConceptName,
		PrevProblem:       s.prevProblem,
		PrevConceptID:     s.prevConceptID,
		PrevConceptName:   s.prevConceptName,
		Beliefs:           s.beliefs,
		Confidence:        s.confidence,
		ProbeCounts:       s.probeCounts,
		CorrectCount:      s.correctCount,
		TotalCount:        s.totalCount,
		DoneSet:           s.doneSet,
		SupplementalCount: s.supplementalCount,
		Focus:             s.focus,
		TotalAsked:        s.totalAsked,
		RetryUsed:         s.retryUsed,
		Theta:             s.theta,
		SE:                s.se,
	}
	return json.Marshal(snap)
}

// Decode rebuilds a session from a snapshot written by Encode.
//
// A concept the current corpus no longer knows is dropped from the path rather
// than restored as a nil pointer; if none survive, the caller must treat the
// session as gone instead of resuming a run that can never serve a question.
func (e *Engine) Decode(data []byte) (*Session, error) {
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("decode diagnostic snapshot: %w", err)
	}
	if snap.ID == "" {
		return nil, fmt.Errorf("diagnostic snapshot has no id")
	}
	order := make([]*concepts.Concept, 0, len(snap.Order))
	for _, id := range snap.Order {
		c := e.dag.Concept(id)
		if c == nil {
			continue
		}
		order = append(order, c)
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("diagnostic snapshot %s has no concepts the corpus still knows", snap.ID)
	}
	state := snap.State
	if state == "" {
		state = StateProbing
	}
	s := &Session{
		ID:                snap.ID,
		StudentID:         snap.StudentID,
		State:             state,
		Attempts:          snap.Attempts,
		order:             order,
		LastProblem:       snap.LastProblem,
		LastConceptID:     snap.LastConceptID,
		LastConceptName:   snap.LastConceptName,
		prevProblem:       snap.PrevProblem,
		prevConceptID:     snap.PrevConceptID,
		focus:             snap.Focus,
		prevConceptName:   snap.PrevConceptName,
		beliefs:           orZero(snap.Beliefs),
		confidence:        orZero(snap.Confidence),
		probeCounts:       orZeroInt(snap.ProbeCounts),
		correctCount:      orZeroInt(snap.CorrectCount),
		totalCount:        orZeroInt(snap.TotalCount),
		doneSet:           orFalse(snap.DoneSet),
		supplementalCount: orZeroInt(snap.SupplementalCount),
		totalAsked:        snap.TotalAsked,
		retryUsed:         snap.RetryUsed,
		theta:             snap.Theta,
		se:                snap.SE,
	}
	if s.se == 0 {
		s.se = 1.0
	}
	return s, nil
}

func orZero(m map[string]float64) map[string]float64 {
	if m == nil {
		return map[string]float64{}
	}
	return m
}

func orZeroInt(m map[string]int) map[string]int {
	if m == nil {
		return map[string]int{}
	}
	return m
}

func orFalse(m map[string]bool) map[string]bool {
	if m == nil {
		return map[string]bool{}
	}
	return m
}

func (e *Engine) Start() *Session {
	order := e.dag.Order()
	return e.StartWithPath(order)
}

// StartWithPath begins an adaptive session over the given path (topo order).
// It compresses the path into a minimal covering set and picks the first
// question by info gain.
func (e *Engine) StartWithPath(path []*concepts.Concept) *Session {
	return e.StartWithPathFocus(path, nil)
}

// StartWithPathFocus is StartWithPath with the concepts the learner asked for.
//
// The path still carries the whole prerequisite closure, because a placement
// run has to know whether the foundations hold. The focus only raises those
// concepts' information gain, so the probes start inside the chosen areas and
// the foundations are asked once the selection is settled. Report 117 was a
// learner who picked linear algebra, analysis and abstract algebra, and whose
// first question was "what is the opposite of 1": the highest-influence node in
// a 246-concept closure is almost always a root.
func (e *Engine) StartWithPathFocus(path []*concepts.Concept, focusIDs []string) *Session {
	if len(path) == 0 {
		return &Session{State: StateDone, order: path}
	}
	focus := make(map[string]bool, len(focusIDs))
	for _, id := range focusIDs {
		focus[id] = true
	}
	s := &Session{
		State:             StateProbing,
		order:             path,
		beliefs:           make(map[string]float64, len(path)),
		confidence:        make(map[string]float64, len(path)),
		probeCounts:       make(map[string]int, len(path)),
		correctCount:      make(map[string]int, len(path)),
		totalCount:        make(map[string]int, len(path)),
		doneSet:           make(map[string]bool, len(path)),
		supplementalCount: make(map[string]int),
		focus:             focus,
		theta:             0.0,
		se:                1.0,
	}
	for _, c := range path {
		s.beliefs[c.ID] = 0.5
	}
	return s
}

// compressedCover returns a topological path-covering set: roots, terminal targets
// and intermediate nodes stratified by depth and subdomain so parallel curriculum tracks
// are all represented.
func (e *Engine) compressedCover(s *Session) []*concepts.Concept {
	return graphCover(s.order, coverSize)
}

func graphCover(order []*concepts.Concept, targetSize int) []*concepts.Concept {
	if len(order) <= targetSize {
		return order
	}

	inOrder := make(map[string]*concepts.Concept, len(order))
	for _, c := range order {
		inOrder[c.ID] = c
	}

	inDegree := make(map[string]int, len(order))
	children := make(map[string][]string, len(order))
	for _, c := range order {
		for _, p := range c.Prerequisites {
			if _, ok := inOrder[p]; ok {
				inDegree[c.ID]++
				children[p] = append(children[p], c.ID)
			}
		}
	}

	// Compute longest path depth from roots
	depth := make(map[string]int, len(order))
	maxDepth := 0
	for _, c := range order {
		d := depth[c.ID]
		if d > maxDepth {
			maxDepth = d
		}
		for _, ch := range children[c.ID] {
			if depth[ch] < d+1 {
				depth[ch] = d + 1
			}
		}
	}

	var roots []*concepts.Concept
	var leaves []*concepts.Concept
	for _, c := range order {
		if inDegree[c.ID] == 0 {
			roots = append(roots, c)
		}
		if len(children[c.ID]) == 0 {
			leaves = append(leaves, c)
		}
	}

	selected := make(map[string]bool, targetSize)
	var out []*concepts.Concept
	add := func(c *concepts.Concept) {
		if c != nil && !selected[c.ID] {
			selected[c.ID] = true
			out = append(out, c)
		}
	}

	// 1. Add all roots (foundational prerequisites)
	for _, r := range roots {
		add(r)
	}

	// 2. Add all leaves (terminal assessment targets)
	for _, l := range leaves {
		add(l)
	}

	// 3. For parallel tracks and chains, sample intermediate nodes by depth tier and subdomain
	if maxDepth > 1 && len(out) < targetSize {
		tiers := 4
		for tier := 1; tier < tiers; tier++ {
			targetD := (maxDepth * tier) / tiers
			bySub := make(map[string]*concepts.Concept)
			minDist := make(map[string]int)
			for _, c := range order {
				if selected[c.ID] {
					continue
				}
				dist := depth[c.ID] - targetD
				if dist < 0 {
					dist = -dist
				}
				sub := c.Subdomain
				if sub == "" {
					sub = c.Domain
				}
				if cur, ok := minDist[sub]; !ok || dist < cur {
					minDist[sub] = dist
					bySub[sub] = c
				}
			}
			for _, c := range bySub {
				if len(out) >= targetSize {
					break
				}
				add(c)
			}
		}
	}

	// 4. Fill to targetSize if short
	if len(out) < targetSize {
		step := int(math.Ceil(float64(len(order)) / float64(targetSize)))
		if step < 1 {
			step = 1
		}
		for i := 0; i < len(order); i += step {
			add(order[i])
			if len(out) >= targetSize {
				break
			}
		}
	}

	// 5. If roots + leaves exceed targetSize, retain roots and uniform samples
	if len(out) > targetSize {
		var trimmed []*concepts.Concept
		for _, r := range roots {
			trimmed = append(trimmed, r)
		}
		nonRoots := make([]*concepts.Concept, 0, len(out)-len(trimmed))
		for _, c := range out {
			if inDegree[c.ID] > 0 {
				nonRoots = append(nonRoots, c)
			}
		}
		rem := targetSize - len(trimmed)
		if rem > 0 && len(nonRoots) > 0 {
			stride := float64(len(nonRoots)) / float64(rem)
			for i := 0; i < rem; i++ {
				idx := int(float64(i) * stride)
				if idx >= len(nonRoots) {
					idx = len(nonRoots) - 1
				}
				trimmed = append(trimmed, nonRoots[idx])
			}
		}
		out = trimmed
	}

	indexMap := make(map[string]int, len(order))
	for i, c := range order {
		indexMap[c.ID] = i
	}
	sort.Slice(out, func(i, j int) bool {
		return indexMap[out[i].ID] < indexMap[out[j].ID]
	})
	return out
}

func (s *Session) index(id string) int {
	for i, c := range s.order {
		if c.ID == id {
			return i
		}
	}
	return len(s.order)
}

// Prob2PL computes the 2-Parameter Logistic item response probability:
// P_i(theta) = 1 / (1 + exp(-a_i * (theta - b_i))).
func Prob2PL(theta, a, b float64) float64 {
	z := a * (theta - b)
	if z > 35 {
		return 1.0
	}
	if z < -35 {
		return 0.0
	}
	return 1.0 / (1.0 + math.Exp(-z))
}

// FisherInformation computes item information under 2PL IRT:
// I_i(theta) = a_i^2 * P_i(theta) * (1 - P_i(theta)).
func FisherInformation(theta, a, b float64) float64 {
	p := Prob2PL(theta, a, b)
	return a * a * p * (1.0 - p)
}

// ItemParameters calculates the discrimination (a) and difficulty (b) parameters for concept c.
// Difficulty b is calibrated relative to the student's current knowledge frontier.
// Discrimination a scales with graph connectivity, key prerequisites and encompasses metadata.
func (e *Engine) ItemParameters(s *Session, c *concepts.Concept) (float64, float64) {
	if c == nil {
		return 1.0, 0.0
	}
	prereqCount := len(c.Prerequisites)
	depCount := 0
	if e.dag != nil {
		depCount = len(e.dag.DependentsOf(c.ID))
	}
	connectivity := prereqCount + depCount
	if connectivity > 6 {
		connectivity = 6
	}
	a := 1.0 + 0.15*float64(connectivity)
	if len(c.KeyPrerequisites) > 0 {
		a += 0.25
	}
	if len(c.Encompasses) > 0 {
		a += 0.25
	}
	if a < 0.8 {
		a = 0.8
	}
	if a > 2.5 {
		a = 2.5
	}

	// Calibrate difficulty b relative to the knowledge frontier in s.order.
	frontierIdx := 0
	for i, concept := range s.order {
		if s.beliefs[concept.ID] >= beliefThreshold {
			frontierIdx = i
		}
	}
	idx := s.index(c.ID)
	bBase := float64(idx-frontierIdx) / 5.0
	probes := s.probeCounts[c.ID]
	if probes > 1 {
		bBase += 0.3 * float64(probes-1)
	}
	return a, bBase
}

// entropy is the binary entropy of a belief p (uncertainty).
func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -(p*math.Log2(p) + (1-p)*math.Log2(1-p))
}

// infoGain scores a candidate concept using 2PL Fisher Information and belief uncertainty scaled by coverage influence.
// Fisher Information is maximized when item difficulty b matches current latent ability theta.
// Coverage = how many other cover concepts it reaches through prereqs/dependents.
func (e *Engine) infoGain(s *Session, c *concepts.Concept, coverSet map[string]bool) float64 {
	a, b := e.ItemParameters(s, c)
	info := FisherInformation(s.theta, a, b)
	unc := entropy(s.beliefs[c.ID])

	influence := 1.0
	for _, dep := range e.dag.DependentsOf(c.ID) {
		if coverSet[dep.ID] {
			influence += 0.25
		}
	}
	for _, pr := range e.dag.PrereqsOf(c.ID) {
		if coverSet[pr.ID] {
			influence += 0.25
		}
	}
	if s.focus[c.ID] {
		// A concept the learner chose outranks a foundation of equal uncertainty.
		// This is a preference between probes, not a filter: every covered concept
		// is still asked, and a miss anywhere still pulls the frontier backwards.
		// The weight was 0.5 before the 2PL change. At 2.5 a focused probe now
		// outweighs a difference in uncertainty, so a learner who names what they
		// want to be assessed on can pull the sequence towards it. That is
		// deliberate, and it is the one place where this file's ordering is not
		// driven by uncertainty alone.
		influence += 2.5
	}
	return unc * influence * (1.0 + 0.2*info)
}

// pickByInfoGain selects the highest-gain cover concept not yet settled.
func (e *Engine) pickByInfoGain(s *Session) string {
	cover := e.compressedCover(s)
	coverSet := make(map[string]bool, len(cover))
	for _, c := range cover {
		coverSet[c.ID] = true
	}
	best := ""
	bestScore := -1.0
	for _, c := range cover {
		if s.doneSet[c.ID] {
			continue
		}
		score := e.infoGain(s, c, coverSet)
		if score > bestScore {
			bestScore = score
			best = c.ID
		}
	}
	return best
}

func (e *Engine) NextQuestion(s *Session) (*generator.Problem, string, error) {
	s.Lock()
	defer s.Unlock()
	if s.State == StateDone {
		s.LastProblem = nil
		return nil, "", nil
	}
	cid := e.pickByInfoGain(s)
	if cid == "" {
		// Cover exhausted: run the supplemental diagnostic for
		// low-confidence concepts before declaring done.
		cid = e.supplementalDiagnostic(s)
	}
	if cid == "" {
		s.State = StateDone
		s.LastProblem = nil
		return nil, "", nil
	}
	concept := s.order[s.index(cid)]
	s.probeCounts[cid]++
	// Vary difficulty: start low, increase on later probes.
	difficulty := 0.3
	if s.probeCounts[cid] >= 2 {
		difficulty = 0.6
	}
	if s.probeCounts[cid] >= 3 {
		difficulty = 0.8
	}
	p, err := e.registry.Generate(cid, difficulty)
	if err != nil {
		return nil, "", err
	}
	// Keep the superseded pending question: a retry voids its attempt and
	// takes the student back to it (responses always stage the next one).
	s.prevProblem, s.prevConceptID, s.prevConceptName = s.LastProblem, s.LastConceptID, s.LastConceptName
	s.LastProblem = &p
	s.LastConceptID = cid
	s.LastConceptName = concept.Label
	s.retryUsed = false
	return s.LastProblem, cid, nil
}

// retryEligibleForLocked reports whether a recorded miss looks like a slip
// worth offering back. Any one of three triggers qualifies:
//  1. established knowledge, pre-answer belief at or above the confidence
//     threshold (the model is surprised);
//  2. level drop, the staged next concept sits earlier in path order than
//     the missed one (the miss demoted the student; heuristic: path order is
//     topological, foundations first, not a strict difficulty scale);
//  3. really simple, the missed question was served at base difficulty
//     (first probe, 0.3 tier).
//
// Admits ("I don't know") never qualify: an admit is not a slip.
// Caller must hold s.Lock. Genuine unknowns on hard material get no offer.
func (e *Engine) retryEligibleForLocked(s *Session, missedCID, nextCID string) bool {
	if s.State == StateDone || s.retryUsed {
		return false
	}
	n := len(s.Attempts)
	if n == 0 {
		return false
	}
	last := s.Attempts[n-1]
	if last.ConceptID != missedCID || last.Correct || last.DontKnow {
		return false
	}
	if prior, ok := last.BeliefSnapshot[last.ConceptID]; ok && prior >= confidenceThreshold {
		return true
	}
	if nextCID != "" && s.index(nextCID) < s.index(missedCID) {
		return true
	}
	return s.probeCounts[missedCID] <= 1
}

// RetryAvailableFor reports whether the missed concept's latest miss may be
// retried ("silly mistake" offer), given the staged next concept ("" when
// none is staged yet). Safe for concurrent use.
func (e *Engine) RetryAvailableFor(s *Session, missedCID, nextCID string) bool {
	s.Lock()
	defer s.Unlock()
	return e.retryEligibleForLocked(s, missedCID, nextCID)
}

// RetryServe voids the just-superseded question's recorded miss for a
// "silly mistake" retry and restores that question as pending: the slip is
// popped, counts decremented, beliefs restored from the pre-answer snapshot,
// and doneSet recomputed. Only the superseded pending question can come
// back, only when its miss passes the methodology gate, and only once
// (retryUsed). Returns the restored problem, concept ID and name.
func (e *Engine) RetryServe(s *Session, conceptID string) (*generator.Problem, string, string, error) {
	s.Lock()
	defer s.Unlock()
	if s.State == StateDone {
		return nil, "", "", fmt.Errorf("diagnostic session already complete")
	}
	if conceptID == "" {
		return nil, "", "", fmt.Errorf("concept_id required")
	}
	if s.retryUsed {
		return nil, "", "", fmt.Errorf("retry already used for this question")
	}
	n := len(s.Attempts)
	if n == 0 || s.prevProblem == nil {
		return nil, "", "", fmt.Errorf("nothing to retry")
	}
	last := s.Attempts[n-1]
	if last.ConceptID != conceptID || last.ConceptID != s.prevConceptID {
		return nil, "", "", fmt.Errorf("latest attempt does not match the retryable question")
	}
	if last.Correct {
		return nil, "", "", fmt.Errorf("correct answers cannot be retried")
	}
	if !e.retryEligibleForLocked(s, last.ConceptID, s.LastConceptID) {
		return nil, "", "", fmt.Errorf("retry not available: the miss does not look like a slip")
	}
	s.Attempts = s.Attempts[:n-1]
	s.totalCount[last.ConceptID]--
	s.totalAsked--
	s.beliefs = last.BeliefSnapshot
	s.doneSet[last.ConceptID] = s.totalCount[last.ConceptID] >= 2
	e.updateThetaLocked(s)
	s.LastProblem, s.LastConceptID, s.LastConceptName = s.prevProblem, s.prevConceptID, s.prevConceptName
	s.prevProblem, s.prevConceptID, s.prevConceptName = nil, "", ""
	s.retryUsed = true
	return s.LastProblem, s.LastConceptID, s.LastConceptName, nil
}

// SettleCurrent marks the current question's concept as settled without
// recording an answer, the skip escape hatch for questions that can't be
// answered (bad content, repeated submit failures). No belief update, no
// attempt, no XP: settling (rather than merely re-asking) is what moves CAT
// past a poisoned transition, since NextQuestion is deterministic on state.
// Returns the settled concept ID ("" if nothing pending).
func (e *Engine) SettleCurrent(s *Session) string {
	s.Lock()
	defer s.Unlock()
	cid := s.LastConceptID
	if cid == "" || s.LastProblem == nil {
		return ""
	}
	s.doneSet[cid] = true
	return cid
}

// RecordAnswer processes an answer: belief update + evidence propagation +
// confidence update + supplemental + stop conditions.
func (e *Engine) RecordAnswer(s *Session, conceptID string, correct, fast bool) {
	elapsed := 0.0
	if fast {
		elapsed = -1 // marker: fast without measured elapsed (legacy callers/tests)
	} else {
		elapsed = -2 // marker: not-fast without measured elapsed
	}
	e.RecordAnswerTimed(s, conceptID, correct, elapsed, 0)
}

// RecordAnswerTimed is the MA-parity path: elapsed seconds and the
// per-concept time threshold drive automaticity weighting. Correct but
// excessively slow answers get diminished weight (MA: "higher likelihood
// the student has not yet learned the topic well enough"), while fast
// correct answers keep the +0.1 automaticity bonus. Incorrect answers are
// unaffected by timing. elapsed < 0 preserves legacy fast/not-fast-only
// behavior (legacy tests; the goal diagnostic is the only route).
func (e *Engine) RecordAnswerTimed(s *Session, conceptID string, correct bool, elapsed, timeThresh float64) {
	e.recordTimed(s, conceptID, correct, elapsed, timeThresh, false)
}

// RecordDontKnow records an admitted unknown as negative evidence through
// the standard incorrect path (belief drop, negative propagation, normal
// 2-probe settle). An admit is never "fast": speed bonuses apply to correct
// answers only. Like skip, it carries no XP; unlike skip, it counts.
func (e *Engine) RecordDontKnow(s *Session, conceptID string, elapsed, timeThresh float64) {
	e.recordTimed(s, conceptID, false, elapsed, timeThresh, true)
}

func (e *Engine) recordTimed(s *Session, conceptID string, correct bool, elapsed, timeThresh float64, dontKnow bool) {
	s.Lock()
	defer s.Unlock()
	fast := false
	if elapsed < 0 {
		fast = elapsed == -1
	} else {
		thresh := timeThresh
		if thresh <= 0 {
			thresh = 10.0
		}
		fast = elapsed < thresh
	}
	if dontKnow {
		fast = false
	}
	snapshot := make(map[string]float64, len(s.beliefs))
	for id, b := range s.beliefs {
		snapshot[id] = b
	}
	s.Attempts = append(s.Attempts, Attempt{
		ConceptID:      conceptID,
		Correct:        correct,
		Fast:           fast,
		ElapsedSeconds: elapsed,
		Timestamp:      time.Now().UTC(),
		DontKnow:       dontKnow,
		BeliefSnapshot: snapshot,
	})
	s.totalCount[conceptID]++
	s.totalAsked++
	if correct {
		s.correctCount[conceptID]++
	}

	// Belief update for the concept itself (MA parity: slow correct answers
	// get diminished weight, "higher likelihood the student has not yet
	// learned the topic well enough to build on it").
	slowFactor := 1.0
	if correct && elapsed >= 0 {
		thresh := timeThresh
		if thresh <= 0 {
			thresh = 10.0
		}
		switch {
		case elapsed > 4*thresh:
			slowFactor = 0.25
		case elapsed > 2*thresh:
			slowFactor = 0.5
		}
	}
	b := s.beliefs[conceptID]
	if correct {
		b += 0.35 * slowFactor * (1 - b)
		if fast {
			b += 0.1 * (1 - b)
		}
	} else {
		b *= 0.55
	}
	if b < 0.05 {
		b = 0.05
	}
	if b > 0.95 {
		b = 0.95
	}
	s.beliefs[conceptID] = b

	// Evidence propagation with Bayes conflict weighting: correct answers
	// lend positive evidence to prerequisites (diminished when slow);
	// incorrect answers lend negative evidence to dependents and strip
	// barely-passing "conditional" credit from prerequisites.
	e.propagateEvidence(s, conceptID, correct, slowFactor)

	// Update latent ability theta and measurement error SE via 2PL IRT MAP estimation.
	e.updateThetaLocked(s)

	// Confidence: 2 probes settle a concept.
	s.confidence[conceptID] = math.Min(1.0, float64(s.probeCounts[conceptID])*0.5)

	// Settle concepts at 2 probes; low-confidence ones are revisited by
	// supplementalDiagnostic (up to maxSupplementalProbes extra each).
	s.doneSet[conceptID] = s.totalCount[conceptID] >= 2

	if s.shouldStop() {
		s.State = StateDone
	}
}

type nodeDistance struct {
	id   string
	dist int
}

func (e *Engine) ancestorsWithDistance(conceptID string) []nodeDistance {
	visited := map[string]int{conceptID: 0}
	queue := []nodeDistance{{id: conceptID, dist: 0}}
	var out []nodeDistance
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, pr := range e.dag.PrereqsOf(curr.id) {
			if pr.ID == curr.id {
				continue
			}
			if _, seen := visited[pr.ID]; !seen {
				d := curr.dist + 1
				visited[pr.ID] = d
				nd := nodeDistance{id: pr.ID, dist: d}
				queue = append(queue, nd)
				out = append(out, nd)
			}
		}
	}
	return out
}

func (e *Engine) descendantsWithDistance(conceptID string) []nodeDistance {
	visited := map[string]int{conceptID: 0}
	queue := []nodeDistance{{id: conceptID, dist: 0}}
	var out []nodeDistance
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, dep := range e.dag.DependentsOf(curr.id) {
			if dep.ID == curr.id {
				continue
			}
			if _, seen := visited[dep.ID]; !seen {
				d := curr.dist + 1
				visited[dep.ID] = d
				nd := nodeDistance{id: dep.ID, dist: d}
				queue = append(queue, nd)
				out = append(out, nd)
			}
		}
	}
	return out
}

// propagateEvidence applies ±evidence weights along DAG edges with
// conflictDiscount damping both directions when prior evidence conflicts.
func (e *Engine) propagateEvidence(s *Session, conceptID string, correct bool, slowFactor float64) {
	if correct {
		ancestors := e.ancestorsWithDistance(conceptID)
		for _, pr := range ancestors {
			decay := math.Pow(0.8, float64(pr.dist-1))
			b := s.beliefs[pr.id]
			b += evidenceCorrectWeight * slowFactor * decay * (1 - b) * conflictDiscount
			if b > 0.95 {
				b = 0.95
			}
			s.beliefs[pr.id] = b
			s.confidence[pr.id] = math.Max(s.confidence[pr.id], math.Min(0.65, s.confidence[pr.id]+0.15*decay*slowFactor))
		}
		return
	}
	descendants := e.descendantsWithDistance(conceptID)
	for _, dep := range descendants {
		decay := math.Pow(0.8, float64(dep.dist-1))
		b := s.beliefs[dep.id]
		b *= (1 - evidenceIncorrectWeight*conflictDiscount*decay)
		if b < 0.05 {
			b = 0.05
		}
		s.beliefs[dep.id] = b
	}
	// Fall-back: a miss also strips conditionally completed credit
	// from barely-passed prerequisites.
	ancestors := e.ancestorsWithDistance(conceptID)
	for _, pr := range ancestors {
		if pb := s.beliefs[pr.id]; pb >= beliefThreshold && pb < beliefThreshold+0.15 {
			s.beliefs[pr.id] = pb * conditionalStripFactor
		}
	}
}

// supplementalDiagnostic re-probes settled concepts when confidence remains
// below confidenceThreshold (0.70). It prioritizes the knowledge frontier concept,
// then examines covering set concepts.
// Caller must hold s.Lock (NextQuestion does).
func (e *Engine) supplementalDiagnostic(s *Session) string {
	// Re-probe the knowledge frontier when its confidence remains below 0.70.
	frontierIdx := -1
	for i, c := range s.order {
		if s.beliefs[c.ID] >= beliefThreshold {
			frontierIdx = i
		}
	}
	if frontierIdx >= 0 && frontierIdx < len(s.order) {
		fc := s.order[frontierIdx]
		if s.confidence[fc.ID] < confidenceThreshold && s.supplementalCount[fc.ID] < maxSupplementalProbes {
			s.supplementalCount[fc.ID]++
			delete(s.doneSet, fc.ID)
			return fc.ID
		}
	}

	cover := e.compressedCover(s)
	for _, c := range cover {
		if !s.doneSet[c.ID] {
			continue
		}
		if s.confidence[c.ID] >= confidenceThreshold {
			continue
		}
		if s.supplementalCount[c.ID] >= maxSupplementalProbes {
			continue
		}
		s.supplementalCount[c.ID]++
		delete(s.doneSet, c.ID)
		return c.ID
	}
	return ""
}

func (s *Session) shouldStop() bool {
	if s.totalAsked >= maxTotalQuestions {
		return true
	}
	cover := s.compressedCoverSafe()
	coverDone := true
	for _, c := range cover {
		if !s.doneSet[c.ID] {
			coverDone = false
			break
		}
	}
	if coverDone && s.totalAsked >= minTotalQuestions {
		return true
	}
	return false
}

// compressedCoverSafe mirrors compressedCover without an Engine receiver
// (session-only helper for stop conditions).
func (s *Session) compressedCoverSafe() []*concepts.Concept {
	return graphCover(s.order, coverSize)
}

func (e *Engine) IsComplete(s *Session) bool {
	s.Lock()
	defer s.Unlock()
	return s.State == StateDone
}

// Progress reports the state of the bar a learner is watching. The budget is
// a ceiling, not a promise: a run ends when every concept in the cover has
// settled, which can happen before the budget is spent. At that point the
// denominator is clamped to what was answered, so the bar reads full instead
// of stopping short on the last question.
func (e *Engine) Progress(s *Session) Progress {
	s.Lock()
	defer s.Unlock()
	answered := s.totalAsked
	cover := s.compressedCoverSafeLocked()
	coverDone := 0
	for _, c := range cover {
		if s.doneSet[c.ID] {
			coverDone++
		}
	}
	remaining := len(cover) - coverDone
	est := answered + remaining
	if est < minTotalQuestions {
		est = minTotalQuestions
	}
	if est > maxTotalQuestions {
		est = maxTotalQuestions
	}
	if est < answered {
		est = answered
	}
	budget := 2 * len(cover)
	if budget > maxTotalQuestions {
		budget = maxTotalQuestions
	}
	if s.State == StateDone {
		budget = answered
	}
	return Progress{
		Answered:       answered,
		EstimatedTotal: est,
		MinTotal:       minTotalQuestions,
		MaxTotal:       maxTotalQuestions,
		ProbeBudget:    budget,
		Done:           s.State == StateDone,
	}
}

// compressedCoverSafeLocked is compressedCoverSafe without locking
// (caller must hold s.Lock).
func (s *Session) compressedCoverSafeLocked() []*concepts.Concept {
	return s.compressedCoverSafe()
}

// FrontierEstimate returns the highest order index whose belief is known.
func (e *Engine) FrontierEstimate(s *Session) int {
	s.Lock()
	defer s.Unlock()
	idx := -1
	for i, c := range s.order {
		if s.beliefs[c.ID] >= beliefThreshold {
			idx = i
		}
	}
	return idx
}

// KnowledgeConfidence returns per-concept confidence 0-1 (copy, lock-safe).
func (e *Engine) KnowledgeConfidence(s *Session) map[string]float64 {
	s.Lock()
	defer s.Unlock()
	out := make(map[string]float64, len(s.confidence))
	for k, v := range s.confidence {
		out[k] = v
	}
	return out
}

// updateThetaLocked estimates latent ability theta via Newton-Raphson MAP with standard normal prior.
// It also updates measurement precision se = 1 / sqrt(total Fisher Information + 1).
func (e *Engine) updateThetaLocked(s *Session) {
	if len(s.Attempts) == 0 {
		s.theta = 0.0
		s.se = 1.0
		return
	}
	theta := s.theta
	for iter := 0; iter < 5; iter++ {
		firstDeriv := -theta
		secondDeriv := -1.0
		for _, att := range s.Attempts {
			var c *concepts.Concept
			if e.dag != nil {
				c = e.dag.Concept(att.ConceptID)
			}
			a, b := e.ItemParameters(s, c)
			p := Prob2PL(theta, a, b)
			u := 0.0
			if att.Correct {
				u = 1.0
				if att.ElapsedSeconds > 0 {
					thresh := 10.0
					if att.ElapsedSeconds > 4*thresh {
						u = 0.6
					} else if att.ElapsedSeconds > 2*thresh {
						u = 0.75
					}
				}
			}
			firstDeriv += a * (u - p)
			secondDeriv -= a * a * p * (1.0 - p)
		}
		delta := -firstDeriv / secondDeriv
		theta += delta
		if theta < -3.5 {
			theta = -3.5
		}
		if theta > 3.5 {
			theta = 3.5
		}
		if math.Abs(delta) < 1e-4 {
			break
		}
	}
	s.theta = theta
	var totalInfo float64
	for _, att := range s.Attempts {
		var c *concepts.Concept
		if e.dag != nil {
			c = e.dag.Concept(att.ConceptID)
		}
		a, b := e.ItemParameters(s, c)
		totalInfo += FisherInformation(theta, a, b)
	}
	s.se = 1.0 / math.Sqrt(totalInfo+1.0)
}

// AbilityEstimate returns the current estimated latent ability theta and its standard error SE.
func (e *Engine) AbilityEstimate(s *Session) (float64, float64) {
	s.Lock()
	defer s.Unlock()
	return s.theta, s.se
}
