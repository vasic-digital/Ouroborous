# CLAUDE.md -- digital.vasic.ouroborous


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# Cycle detector (10 triggers + 4-word repeated-phrase heuristic) + Refine loop
cd Ouroborous && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -v ./pkg/client
```
Expect: PASS; `Refine` converges before MaxIterations; `DetectCycle` catches self-referential outputs.


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

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->

