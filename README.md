# ddgo

`ddgo` is a Go port of Matomo Device Detector with deterministic snapshot syncing and third-party provenance tracking.

## Library usage

```go
import "github.com/metalagman/ddgo"

d := ddgo.New()
r := d.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
// r.Client.Name == "Firefox"
```

For browser client hints, use:

- `ParseWithHeaders(userAgent, headers)` for `Sec-CH-UA*` HTTP headers
- `ParseWithClientHints(userAgent, hints)` for pre-parsed hints

Runnable examples are in [`example_test.go`](example_test.go).

## ddsync CLI (Cobra)

`ddsync` manages deterministic artifacts from pinned snapshots.

```bash
go run ./cmd/ddsync update --version v1 --upstream-version <matomo-tag-or-commit>
go run ./cmd/ddsync verify --version v1 --upstream-version <matomo-tag-or-commit>
go run ./cmd/ddsync status --version v1 --upstream-version <matomo-tag-or-commit>
```

Additional commands:

```bash
go run ./cmd/ddsync version
go run ./cmd/ddsync completion bash
```

Machine-readable output is available with `--json`.

Version semantics:

- `--version`: internal snapshot/artifact version used by this repository (for example `v1`).
- `--upstream-version`: required upstream source reference from `matomo-org/device-detector` (tag or commit).
- `--upstream-repo`: upstream repository slug (defaults to `matomo-org/device-detector`).

## Development checks

Run local quality gates:

```bash
go test ./...
go test ./... -run '^Example'
go test -race ./...
go mod tidy && git diff --exit-code go.mod go.sum
go mod verify
go run ./cmd/ddsync verify --version v1 --upstream-version <matomo-tag-or-commit> --json
```

## Snapshot artifacts

Pinned source snapshot directory: `sync/snapshots/v1/`

Generated artifacts:

- `sync/compiled.json`
- `sync/manifest.json`

Golden parity fixtures:

- `testdata/parity/cases.json`
- `testdata/parity/golden.json`

Refresh golden outputs after parser changes:

```bash
UPDATE_GOLDEN=1 go test ./... -run TestParityGoldenFixtures
```

## Licensing

- Project source code is licensed under MIT (see `LICENSE`).
- Snapshot-derived third-party data is covered by upstream terms and tracked in:
  - `THIRD_PARTY_NOTICES.md`
  - `licenses/LGPL-3.0-or-later.txt`
  - `compliance/provenance.json`

When redistributing, keep project and third-party license/provenance files together.

## Maintainer runbook

1. Keep `go.mod` and `go.sum` clean (`go mod tidy`, `go mod verify`).
2. Keep CI green (`CI` and `Lint` workflows).
3. Re-run `ddsync update/verify` when snapshot sources change.
4. Keep docs/examples aligned with behavior changes.
5. Tag releases only from a clean main branch with passing tests.

Support policy:

- Best-effort support for the Go version declared in `go.mod`.
- Security and correctness fixes take priority over feature work.
