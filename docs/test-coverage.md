# Test-Coverage Ledger — round-269

This ledger maps every exported symbol of `digital.vasic.ouroborous`
to the test or Challenge that exercises it with captured runtime
evidence. Per CONST-035, CONST-050(B), and the 2026-05-19 operator
mandate quoted below, no symbol may PASS without a corresponding
runtime-evidence exercise.

> Verbatim 2026-05-19 operator mandate: "all existing tests and
> Challenges do work in anti-bluff manner - they MUST confirm that
> all tested codebase really works as expected! We had been in
> position that all tests do execute with success and all
> Challenges as well, but in reality the most of the features does
> not work and can't be used! This MUST NOT be the case and
> execution of tests and Challenges MUST guarantee the quality, the
> completition and full usability by end users of the product!"

Operative rule (Article XI §11.9): **The bar for shipping is not
"tests pass" but "users can use the feature."** Every PASS in the
table below carries either a unit test, a paired-mutation gate, or
a challenge-runner section that produces positive runtime evidence —
no metadata-only / grep-only PASS counts.

## Module surface

`digital.vasic.ouroborous` ships two Go packages:

- **`pkg/types`** — value types: `MetaPrompt`, `SelfReflection`,
  `CycleDetection`, `IterationResult`, `RefinementConfig`,
  `RefinementResult`, `MetaEvaluation`. Validators where applicable;
  `RefinementConfig.Defaults()` installs Iterations=3,
  TargetScore=0.8 on zero.
- **`pkg/client`** — self-referential safety orchestration: `Client`,
  `New`, `NewFromConfig`, `Close`, `Config`, `SetRunner`,
  `SelfReflect`, `Refine`, `MetaEvaluate`, `SelfImprove`,
  `GetMetaPatterns`, `DetectCycle`. One function type — `Runner`.
  One sentinel — `ErrBaselineRunnerNotConfigured`.

## Symbol → exerciser map

### `pkg/types` (`types.go`)

| Symbol | Kind | Exercised by |
|--------|------|--------------|
| `MetaPrompt` | struct | runner Section 1 (2 seeded patterns retrieved + Validate exercised inline) + `pkg/types/types_test.go` |
| `MetaPrompt.Validate` | method | runner Section 1 (per seeded pattern) + `pkg/types/types_test.go` (empty Description / ID / Name rejected) |
| `SelfReflection` | struct | runner Section 2 (5 locales — OriginalPrompt + SelfAssessment byte-exact round-trip) + `pkg/client/client_test.go` |
| `SelfReflection.Validate` | method | `pkg/types/types_test.go` (Confidence out-of-range rejected) |
| `CycleDetection` | struct | runner Section 6 (5 locales × {trigger, repeated, benign}) + `pkg/client/client_test.go` |
| `IterationResult` | struct | runner Section 3 (per-iteration Prompt + Output byte-preservation across 5 locales) + `pkg/types/types_test.go` |
| `IterationResult.Validate` | method | `pkg/types/types_test.go` (empty Prompt rejected) |
| `RefinementConfig` | struct | runner Section 3+5 (real Refine + SelfImprove dispatch) + `pkg/types/types_test.go` |
| `RefinementConfig.Validate` | method | `pkg/client/client_test.go` (TestRefineInvalid — empty config rejected) + `pkg/types/types_test.go` |
| `RefinementConfig.Defaults` | method | exercised transitively when Refine is invoked with Iterations==0 (runner Section 3 supplies explicit 3 so Defaults is a no-op there; covered by unit test path) |
| `RefinementResult` | struct | runner Section 3 (3 iterations recorded, FinalScore in (0,1]) + Section 5 (SelfImprove FinalPrompt locale-bytes preserved) |
| `MetaEvaluation` | struct | runner Section 4 (5 locales × 3 criteria, Scores+Criteria+Prompt+Output byte-exact) |

### `pkg/client` (`client.go`)

| Symbol | Kind | Exercised by |
|--------|------|--------------|
| `Runner` | type alias | runner Section 2+3+5 (capturingRunner.Run wired via SetRunner) |
| `Client` | struct | runner Sections 1-7 |
| `New` | func | runner Sections 1-7 (every section constructs a fresh client) + `pkg/client/client_test.go` (TestNew) |
| `NewFromConfig` | func | `pkg/client/client_extra_test.go` (config injection path) |
| `Client.Close` | method | runner Sections 1-7 (defer Close + double-close idempotency in Section 1) + `pkg/client/client_test.go` (TestDoubleClose) |
| `Client.Config` | method | runner Section 1 (non-nil cfg asserted) + `pkg/client/client_test.go` (TestConfig) |
| `Client.SetRunner` | method | runner Section 2+3+5 (capturing Runner installed) + `pkg/client/client_test.go` (TestSetRunner — override semantics) |
| `Client.SelfReflect` | method | runner Section 2 (5 locales, capturing Runner round-trip) + `pkg/client/client_test.go` (TestSelfReflect + empty-prompt rejection + sentinel re-validation) |
| `Client.Refine` | method | runner Section 3 (5 locales, 3 iterations, per-iter byte preservation, FinalScore range) + `pkg/client/client_test.go` (TestRefine + TestRefineInvalid + sentinel re-validation) |
| `Client.MetaEvaluate` | method | runner Section 4 (5 locales × 3 criteria, byte-exact prompt+output) + `pkg/client/client_test.go` (TestMetaEvaluate) |
| `Client.SelfImprove` | method | runner Section 5 (5 locales, EarlyStop, FinalPrompt locale bytes) + `pkg/client/client_test.go` (TestSelfImprove + sentinel re-validation) |
| `Client.GetMetaPatterns` | method | runner Section 1 (2 seeded patterns retrieved, IDs verified, MetaPrompt.Validate exercised inline) + `pkg/client/client_test.go` (TestGetMetaPatterns) |
| `Client.DetectCycle` | method | runner Section 6 (5 locales × {trigger phrase, repeated phrase, benign clean, confidence monotonicity}) + `pkg/client/client_test.go` (TestDetectCycleTriggerHit + TestDetectCycleRepeatedPhrase + TestDetectCycleNoCycle) |
| `ErrBaselineRunnerNotConfigured` | var | runner Section 7 (SelfReflect + Refine + SelfImprove ALL surface sentinel without SetRunner) + `pkg/client/client_test.go` (TestSelfReflectWithoutInjectedRunner_ReturnsSentinel + TestRefineWithoutInjectedRunner_ReturnsSentinel + TestSelfImproveWithoutInjectedRunner_ReturnsSentinel) |

## Test runs (round-269 evidence captured)

### `go test -race -count=1 ./...`

```
ok  	digital.vasic.ouroborous/pkg/client	(race ~1s)
ok  	digital.vasic.ouroborous/pkg/types	(race ~1s)
```

Both packages pass with `-race` enabled — no data-race detected at
the Runner mutex, the patterns map, or the `Client.closed` flag.

### `challenges/runner/main.go -fixtures tests/fixtures/ouroborous/payloads.json`

```
=== Round-269 Ouroborous Challenge Runner ===
... 49 PASS lines across 7 sections, 5 locales ...
=== Summary: 49 PASS, 0 FAIL ===
```

Per-locale runtime evidence captured:

- **Section 1** — 6 default-surface PASS: client.New, Config non-nil,
  GetMetaPatterns returns >=2 seeded patterns, both seeded
  patterns (`self-critique`, `refine-once`) Validate, double-Close
  idempotent.
- **Section 2** — 5 SelfReflect PASS: capturing Runner asserts
  byte-exact prompt dispatch + OriginalPrompt round-trip + "OUT:"
  marker preservation in 5 locales (rune counts captured).
- **Section 3** — 5 Refine PASS: 3 iterations per locale,
  per-iteration Prompt+Output locale-byte preservation, FinalScore
  in (0,1], capturing Runner dispatch count == 3 per locale.
- **Section 4** — 5 MetaEvaluate PASS: 3 criteria scored,
  Scores map size==3, Prompt+Output byte-exact, OverallScore in [0,1].
- **Section 5** — 5 SelfImprove PASS: EarlyStop semantics honoured,
  FinalPrompt carries locale prompt bytes after refinement marker
  append.
- **Section 6** — 20 DetectCycle PASS per locale (3 payload checks +
  monotonicity assertion × 5 locales): trigger payload flags with
  `trigger phrase:` prefix and confidence >= 0.7; repeated-phrase
  payload flags with non-empty RepeatedPhrase and Depth >= 3;
  benign payload does NOT flag; trigger conf >= repeated conf.
- **Section 7** — 3 PASS: ErrBaselineRunnerNotConfigured re-validated
  for SelfReflect, Refine, SelfImprove dispatch paths.

### `bash challenges/scripts/ouroborous_describe_challenge.sh`

Clean mode exit 0; `--anti-bluff-mutate` exit 99 (paired mutation
correctly detected — ledger-vs-source drift caught when the gate
plants a `DetectCycle -> DetectCycle_MUTATED` rename in a tmp copy
of this ledger and the structural cross-reference check trips).

## Anti-bluff invariants

This round addresses every taxonomy entry in CLAUDE.md §"Bluff
taxonomy":

- **Wrapper bluff** — the describe-challenge wrapper uses PASS/FAIL
  counters with a separate `set -uo pipefail` guard, never inline
  arithmetic on a command that prints + exits non-zero.
- **Contract bluff** — every public method on `Client` and every
  exported type listed above is exercised by a runtime test or
  challenge section. The ledger surface is closed and audited
  symbol-by-symbol. DetectCycle's three advertised signals (trigger
  keyword, repeated phrase, benign-clean) are independently
  exercised with positive + negative cases — no advertised
  capability is left untested.
- **Structural bluff** — no `check_file_exists` PASS without a
  paired functional assertion. Every PASS carries either a rune
  count, an iteration count, a dispatch count, a round-trip
  equality, a confidence value, a depth value, or an `errors.Is`
  sentinel match.
- **Comment bluff** — the README's `## Anti-bluff guarantees`
  section is enforced by `ouroborous_describe_challenge.sh` Section 5.
- **Skip bluff** — no `t.Skip()` in the unit tests; the runner has
  no `if false { … }` dead branches.

## Cross-reference to constitutional anchors

| Anchor | Layer | How honoured |
|--------|-------|--------------|
| CONST-035 / Article XI §11.9 | end-user-usability | every PASS line carries runtime evidence (locale, rune count, score, confidence, sentinel match) |
| CONST-050(A) | no-fakes-beyond-unit-tests | runner uses only the public client API; the capturingRunner is the consumer's injected dependency, NOT a library-internal mock |
| CONST-050(B) | 100%-test-type coverage | unit tests + challenge runner + paired-mutation gate together cover unit + integration-style + meta-test layers; sibling Challenges (chaos / ddos / scaling / stress / ui / ux) cover the other test types |
| CONST-053 | .gitignore | `.gitignore` covers `/bin/`, `*.test`, `coverage.out`, `*.log`, `.env*`, secrets, IDE state |

The 2026-05-19 operator mandate is preserved verbatim above and in
the runner's package doc comment.
