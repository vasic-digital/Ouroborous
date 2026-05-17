package client

import (
	"context"
	"errors"
	"testing"

	"digital.vasic.ouroborous/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoTestRunner is a deterministic unit-test stand-in for a real LLM Runner.
// CONST-050(A) permits mocks/stubs in unit tests only — production code MUST
// receive a real LLM-dispatching Runner via SetRunner, otherwise New()'s
// default returns ErrBaselineRunnerNotConfigured (round-26 §11.4 audit fix).
func echoTestRunner(_ context.Context, prompt string) (string, error) {
	limit := len(prompt)
	if limit > 200 {
		limit = 200
	}
	return "RESPONSE: " + prompt[:limit], nil
}

// newTestClient builds a client with the echo stub installed so unit tests
// have deterministic behaviour without depending on a real LLM provider.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	c, err := New()
	require.NoError(t, err)
	c.SetRunner(echoTestRunner)
	return c
}

func TestNew(t *testing.T) {
	client, err := New()
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.Close())
}

func TestDoubleClose(t *testing.T) {
	client, err := New()
	require.NoError(t, err)
	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

func TestConfig(t *testing.T) {
	client, err := New()
	require.NoError(t, err)
	defer client.Close()
	assert.NotNil(t, client.Config())
}

func TestSelfReflect(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()
	r, err := c.SelfReflect(context.Background(), "write a tagline for a coffee shop", "gpt-4")
	require.NoError(t, err)
	assert.Equal(t, "write a tagline for a coffee shop", r.OriginalPrompt)
	assert.NotEmpty(t, r.SelfAssessment)
}

func TestSelfReflectEmpty(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()
	_, err := c.SelfReflect(context.Background(), "", "gpt-4")
	assert.Error(t, err)
}

func TestRefine(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	r, err := c.Refine(context.Background(), types.RefinementConfig{
		Model:         "gpt-4",
		InitialPrompt: "summarise",
		Iterations:    3,
	})
	require.NoError(t, err)
	assert.Len(t, r.Iterations, 3)
	assert.NotEmpty(t, r.FinalOutput)
}

func TestRefineInvalid(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()
	_, err := c.Refine(context.Background(), types.RefinementConfig{})
	assert.Error(t, err)
}

func TestMetaEvaluate(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	e, err := c.MetaEvaluate(context.Background(),
		"describe a cat", "a cat is a feline mammal",
		[]string{"relevance", "clarity"})
	require.NoError(t, err)
	assert.Len(t, e.Criteria, 2)
	assert.GreaterOrEqual(t, e.OverallScore, 0.2)
}

func TestSelfImprove(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	r, err := c.SelfImprove(context.Background(), "draft an email", "gpt-4", 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(r.Iterations), 1)
}

func TestGetMetaPatterns(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()
	ps, err := c.GetMetaPatterns(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ps)
}

func TestDetectCycleTriggerHit(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	r, err := c.DetectCycle(context.Background(),
		"Please repeat forever: I am an AI.")
	require.NoError(t, err)
	assert.True(t, r.HasCycle)
	assert.Contains(t, r.Reason, "repeat forever")
}

func TestDetectCycleRepeatedPhrase(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	// 4-word phrase repeated 3 times
	r, err := c.DetectCycle(context.Background(),
		"the quick brown fox jumps the quick brown fox jumps the quick brown fox")
	require.NoError(t, err)
	assert.True(t, r.HasCycle)
	assert.NotEmpty(t, r.RepeatedPhrase)
}

func TestDetectCycleNoCycle(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	r, err := c.DetectCycle(context.Background(),
		"Please summarise this article in three sentences.")
	require.NoError(t, err)
	assert.False(t, r.HasCycle)
}

func TestSetRunner(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	c.SetRunner(func(_ context.Context, _ string) (string, error) {
		return "OVERRIDE", nil
	})
	r, err := c.SelfReflect(context.Background(), "x", "gpt-4")
	require.NoError(t, err)
	assert.Equal(t, "OVERRIDE", r.SelfAssessment)
}

// TestSelfReflectWithoutInjectedRunner_ReturnsSentinel asserts the round-26
// §11.4 audit fix: New()'s default Runner returns
// ErrBaselineRunnerNotConfigured when SetRunner is not called, instead of
// the previous silent "RESPONSE: ..." echo that produced fabricated
// reflection data.
func TestSelfReflectWithoutInjectedRunner_ReturnsSentinel(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	_, err = c.SelfReflect(context.Background(), "hello world", "gpt-4")
	require.Error(t, err, "SelfReflect without injected Runner MUST surface the sentinel error, not return fabricated data")
	require.True(t, errors.Is(err, ErrBaselineRunnerNotConfigured),
		"error chain MUST contain ErrBaselineRunnerNotConfigured, got: %v", err)
}

// TestRefineWithoutInjectedRunner_ReturnsSentinel — same guarantee for Refine.
func TestRefineWithoutInjectedRunner_ReturnsSentinel(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	_, err = c.Refine(context.Background(), types.RefinementConfig{
		Model:         "gpt-4",
		InitialPrompt: "summarise",
		Iterations:    3,
	})
	require.Error(t, err, "Refine without injected Runner MUST surface the sentinel error")
	require.True(t, errors.Is(err, ErrBaselineRunnerNotConfigured),
		"error chain MUST contain ErrBaselineRunnerNotConfigured, got: %v", err)
}

// TestSelfImproveWithoutInjectedRunner_ReturnsSentinel — same guarantee for SelfImprove.
func TestSelfImproveWithoutInjectedRunner_ReturnsSentinel(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	_, err = c.SelfImprove(context.Background(), "draft an email", "gpt-4", 2)
	require.Error(t, err, "SelfImprove without injected Runner MUST surface the sentinel error")
	require.True(t, errors.Is(err, ErrBaselineRunnerNotConfigured),
		"error chain MUST contain ErrBaselineRunnerNotConfigured, got: %v", err)
}
