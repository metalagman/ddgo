# ddgo

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/ddgo)](https://goreportcard.com/report/github.com/metalagman/ddgo)
[![Go Reference](https://pkg.go.dev/badge/github.com/metalagman/ddgo.svg)](https://pkg.go.dev/github.com/metalagman/ddgo)
[![lint](https://github.com/metalagman/ddgo/actions/workflows/lint.yml/badge.svg)](https://github.com/metalagman/ddgo/actions/workflows/lint.yml)
[![ci](https://github.com/metalagman/ddgo/actions/workflows/ci.yml/badge.svg)](https://github.com/metalagman/ddgo/actions/workflows/ci.yml)
[![version](https://img.shields.io/github/v/release/metalagman/ddgo?sort=semver)](https://github.com/metalagman/ddgo/releases)
[![license](https://img.shields.io/github/license/metalagman/ddgo)](LICENSE)

`ddgo` is a Go port of Matomo Device Detector with deterministic snapshot syncing and third-party provenance tracking.

Go docs: <https://pkg.go.dev/github.com/metalagman/ddgo>

## Library usage

```go
import "github.com/metalagman/ddgo"

d, err := ddgo.New()
if err != nil {
    // handle initialization error
}
r := d.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
// r.Client.Name == "Firefox"
```

For browser client hints, use:

- `ParseWithHeaders(userAgent, headers)` for `Sec-CH-UA*` HTTP headers
- `ParseWithClientHints(userAgent, hints)` for pre-parsed hints

Runnable examples are in [`example_test.go`](example_test.go).

## ddsync CLI (Cobra)

`ddsync` mirrors upstream Device Detector regex data and generates deterministic artifacts.

```bash
go run ./cmd/ddsync update
go run ./cmd/ddsync verify
go run ./cmd/ddsync status
```

Additional commands:

```bash
go run ./cmd/ddsync version
go run ./cmd/ddsync completion bash
```

Machine-readable output is available with `--json`.

For development and contribution workflow, see [`CONTRIBUTING.md`](CONTRIBUTING.md).
