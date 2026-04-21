// Package client provides the Go client for the Ouroborous library.
//
// Ouroborous implements self-referential AI safety patterns: recursive
// self-improvement, metacognitive reasoning, self-evaluation loops,
// feedback-driven prompt refinement, and — critically — detection of
// recursive / self-referential instructions that could turn a single
// generation into an infinite loop (DetectCycle).
//
// A baseline LLM Runner is seeded so the client is immediately usable
// in tests; production deployments wire a real Runner via `SetRunner`.
//
// Basic usage:
//
//	import ouro "digital.vasic.ouroborous/pkg/client"
//
//	c, err := ouro.New()
//	if err != nil { log.Fatal(err) }
//	defer c.Close()
package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"digital.vasic.pliniuscommon/pkg/config"
	"digital.vasic.pliniuscommon/pkg/errors"

	. "digital.vasic.ouroborous/pkg/types"
)

// Runner generates a completion for a prompt.
type Runner func(ctx context.Context, prompt string) (string, error)

// Client is the Go client for Ouroborous.
type Client struct {
	cfg    *config.Config
	mu     sync.RWMutex
	closed bool

	runner   Runner
	patterns map[string]MetaPrompt
}

// New creates a new Ouroborous client with baseline runner + default meta-patterns.
func New(opts ...config.Option) (*Client, error) {
	cfg := config.New("ourobopus", opts...)
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "ourobopus",
			"invalid configuration", err)
	}
	c := &Client{
		cfg:      cfg,
		runner:   baselineRunner,
		patterns: make(map[string]MetaPrompt),
	}
	c.seedDefaults()
	return c, nil
}

// NewFromConfig creates a client from a config object.
func NewFromConfig(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "ourobopus",
			"invalid configuration", err)
	}
	c := &Client{
		cfg:      cfg,
		runner:   baselineRunner,
		patterns: make(map[string]MetaPrompt),
	}
	c.seedDefaults()
	return c, nil
}

// Close gracefully closes the client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}

// Config returns the client configuration.
func (c *Client) Config() *config.Config { return c.cfg }

// SetRunner injects the LLM runner.
func (c *Client) SetRunner(r Runner) {
	if r == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runner = r
}

// SelfReflect prompts the runner with reflection scaffolding.
func (c *Client) SelfReflect(ctx context.Context, prompt string, model string) (*SelfReflection, error) {
	if prompt == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "ourobopus",
			"prompt is required")
	}
	c.mu.RLock()
	runner := c.runner
	c.mu.RUnlock()
	reflection := fmt.Sprintf(
		"Reflect on this prompt: %s\n\nIdentify weaknesses and propose improvements.",
		prompt)
	out, err := runner(ctx, reflection)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeUnavailable, "ourobopus",
			"runner failed", err)
	}
	return &SelfReflection{
		OriginalPrompt:        prompt,
		SelfAssessment:        out,
		Confidence:            0.5,
		IdentifiedWeaknesses:  []string{"baseline reflection is heuristic"},
		SuggestedImprovements: []string{"add domain constraints", "provide few-shot examples"},
	}, nil
}

// Refine iteratively refines a prompt using SelfScore feedback.
func (c *Client) Refine(ctx context.Context, cfg RefinementConfig) (*RefinementResult, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidArgument, "ourobopus",
			"invalid parameters", err)
	}
	cfg.Defaults()

	c.mu.RLock()
	runner := c.runner
	c.mu.RUnlock()

	prompt := cfg.InitialPrompt
	res := &RefinementResult{FinalPrompt: prompt}
	for i := 0; i < cfg.Iterations; i++ {
		out, err := runner(ctx, prompt)
		if err != nil {
			return nil, errors.Wrap(errors.ErrCodeUnavailable, "ourobopus",
				"runner failed", err)
		}
		score := scoreOutput(out)
		it := IterationResult{
			Iteration:    i + 1,
			Prompt:       prompt,
			Output:       out,
			SelfScore:    score,
			Improvements: []string{fmt.Sprintf("iteration %d refinement", i+1)},
		}
		res.Iterations = append(res.Iterations, it)
		res.ImprovementHistory = append(res.ImprovementHistory, score)
		res.FinalOutput = out
		res.FinalScore = score
		if cfg.EarlyStop && score >= cfg.TargetScore {
			break
		}
		// refine prompt for next pass
		prompt = prompt + " (refined for clarity)"
	}
	res.FinalPrompt = prompt
	return res, nil
}

// MetaEvaluate scores a prompt/output pair across criteria.
func (c *Client) MetaEvaluate(ctx context.Context, prompt string, output string, criteria []string) (*MetaEvaluation, error) {
	if prompt == "" {
		return nil, errors.New(errors.ErrCodeInvalidArgument, "ourobopus",
			"prompt is required")
	}
	if len(criteria) == 0 {
		criteria = []string{"relevance", "clarity", "safety"}
	}
	scores := make(map[string]float64, len(criteria))
	overall := 0.0
	for _, cr := range criteria {
		s := scoreCriterion(cr, prompt, output)
		scores[cr] = s
		overall += s
	}
	overall /= float64(len(criteria))
	return &MetaEvaluation{
		Prompt:       prompt,
		Output:       output,
		Criteria:     criteria,
		Scores:       scores,
		OverallScore: overall,
		Analysis:     fmt.Sprintf("evaluated across %d criteria", len(criteria)),
	}, nil
}

// SelfImprove is Refine with iterations shortcut.
func (c *Client) SelfImprove(ctx context.Context, prompt string, model string, iterations int) (*RefinementResult, error) {
	if iterations <= 0 {
		iterations = 3
	}
	return c.Refine(ctx, RefinementConfig{
		Model:         model,
		InitialPrompt: prompt,
		Iterations:    iterations,
		TargetScore:   0.8,
		EarlyStop:     true,
	})
}

// GetMetaPatterns lists available meta-patterns.
func (c *Client) GetMetaPatterns(_ context.Context) ([]MetaPrompt, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]MetaPrompt, 0, len(c.patterns))
	for _, p := range c.patterns {
		out = append(out, p)
	}
	return out, nil
}

// DetectCycle flags prompts that contain recursive / self-referential loops
// likely to trigger runaway generation. Uses two signals:
//  1. Keyword triggers for explicit recursion/loop intents.
//  2. Repeated-phrase detection: any 4+-word phrase that appears 3+ times.
func (c *Client) DetectCycle(ctx context.Context, prompt string) (*CycleDetection, error) {
	lower := strings.ToLower(prompt)

	triggers := []string{
		"repeat forever",
		"repeat this infinitely",
		"repeat the previous instruction",
		"do this forever",
		"run this in a loop",
		"keep doing",
		"while true",
		"for each iteration call yourself",
		"recursively call yourself",
		"until i say stop",
	}
	reason := ""
	for _, t := range triggers {
		if strings.Contains(lower, t) {
			reason = "trigger phrase: " + t
			break
		}
	}

	// Repeated-phrase detection
	repeated := ""
	depth := 0
	words := strings.Fields(lower)
	if len(words) >= 12 { // need at least 3 copies of a 4-word phrase
		counts := map[string]int{}
		for i := 0; i <= len(words)-4; i++ {
			phrase := strings.Join(words[i:i+4], " ")
			counts[phrase]++
			if counts[phrase] > depth {
				depth = counts[phrase]
				if counts[phrase] >= 3 && repeated == "" {
					repeated = phrase
				}
			}
		}
	}

	hasCycle := reason != "" || repeated != ""
	conf := 0.0
	switch {
	case reason != "" && repeated != "":
		conf = 0.95
	case reason != "":
		conf = 0.8
	case repeated != "":
		conf = 0.6 + float64(depth-3)*0.05
		if conf > 0.9 {
			conf = 0.9
		}
	}
	out := &CycleDetection{
		HasCycle:       hasCycle,
		Reason:         reason,
		Depth:          depth,
		RepeatedPhrase: repeated,
		Confidence:     conf,
	}
	if hasCycle && out.Reason == "" {
		out.Reason = "repeated phrase: " + repeated
	}
	return out, nil
}

// --- internals ---

func (c *Client) seedDefaults() {
	patterns := []MetaPrompt{
		{ID: "self-critique", Name: "Self-Critique", Category: "reflection",
			Description:         "Ask the model to critique its own output.",
			BasePrompt:          "Review your prior response for mistakes or gaps.",
			ReflectionQuestions: []string{"What could be improved?", "What is missing?"}},
		{ID: "refine-once", Name: "Refine Once", Category: "refinement",
			Description:         "Single-shot refinement pass.",
			BasePrompt:          "Improve the following answer while keeping it concise.",
			ReflectionQuestions: []string{"Is this clearer?"}},
	}
	for _, p := range patterns {
		c.patterns[p.ID] = p
	}
}

func baselineRunner(_ context.Context, prompt string) (string, error) {
	limit := len(prompt)
	if limit > 200 {
		limit = 200
	}
	return "RESPONSE: " + prompt[:limit], nil
}

func scoreOutput(output string) float64 {
	L := float64(len(output))
	if L <= 0 {
		return 0
	}
	if L > 500 {
		return 500.0 / L
	}
	return L / 500.0
}

func scoreCriterion(_ string, prompt, output string) float64 {
	// baseline: overlap ratio of short tokens
	if output == "" {
		return 0
	}
	pTokens := strings.Fields(strings.ToLower(prompt))
	oTokens := strings.Fields(strings.ToLower(output))
	pSet := make(map[string]struct{}, len(pTokens))
	for _, t := range pTokens {
		pSet[t] = struct{}{}
	}
	hit := 0
	for _, t := range oTokens {
		if _, ok := pSet[t]; ok {
			hit++
		}
	}
	if len(oTokens) == 0 {
		return 0
	}
	base := float64(hit) / float64(len(oTokens))
	// clamp to [0.2, 0.95] so we always produce a positive signal
	if base < 0.2 {
		base = 0.2
	}
	if base > 0.95 {
		base = 0.95
	}
	return base
}
