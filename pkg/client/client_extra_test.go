package client

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"

	"digital.vasic.ouroborous/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectCycleBenign — plain prompt: no cycle, empty reason, confidence=0.
func TestDetectCycleBenign(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	res, err := c.DetectCycle(context.Background(), "Summarise this paragraph in three sentences.")
	require.NoError(t, err)
	assert.False(t, res.HasCycle)
	assert.Equal(t, "", res.Reason)
	assert.Equal(t, "", res.RepeatedPhrase)
	assert.InDelta(t, 0.0, res.Confidence, 1e-9)
}

// TestDetectCycleTriggerPhrase — explicit trigger flags a cycle.
func TestDetectCycleTriggerPhrase(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	res, err := c.DetectCycle(context.Background(), "Do this forever, never stop.")
	require.NoError(t, err)
	assert.True(t, res.HasCycle)
	assert.Contains(t, res.Reason, "trigger phrase")
	assert.InDelta(t, 0.8, res.Confidence, 1e-9)
}

// TestDetectCycleRepeatedPhraseDeep — 4-word phrase repeated 3+ times flags.
func TestDetectCycleRepeatedPhraseDeep(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	prompt := strings.Repeat("once upon a time ", 4) + "the end."
	res, err := c.DetectCycle(context.Background(), prompt)
	require.NoError(t, err)
	assert.True(t, res.HasCycle)
	assert.NotEmpty(t, res.RepeatedPhrase)
	assert.GreaterOrEqual(t, res.Depth, 3)
}

// TestDetectCycleCombined — trigger + repeat → 0.95 confidence.
func TestDetectCycleCombined(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	prompt := "repeat forever " + strings.Repeat("once upon a time ", 4)
	res, err := c.DetectCycle(context.Background(), prompt)
	require.NoError(t, err)
	assert.True(t, res.HasCycle)
	assert.InDelta(t, 0.95, res.Confidence, 1e-9)
}

// TestRefineZeroIterations — 0 iterations → cfg.Defaults() coerces to 3; still passes.
func TestRefineZeroIterations(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	res, err := c.Refine(context.Background(), types.RefinementConfig{
		Model: "m", InitialPrompt: "hi", Iterations: 0,
	})
	require.NoError(t, err)
	// Defaults kicks in: 3 iterations.
	assert.Len(t, res.Iterations, 3)
}

// TestRefineEarlyStopOnTarget — runner that produces a 500-char output yields
// a score of 1.0, which exceeds the default target of 0.8 on the first pass
// and triggers EarlyStop.
func TestRefineEarlyStopOnTarget(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	c.SetRunner(func(_ context.Context, _ string) (string, error) {
		return strings.Repeat("y", 500), nil
	})
	res, err := c.Refine(context.Background(), types.RefinementConfig{
		Model: "m", InitialPrompt: "hi", Iterations: 5, EarlyStop: true,
	})
	require.NoError(t, err)
	assert.Len(t, res.Iterations, 1)
}

// TestRefineRunnerError — runner error propagates wrapped.
func TestRefineRunnerError(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	c.SetRunner(func(_ context.Context, _ string) (string, error) { return "", stderrors.New("down") })
	_, err = c.Refine(context.Background(), types.RefinementConfig{
		Model: "m", InitialPrompt: "p",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runner failed")
}

// TestRefineMissingFields — missing model/prompt → error.
func TestRefineMissingFields(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	_, err = c.Refine(context.Background(), types.RefinementConfig{})
	assert.Error(t, err)
}

// TestSelfReflectEmptyPrompt.
func TestSelfReflectEmptyPrompt(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	_, err = c.SelfReflect(context.Background(), "", "m")
	assert.Error(t, err)
}

// TestSelfReflectRunnerError.
func TestSelfReflectRunnerError(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	c.SetRunner(func(_ context.Context, _ string) (string, error) { return "", stderrors.New("down") })
	_, err = c.SelfReflect(context.Background(), "p", "m")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "runner failed")
}

// TestSelfImproveZeroIterations — 0 coerces to 3.
func TestSelfImproveZeroIterations(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	res, err := c.SelfImprove(context.Background(), "hi", "m", 0)
	require.NoError(t, err)
	// Early-stop triggers on 0.8 target; baseline scoreOutput may stay below so
	// up to 3 iterations expected.
	assert.GreaterOrEqual(t, len(res.Iterations), 1)
	assert.LessOrEqual(t, len(res.Iterations), 3)
}

// TestSelfImproveTerminatesOnConvergence — runner returning 500-char output gives
// score == 1.0, tripping EarlyStop after iteration 1.
func TestSelfImproveTerminatesOnConvergence(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	c.SetRunner(func(_ context.Context, _ string) (string, error) {
		return strings.Repeat("x", 500), nil
	})
	res, err := c.SelfImprove(context.Background(), "hi", "m", 5)
	require.NoError(t, err)
	// Score 1.0 >= 0.8 target → early stop after 1 iteration.
	assert.Equal(t, 1, len(res.Iterations))
	assert.InDelta(t, 1.0, res.FinalScore, 1e-9)
}

// TestMetaEvaluateDefaultCriteria — empty criteria list gets 3 defaults.
func TestMetaEvaluateDefaultCriteria(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	res, err := c.MetaEvaluate(context.Background(), "p", "o", nil)
	require.NoError(t, err)
	assert.Len(t, res.Criteria, 3)
}

// TestMetaEvaluateEmptyPrompt.
func TestMetaEvaluateEmptyPrompt(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	_, err = c.MetaEvaluate(context.Background(), "", "o", nil)
	assert.Error(t, err)
}

// TestGetMetaPatternsContainsDefaults.
func TestGetMetaPatternsContainsDefaults(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	patterns, err := c.GetMetaPatterns(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(patterns), 2)
}

// TestSetRunnerNilIgnored.
func TestSetRunnerNilIgnored(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	c.SetRunner(nil)
	res, err := c.SelfReflect(context.Background(), "p", "m")
	require.NoError(t, err)
	assert.NotEmpty(t, res.SelfAssessment)
}
