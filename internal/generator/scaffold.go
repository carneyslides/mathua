package generator

// ScaffoldStep represents an intermediate verifiable subgoal in a composite problem.
type ScaffoldStep struct {
	Ordinal       int    `json:"ordinal"`
	Prompt        string `json:"prompt"`
	Answer        string `json:"answer"`
	Explanation   string `json:"explanation"`
	PrereqConcept string `json:"prereq_concept"`
	AnswerHint    string `json:"answer_hint,omitempty"`
	InputMode     string `json:"input_mode,omitempty"`
}

// ScaffoldProblem represents a generator that decomposes composite problems into interactive subgoals.
type ScaffoldProblem interface {
	Generator
	ScaffoldSteps(ctx GeneratorContext) []ScaffoldStep
}
