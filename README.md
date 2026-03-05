# ddgo - Device Detector for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/ddgo)](https://goreportcard.com/report/github.com/metalagman/ddgo)
[![codecov](https://codecov.io/github/metalagman/ddgo/graph/badge.svg)](https://codecov.io/github/metalagman/ddgo)
[![ci](https://github.com/metalagman/ddgo/actions/workflows/ci.yml/badge.svg)](https://github.com/metalagman/ddgo/actions/workflows/ci.yml)
[![version](https://img.shields.io/github/v/release/metalagman/ddgo?sort=semver)](https://github.com/metalagman/ddgo/releases)
[![license](https://img.shields.io/github/license/metalagman/ddgo)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/metalagman/ddgo.svg)](https://pkg.go.dev/github.com/metalagman/ddgo)

`ddgo` is a Go port of [Matomo Device Detector](https://github.com/matomo-org/device-detector).

It parses user-agent strings (and optional Client Hints) into normalized bot/client/OS/device metadata.

## What it detects

- **Bot**: bot flag, bot name, category, producer metadata.
- **Client**: client type, name, version, engine, engine version.
- **OS**: operating system name, version, and platform.
- **Device**: device type, brand, and model.

## Why ddgo

- Uses upstream Matomo regex snapshot data.
- Produces deterministic compiled artifacts (`sync/compiled.json`, `sync/manifest.json`).
- Supports Client Hints enrichment (`ParseWithClientHints`, `ParseWithHeaders`).
- Concurrency-safe detector usage.
- Optional pluggable parse-result cache.

## Install

```bash
go get github.com/metalagman/ddgo
```

## Library usage

The sections below map to runnable examples in [`example_test.go`](example_test.go).

### `New` + `Parse` (`ExampleNew`)

```go
import "github.com/metalagman/ddgo"

detector, err := ddgo.New()
if err != nil {
    // handle initialization error
}
result, err := detector.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
if err != nil {
    // handle parse error
}
// result.Client.Name == "Firefox"
// result.Client.Version == "124.0"
// result.Bot.IsBot == false
// result.OS.Name == "Windows"
// result.Device.Type == "Desktop"
```

### `Detector.Parse` bot detection (`ExampleDetector_Parse`)

```go
result, _ := detector.Parse("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
// result.Bot.IsBot == true
// result.Bot.Name == "Googlebot"
// result.Device.Type == "Bot"
```

### `Detector.ParseWithClientHints` (`ExampleDetector_ParseWithClientHints`)

```go
mobile := true
hints := ddgo.ClientHints{
    Brands:          []ddgo.ClientHintBrand{{Name: "Google Chrome", Version: "122.0.6261.128"}},
    Platform:        "Android",
    PlatformVersion: "14.0.0",
    Model:           "SM-G991B",
    Mobile:          &mobile,
}
result, _ := detector.ParseWithClientHints("Mozilla/5.0", hints)
// result.Client.Name == "Chrome"
// result.OS.Name == "Android"
// result.Device.Model == "SM-G991B"
```

### `Detector.ParseWithHeaders` (`ExampleDetector_ParseWithHeaders`)

```go
headers := map[string]string{
    "Sec-CH-UA":                  "\"Not(A:Brand\";v=\"99\", \"Microsoft Edge\";v=\"123.0.0.0\", \"Chromium\";v=\"123.0.0.0\"",
    "Sec-CH-UA-Platform":         "\"Windows\"",
    "Sec-CH-UA-Platform-Version": "\"15.0.0\"",
    "Sec-CH-UA-Mobile":           "?0",
}
result, _ := detector.ParseWithHeaders("Mozilla/5.0", headers)
// result.Client.Name == "Microsoft Edge"
// result.Client.Version == "123.0.0.0"
// result.Device.Type == "Desktop"
```

### `WithMaxUserAgentLen` (`ExampleWithMaxUserAgentLen`)

```go
detector, _ = ddgo.New(ddgo.WithMaxUserAgentLen(7))
result, _ = detector.Parse("Mozilla/5.0")
// result.UserAgent == "Mozilla"
```

### `WithUserAgentTrimming` (`ExampleWithUserAgentTrimming`)

```go
detector, _ = ddgo.New(ddgo.WithUserAgentTrimming(false))
result, _ = detector.Parse("  Mozilla/5.0  ")
// result.UserAgent == "  Mozilla/5.0  "
```

### `ParseClientHintsFromHeaders` (`ExampleParseClientHintsFromHeaders`)

```go
headers := map[string]string{
    "Sec-CH-UA-Full-Version-List": "\"Not A;Brand\";v=\"24\", \"Chromium\";v=\"122.0.6261.128\", \"Google Chrome\";v=\"122.0.6261.128\"",
    "Sec-CH-UA-Platform":          "\"Android\"",
    "Sec-CH-UA-Mobile":            "?1",
}
hints := ddgo.ParseClientHintsFromHeaders(headers)
// len(hints.Brands) == 3
// hints.Platform == "Android"
```

### Cache configuration

```go
// Preferred: choose implementation explicitly via the cache interface.
detector, _ = ddgo.New(ddgo.WithResultCache(ddgo.NewLRUResultCache(512)))
```

Independent caching interface:

```go
type ResultCache interface {
    Get(key string) (ddgo.Result, bool)
    Set(key string, result ddgo.Result)
}
```

Built-in cache implementations:

```go
// Bounded LRU-style cache:
detector, _ := ddgo.New(ddgo.WithResultCache(ddgo.NewLRUResultCache(512)))

// Unbounded in-memory cache:
detector, _ = ddgo.New(ddgo.WithResultCache(ddgo.NewMemoryResultCache()))
```

Custom cache implementation:

```go
type myCache struct{}

func (m *myCache) Get(key string) (ddgo.Result, bool) { return ddgo.Result{}, false }
func (m *myCache) Set(key string, result ddgo.Result) {}

detector, _ := ddgo.New(ddgo.WithResultCache(&myCache{}))
```

## Benchmark snapshot

Measured on Linux/amd64 (AMD Ryzen 7 PRO 7840U), March 3, 2026.

| Scenario | Throughput/Latency | Memory | Allocations |
| --- | --- | --- | --- |
| `Parse` typical desktop browser UA (`BenchmarkParseFirefox`) | ~8.2-8.6 ms/op | ~14.6-15.2 KB/op | ~75-77 allocs/op |
| `Parse` common bot UA (`BenchmarkParseGooglebot`) | ~1.34-1.37 ms/op | ~1.63-1.65 KB/op | 14 allocs/op |
| `Parse` with warm cache hit (`BenchmarkParseCachedFirefox`) | ~1.10-1.15 us/op | 1296 B/op | 13 allocs/op |
| `ParseWithHeaders` (`BenchmarkParseWithHeaders`) | ~1.44-1.67 ms/op | ~3.6-3.7 KB/op | 35 allocs/op |

Run benchmarks locally:

```bash
go test -run '^$' -bench 'BenchmarkParse' -benchmem .
```

Full performance notes are in [`PERFORMANCE.md`](PERFORMANCE.md).

## Examples

Runnable examples are in [`example_test.go`](example_test.go) (`Example*` functions).

## Data source and sync model

- Upstream source: `matomo-org/device-detector` regex definitions.
- Snapshot mirror path: `sync/current/`.
- Compiled runtime artifact: `sync/compiled.json`.
- Manifest metadata is maintained for reproducibility.

## Licensing

- `ddgo` is licensed under LGPL-3.0-or-later (same as Matomo Device Detector).
- License and notice references:
  - `LICENSE`
  - `THIRD_PARTY_NOTICES.md`
