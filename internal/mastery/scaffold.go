package mastery

// RequiredKPStreak is the count of consecutive correct answers required to clear
// a Knowledge Point before advancing to the next.
// Justin Skycak, The Math Academy Way, Chapter 4 and Chapter 14.
const RequiredKPStreak = 2

// KPStateMachine tracks progress through a sequence of Knowledge Points.
type KPStateMachine struct {
	CurrentOrdinal int  `json:"current_ordinal"`
	CurrentStreak  int  `json:"current_streak"`
	TotalKPs       int  `json:"total_kps"`
	Completed      bool `json:"completed"`
}

// NewKPStateMachine initializes a state machine for a concept with totalKPs Knowledge Points.
func NewKPStateMachine(totalKPs int) *KPStateMachine {
	if totalKPs <= 0 {
		return &KPStateMachine{
			CurrentOrdinal: 0,
			CurrentStreak:  0,
			TotalKPs:       0,
			Completed:      true,
		}
	}
	return &KPStateMachine{
		CurrentOrdinal: 1,
		CurrentStreak:  0,
		TotalKPs:       totalKPs,
		Completed:      false,
	}
}

// Transition evaluates an attempt.
// It returns advanced (true if this attempt advanced to the next Knowledge Point or completed)
// and completed (true if all Knowledge Points are now completed).
func (sm *KPStateMachine) Transition(correct bool) (bool, bool) {
	if sm == nil || sm.TotalKPs <= 0 || sm.Completed {
		return false, sm != nil && sm.Completed
	}
	if !correct {
		sm.CurrentStreak = 0
		return false, false
	}
	sm.CurrentStreak++
	if sm.CurrentStreak >= RequiredKPStreak {
		if sm.CurrentOrdinal >= sm.TotalKPs {
			sm.Completed = true
			return true, true
		}
		sm.CurrentOrdinal++
		sm.CurrentStreak = 0
		return true, false
	}
	return false, false
}
