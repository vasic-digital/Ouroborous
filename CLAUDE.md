# CLAUDE.md -- digital.vasic.ouroborous

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
