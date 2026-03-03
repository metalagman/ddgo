# Contributing to ddgo

This document covers development workflow, local quality gates, and artifact maintenance for contributors.

## Development checks

Run local quality gates:

```bash
go test ./...
go test -tags fulltests ./...
go test -race ./...
go mod tidy && git diff --exit-code go.mod go.sum
go mod verify
go run ./cmd/ddsync verify --json
```

## Snapshot workflow

`ddsync` resolves the latest stable semver tag from `--upstream-repo` (defaults to `matomo-org/device-detector`) using git tags.

`ddsync update` clones that upstream tag and mirrors the full `regexes/` directory into `sync/current/`.

If upstream tags cannot be resolved, `ddsync` fails fast.

Current mirrored source snapshot directory:

- `sync/current/`

Generated artifacts:

- `sync/compiled.json`
- `sync/manifest.json`

## Parity fixtures

Golden parity fixtures:

- `testdata/parity/cases.json`
- `testdata/parity/golden.json`

Refresh golden outputs after parser changes:

```bash
UPDATE_GOLDEN=1 go test ./... -run TestParityGoldenFixtures
```

## Licensing and notices

- Project source code is licensed under LGPL-3.0-or-later (see `LICENSE`).
- Snapshot-derived third-party data notices are tracked in:
  - `THIRD_PARTY_NOTICES.md`

When redistributing, keep project and third-party license/notice files together.

## Maintainer runbook

1. Keep `go.mod` and `go.sum` clean (`go mod tidy`, `go mod verify`).
2. Keep CI green (`CI` and `Lint` workflows).
3. Re-run `ddsync update/verify` when snapshot sources change.
4. Keep docs/examples aligned with behavior changes.
5. Tag releases only from a clean main branch with passing tests.

Support policy:

- Best-effort support for the Go version declared in `go.mod`.
- Security and correctness fixes take priority over feature work.
