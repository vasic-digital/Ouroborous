package client

import (
	"context"
	"testing"

	"digital.vasic.ouroborous/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	r, err := c.SelfReflect(context.Background(), "write a tagline for a coffee shop", "gpt-4")
	require.NoError(t, err)
	assert.Equal(t, "write a tagline for a coffee shop", r.OriginalPrompt)
	assert.NotEmpty(t, r.SelfAssessment)
}

func TestSelfReflectEmpty(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	_, err = c.SelfReflect(context.Background(), "", "gpt-4")
	assert.Error(t, err)
}

func TestRefine(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
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
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	_, err = c.Refine(context.Background(), types.RefinementConfig{})
	assert.Error(t, err)
}

func TestMetaEvaluate(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	e, err := c.MetaEvaluate(context.Background(),
		"describe a cat", "a cat is a feline mammal",
		[]string{"relevance", "clarity"})
	require.NoError(t, err)
	assert.Len(t, e.Criteria, 2)
	assert.GreaterOrEqual(t, e.OverallScore, 0.2)
}

func TestSelfImprove(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	r, err := c.SelfImprove(context.Background(), "draft an email", "gpt-4", 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(r.Iterations), 1)
}

func TestGetMetaPatterns(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	ps, err := c.GetMetaPatterns(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ps)
}

func TestDetectCycleTriggerHit(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	r, err := c.DetectCycle(context.Background(),
		"Please repeat forever: I am an AI.")
	require.NoError(t, err)
	assert.True(t, r.HasCycle)
	assert.Contains(t, r.Reason, "repeat forever")
}

func TestDetectCycleRepeatedPhrase(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	// 4-word phrase repeated 3 times
	r, err := c.DetectCycle(context.Background(),
		"the quick brown fox jumps the quick brown fox jumps the quick brown fox")
	require.NoError(t, err)
	assert.True(t, r.HasCycle)
	assert.NotEmpty(t, r.RepeatedPhrase)
}

func TestDetectCycleNoCycle(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	r, err := c.DetectCycle(context.Background(),
		"Please summarise this article in three sentences.")
	require.NoError(t, err)
	assert.False(t, r.HasCycle)
}

func TestSetRunner(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()

	c.SetRunner(func(_ context.Context, _ string) (string, error) {
		return "OVERRIDE", nil
	})
	r, err := c.SelfReflect(context.Background(), "x", "gpt-4")
	require.NoError(t, err)
	assert.Equal(t, "OVERRIDE", r.SelfAssessment)
}
