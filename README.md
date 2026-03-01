# ddgo

`ddgo` is a Go port of Matomo Device Detector with pinned regex snapshots, deterministic sync artifacts, and compliance metadata.

## Development checks

Run the full local quality gate:

```bash
go test ./...
go test -race ./...
go run ./cmd/ddsync verify --version v1
```

## Snapshot sync flow

Use `ddsync` to regenerate deterministic artifacts when snapshot inputs change:

```bash
go run ./cmd/ddsync update --version v1
go run ./cmd/ddsync verify --version v1
go run ./cmd/ddsync status --version v1
```

Current snapshot source directory: `sync/snapshots/v1/`.
Generated artifacts:

- `sync/compiled.json`
- `sync/manifest.json`

## Golden parity fixtures

Parity fixtures live in `testdata/parity/`:

- `cases.json`: fixture inputs and source metadata
- `golden.json`: expected parser output snapshot

To intentionally refresh golden outputs after parser changes:

```bash
UPDATE_GOLDEN=1 go test ./... -run TestParityGoldenFixtures
```

## Release checklist

1. Run the development checks.
2. Verify compliance files exist and are current:
   - `THIRD_PARTY_NOTICES.md`
   - `licenses/LGPL-3.0-or-later.txt`
   - `compliance/provenance.json`
3. If snapshots changed, run `ddsync update`, commit regenerated artifacts, and re-run `ddsync verify`.
4. Confirm parity fixtures and benchmarks still pass:
   - `go test ./...`
   - `go test -run '^$' -bench 'BenchmarkParse' -benchmem .`
5. Update docs/changelog as needed and tag the release.
