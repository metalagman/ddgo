# Performance Envelope

This document tracks the current parser throughput and concurrency safety envelope for `ddgo`.

## Safety checks

Run race detection for all packages:

```bash
GOCACHE=.cache/go-build go test -race ./...
```

Current status: passing (last verified on 2026-03-01).

## Benchmarks

Run parser benchmarks with memory stats:

```bash
GOCACHE=.cache/go-build go test -run '^$' -bench 'BenchmarkParse' -benchmem .
```

Latest sample output (Linux/amd64, AMD Ryzen 7 PRO 7840U):

- `BenchmarkParseFirefox`: `13548 ns/op`, `1396 B/op`, `16 allocs/op`
- `BenchmarkParseGooglebot`: `5624 ns/op`, `1221 B/op`, `13 allocs/op`
- `BenchmarkParseCachedFirefox`: `1012 ns/op`, `1296 B/op`, `13 allocs/op`
- `BenchmarkParseWithHeaders`: `6370 ns/op`, `1723 B/op`, `29 allocs/op`

## Notes

- Regex patterns are precompiled once at package init time.
- `Detector` now includes an internal bounded LRU cache (default size `256`) for `Parse` calls.
- Cache size can be configured with `WithResultCacheSize(size)`; use `0` to disable caching.
