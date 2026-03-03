# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Project Goal

`ddgo` is a Go user-agent detection library. The project goal is to:

- Provide stable, deterministic parsing results for bot/client/os/device detection.
- Keep behavior aligned with upstream Device Detector fixtures and parity tests.
- Track and sync snapshot-derived data reproducibly via `ddsync`.

## Project Structure

- Library core:
  - `detector.go`, `parser*.go`, `*_rules.go`, `hints.go`, `types.go`, `options.go`, `cache.go`
- Embedded snapshot/runtime assets:
  - `sync/compiled.json`, `snapshot_data.go`, `runtime_init.go`
- Snapshot sync and artifact pipeline:
  - `internal/ddsync/`
- CLI for snapshot maintenance:
  - `cmd/ddsync/`
- Test suites and fixtures:
  - `*_test.go`, `testdata/`, `sync/current/`
- Compliance and licensing:
  - `compliance/`, `licenses/`, `THIRD_PARTY_NOTICES.md`

## Workflow

1. Pick and claim work with `bd`.
2. Keep changes scoped and deterministic (no hidden generated drift).
3. Run relevant quality gates before finishing:
   - `task test`
   - `task lint`
   - `task codex:test` (sandbox/codex environments)
   - `task codex:lint` (sandbox/codex environments)
   - `go test -race ./...`
   - `go mod tidy && git diff --exit-code go.mod go.sum`
   - `go run ./cmd/ddsync verify --json` (when snapshot/artifact code is touched)
4. Update docs/examples when behavior changes.
5. Complete the session with the mandatory push flow below.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
