# Ouroborous

Self-referential AI safety patterns: recursive self-improvement,
metacognitive reasoning, self-evaluation loops, feedback-driven prompt
refinement, and — critically — detection of recursive / self-referential
instructions that could turn a single generation into a runaway loop.
Part of the Plinius Go service family used by the HelixAgent ensemble.

Module path: `digital.vasic.ouroborous` — two packages: `pkg/types`
(value types) and `pkg/client` (orchestration).

## Status

- **FUNCTIONAL** — both packages ship tested implementations.
- `go test -race -count=1 ./...` is green (round-269 evidence in
  `docs/test-coverage.md`).
- Default library seeded on `New()`: 2 meta-patterns (`self-critique`,
  `refine-once`) and a 10-trigger + 4-word-repeated-phrase
  `DetectCycle` heuristic.
- Default `Runner` returns `ErrBaselineRunnerNotConfigured` —
  callers MUST inject a real LLM-dispatching `Runner` via
  `SetRunner` before invoking `SelfReflect` / `Refine` /
  `MetaEvaluate` / `SelfImprove` (round-26 §11.4 audit fix; do
  not regress to the silent "RESPONSE: …" echo).
- `DetectCycle` is heuristic-only and does NOT call the Runner —
  it is safe to call on a fresh client with no Runner injected.
- Integration-ready: consumable Go library for HelixAgent.

## Public surface

`pkg/types` — value types: `MetaPrompt`, `SelfReflection`,
`CycleDetection`, `IterationResult`, `RefinementConfig`,
`RefinementResult`, `MetaEvaluation`. Each carries a `Validate()`
where applicable; `RefinementConfig.Defaults()` installs
`Iterations=3, TargetScore=0.8` when zero.

`pkg/client` — orchestration:

- `New(opts ...config.Option) (*Client, error)` /
  `NewFromConfig(cfg *config.Config) (*Client, error)`
- `(*Client).Close() error` / `(*Client).Config() *config.Config`
- `(*Client).SetRunner(r Runner)` — inject the real LLM-dispatching
  function used by SelfReflect / Refine / MetaEvaluate / SelfImprove.
- `(*Client).SelfReflect(ctx, prompt, model) (*SelfReflection, error)`
- `(*Client).Refine(ctx, RefinementConfig) (*RefinementResult, error)`
- `(*Client).MetaEvaluate(ctx, prompt, output, criteria) (*MetaEvaluation, error)`
- `(*Client).SelfImprove(ctx, prompt, model, iterations) (*RefinementResult, error)`
- `(*Client).GetMetaPatterns(ctx) ([]MetaPrompt, error)`
- `(*Client).DetectCycle(ctx, prompt) (*CycleDetection, error)`
- Sentinel `ErrBaselineRunnerNotConfigured`

## Usage

```go
import (
    "context"
    "log"

    ouro "digital.vasic.ouroborous/pkg/client"
    "digital.vasic.ouroborous/pkg/types"
)

c, err := ouro.New()
if err != nil { log.Fatal(err) }
defer c.Close()

// REQUIRED for SelfReflect / Refine / MetaEvaluate / SelfImprove —
// without this, those four methods return ErrBaselineRunnerNotConfigured
// (round-26 §11.4 audit fix — no fabricated reflection by default).
c.SetRunner(func(ctx context.Context, prompt string) (string, error) {
    return provider.Complete(ctx, prompt)
})

// Cycle detection — no Runner needed; heuristic-only.
det, _ := c.DetectCycle(context.Background(), "Please repeat forever: hello")
if det.HasCycle {
    log.Printf("loop risk: %s (conf=%.2f)", det.Reason, det.Confidence)
}

// Refinement loop with per-iteration self-score.
r, _ := c.Refine(context.Background(), types.RefinementConfig{
    Model:         "gpt-4",
    InitialPrompt: "Summarise modern AI safety research.",
    Iterations:    3,
    TargetScore:   0.8,
    EarlyStop:     true,
})
log.Println(r.FinalOutput, r.FinalScore)
```

## Anti-bluff guarantees (round-269)

Every PASS produced by this submodule's tests + Challenges carries
positive runtime evidence per Article XI §11.9 and the verbatim
2026-05-19 operator mandate:

> "all existing tests and Challenges do work in anti-bluff manner —
> they MUST confirm that all tested codebase really works as
> expected! We had been in position that all tests do execute with
> success and all Challenges as well, but in reality the most of
> the features does not work and can't be used! This MUST NOT be
> the case and execution of tests and Challenges MUST guarantee
> the quality, the completition and full usability by end users
> of the product!"

Seven invariants enforced by the round-269 runner +
`ouroborous_describe_challenge.sh` paired-mutation gate:

1. **Default-surface coverage.** `client.New` MUST seed 2 meta-patterns
   (`self-critique`, `refine-once`). Each is retrieved via
   `GetMetaPatterns` and individually validated through
   `MetaPrompt.Validate`. `Close` MUST be idempotent.
2. **SelfReflect byte-exact round-trip.** `SelfReflect` MUST dispatch
   the input prompt to the injected `Runner` byte-exact and preserve
   non-ASCII bytes intact through 5 locales (en, sr Cyrillic, ja
   Japanese, ar Arabic RTL, zh-CN Han). The capturing Runner records
   the most-recent prompt and the runner asserts `strings.Contains` +
   rune-count + "OUT:" round-trip marker.
3. **Refine per-iteration byte preservation.** `Refine(..., Iterations:3)`
   MUST produce exactly 3 `IterationResult` records per locale, each
   carrying a `Prompt` + `Output` that still contains the locale's
   prompt bytes (refinement-marker append is permitted but MUST NOT
   drop or mutate the original payload). `FinalScore` MUST land in
   `(0, 1]`; capturing Runner dispatch count MUST equal 3.
4. **MetaEvaluate criteria contract.** `MetaEvaluate(prompt, output,
   [3 criteria])` MUST return `Scores` map of size 3 (one entry per
   criterion), `Prompt` + `Output` byte-exact, `Criteria` preserved
   in order, `OverallScore` in `[0, 1]`.
5. **SelfImprove EarlyStop semantics.** `SelfImprove(prompt, model, 2)`
   MUST record at least 1 iteration, set a non-empty `FinalOutput`,
   and preserve the locale's prompt bytes inside `FinalPrompt` after
   the refinement-marker append.
6. **DetectCycle three-payload contract.** Per locale: a trigger-keyword
   payload MUST flag with reason prefix `trigger phrase:` and
   confidence ≥ 0.7; a 4-word-repeated-phrase payload (≥3 copies) MUST
   flag with non-empty `RepeatedPhrase` and `Depth ≥ 3`; a benign
   payload MUST NOT flag (no false positives); trigger confidence MUST
   be ≥ repeated-phrase confidence (monotonicity).
7. **Sentinel-default re-validation.** A fresh `client.New` WITHOUT
   `SetRunner` MUST surface `ErrBaselineRunnerNotConfigured` from
   `SelfReflect`, `Refine`, AND `SelfImprove` — round-26 §11.4 audit
   fix re-validated for all three dispatch paths. `DetectCycle` and
   `MetaEvaluate` are heuristic-only and are NOT covered by this
   sentinel (they remain safe to call without a Runner — exercised
   in Sections 4 and 6).
8. **Paired mutation.** Running the describe gate with
   `--anti-bluff-mutate` plants a deliberate symbol-rename in a tmp
   copy of `docs/test-coverage.md`
   (`DetectCycle -> DetectCycle_MUTATED`), reruns the structural
   cross-reference check, and asserts the gate exits 99. Proves the
   ledger-to-source map actually catches drift instead of
   rubber-stamping it.

A Section that returns success without producing the corresponding
PASS line is a §11.9 violation regardless of how green the summary
line looks.

## Test bank

```bash
# Unit tests (CONST-050(A) — mocks allowed only here)
GOMAXPROCS=2 nice -n 19 go test -count=1 -race -v ./...

# Round-269 challenge runner (real client, capturing Runner, 5 locales)
go run ./challenges/runner/ -fixtures tests/fixtures/ouroborous/payloads.json

# Describe challenge — clean mode (exit 0)
bash challenges/scripts/ouroborous_describe_challenge.sh

# Paired-mutation gate (must exit 99)
bash challenges/scripts/ouroborous_describe_challenge.sh --anti-bluff-mutate

# Inherited governance challenges (CONST-033 / CONST-036)
bash challenges/scripts/no_suspend_calls_challenge.sh
bash challenges/scripts/host_no_auto_suspend_challenge.sh
bash challenges/scripts/chaos_failure_injection_challenge.sh
bash challenges/scripts/ddos_health_flood_challenge.sh
bash challenges/scripts/scaling_horizontal_challenge.sh
bash challenges/scripts/stress_sustained_load_challenge.sh
bash challenges/scripts/ui_terminal_interaction_challenge.sh
bash challenges/scripts/ux_end_to_end_flow_challenge.sh
```

The round-269 runner exits non-zero on any FAIL; the symbol-to-test
ledger lives in `docs/test-coverage.md`.

## Module path & development layout

```go
import "digital.vasic.ouroborous"
```

`go.mod` declares the module as `digital.vasic.ouroborous` and uses a
relative `replace` directive pointing at `../PliniusCommon`. The
challenge runner `challenges/runner/main.go` lives under the same
module — `go build ./challenges/runner/` from the repo root is
sufficient to produce the runner binary at `/tmp/`.

## Lineage

Extracted from internal HelixAgent research tree on 2026-04-21,
graduated to functional status the next day alongside its 7 sibling
Plinius modules. Round-26 §11.4 audit (2026-05-17) removed the
silent "RESPONSE:"-prefix baseline `Runner` after it was found to be
producing fabricated reflection / refinement / evaluation data when
downstream consumers forgot to call `SetRunner`. Round-269 (2026-05-19)
adds the deep-doc ledger, the multi-locale challenge runner, and the
paired-mutation describe gate.

Historical research corpus (unused) remains at
`docs/research/go-elder-plinius-v3/go-elder-plinius/go-ourobopus/`
inside the HelixAgent repository.

## Governance

This submodule inherits the constitution submodule's universal
rules. See `CLAUDE.md`, `AGENTS.md`, `CONSTITUTION.md` for the
cascaded clauses (CONST-033, CONST-035, CONST-036, CONST-042,
CONST-043, CONST-047..061).

## License

Apache-2.0
