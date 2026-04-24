# CLAUDE.md -- digital.vasic.ouroborous


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

<!-- TODO: replace this block with the exact command(s) that exercise this
     module end-to-end against real dependencies, and the expected output.
     The commands must run the real artifact (built binary, deployed
     container, real service) — no in-process fakes, no mocks, no
     `httptest.NewServer`, no Robolectric, no JSDOM as proof of done. -->

```bash
# TODO
```

Module-specific guidance for Claude Code.

## Status

**FUNCTIONAL.** 2 packages (types, client) ship tested implementations;
`go test -race ./...` all green. Baseline runner + 2 default
meta-patterns seeded on `New()`. `DetectCycle` ships with 10 trigger
phrases + 4-word repeated-phrase detector.

## Hard rules

1. **NO CI/CD pipelines** -- no `.github/workflows/`, `.gitlab-ci.yml`,
   `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated
   pipeline. No Git hooks either. Permanent.
2. **SSH-only for Git** -- `git@github.com:...` / `git@gitlab.com:...`.
3. **Conventional Commits** -- `feat(ouroborous): ...`, `fix(...)`,
   `docs(...)`, `test(...)`, `refactor(...)`.
4. **Code style** -- `gofmt`, `goimports`, 100-char line ceiling,
   errors always checked and wrapped (`fmt.Errorf("...: %w", err)`).
5. **Resource cap for tests** --
   `GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...`

## Purpose

Recursive/self-referential safety patterns + feedback-driven prompt
refinement. Key surface: `SelfReflect`, `Refine`, `MetaEvaluate`,
`SelfImprove`, `GetMetaPatterns`, `DetectCycle`, `SetRunner`.

## Primary consumer

HelixAgent (`dev.helix.agent`) — guardrail ingress + prompt refinement.

## Testing

```
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...
```

## API Cheat Sheet

**Module path:** `digital.vasic.ouroborous`.

```go
type Runner func(ctx, prompt string) (string, error)

type MetaPrompt struct {
    ID, Name, Template, Category, Description string
    Variables []string
    CycleDetectionEnabled bool
}
type SelfReflection struct {
    OriginalOutput, Critique string
    ImproveScore float64
}
type CycleDetection struct {
    CycleDetected bool
    TriggerPhrases, RepeatedPhrases []string
    LoopDepth int
}
type RefinementConfig struct {
    MaxIterations int
    Threshold     float64
}
type RefinementResult struct {
    Success bool
    InitialOutput, RefinedOutput string
    Iterations int
    CycleHistory []CycleDetection
}

type Client struct { /* runner + meta-patterns + cycle detector */ }

func New(opts ...config.Option) (*Client, error)
func (c *Client) SetRunner(r Runner)
func (c *Client) SelfReflect(ctx, output string) (*SelfReflection, error)
func (c *Client) Refine(ctx, prompt string, cfg RefinementConfig) (*RefinementResult, error)
func (c *Client) MetaEvaluate(ctx, output string) (*MetaEvaluation, error)
func (c *Client) SelfImprove(ctx, prompt string, maxRounds int) (*RefinementResult, error)
func (c *Client) DetectCycle(ctx, text string) (*CycleDetection, error)
func (c *Client) Close() error
```

**Typical usage:**
```go
c, _ := ouroborous.New()
defer c.Close()
c.SetRunner(llmRunner)
r, _ := c.Refine(ctx, prompt, ouroborous.RefinementConfig{MaxIterations: 3})
if len(r.CycleHistory) > 0 && r.CycleHistory[0].CycleDetected { /* handle loop */ }
```

**Injection points:** `Runner` (LLM provider).
**Defaults on `New`:** baseline runner, 2 default meta-patterns, cycle detector (10 trigger phrases + 4-word repeated-phrase heuristic).

## Integration Seams

| Direction | Sibling modules |
|-----------|-----------------|
| Upstream (this module imports) | PliniusCommon |
| Downstream (these import this module) | root only |

*Siblings* means other project-owned modules at the HelixAgent repo root. The root HelixAgent app and external systems are not listed here — the list above is intentionally scoped to module-to-module seams, because drift *between* sibling modules is where the "tests pass, product broken" class of bug most often lives. See root `CLAUDE.md` for the rules that keep these seams contract-tested.
