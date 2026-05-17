package client

import (
	"context"
	"strings"
	"testing"

	"digital.vasic.ouroborous/pkg/types"
)

func BenchmarkDetectCycle(b *testing.B) {
	c, err := New()
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()
	ctx := context.Background()
	prompt := strings.Repeat("once upon a time ", 4) + "the end."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.DetectCycle(ctx, prompt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRefine(b *testing.B) {
	c, err := New()
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()
	// Per round-26 §11.4 audit: a real / stub Runner MUST be injected before
	// Refine — the default returns ErrBaselineRunnerNotConfigured.
	c.SetRunner(echoTestRunner)
	ctx := context.Background()
	cfg := types.RefinementConfig{Model: "m", InitialPrompt: "hi", Iterations: 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Refine(ctx, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
