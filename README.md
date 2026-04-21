# Ouroborous

Self-referential AI safety patterns: recursive self-improvement,
metacognitive reasoning, self-evaluation loops, feedback-driven prompt
refinement, and — critically — detection of recursive / self-referential
instructions that could turn a single generation into a runaway loop.
Part of the Plinius Go service family used by HelixAgent.

## Status

- Compiles: `go build ./...` exits 0.
- Tests pass under `-race`: 2 packages (types, client), all green.
- Baseline deterministic runner + default meta-patterns
  (`self-critique`, `refine-once`) seeded on `New()`.
- Integration-ready: consumable Go library for the HelixAgent ensemble.

## Purpose

- `pkg/types` — value types: `MetaPrompt`, `SelfReflection` (with
  `Validate`), `IterationResult`, `RefinementConfig`, `RefinementResult`,
  `MetaEvaluation`, `CycleDetection`.
- `pkg/client` — self-referential orchestration:
  - `SelfReflect(prompt, model)` — reflection scaffold
  - `Refine(cfg)` — iterative refinement with per-iteration SelfScore
  - `MetaEvaluate(prompt, output, criteria)` — multi-criterion scoring
  - `SelfImprove(prompt, model, iterations)` — shortcut for `Refine`
  - `GetMetaPatterns()` — lists seeded meta-patterns
  - `DetectCycle(prompt)` — flags runaway-loop patterns (keyword +
    repeated-4-word-phrase detection)
  - `SetRunner(Runner)` — wire in a real LLM

## Usage

```go
import (
    "context"
    "log"

    ouro "digital.vasic.ouroborous/pkg/client"
)

c, err := ouro.New()
if err != nil { log.Fatal(err) }
defer c.Close()

det, err := c.DetectCycle(context.Background(), "Please repeat forever: hello")
if err != nil { log.Fatal(err) }
if det.HasCycle {
    log.Printf("loop risk: %s (conf=%.2f)", det.Reason, det.Confidence)
}
```

## Module path

```go
import "digital.vasic.ouroborous"
```

## Lineage

Extracted from internal HelixAgent research tree on 2026-04-21. The
earlier Python upstream name was obfuscated; this Go port uses a clean
readable name. Graduated to functional status alongside its 7 sibling
Plinius modules.

Historical research corpus (unused) remains at
`docs/research/go-elder-plinius-v3/go-elder-plinius/go-ourobopus/`
inside the HelixAgent repository.

## Development layout

This module's `go.mod` declares the module as `digital.vasic.ouroborous`
and uses a relative `replace` directive pointing at `../PliniusCommon`.

## License

Apache-2.0
