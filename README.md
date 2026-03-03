# ddgo

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/ddgo)](https://goreportcard.com/report/github.com/metalagman/ddgo)
[![Go Reference](https://pkg.go.dev/badge/github.com/metalagman/ddgo.svg)](https://pkg.go.dev/github.com/metalagman/ddgo)
[![lint](https://github.com/metalagman/ddgo/actions/workflows/lint.yml/badge.svg)](https://github.com/metalagman/ddgo/actions/workflows/lint.yml)
[![ci](https://github.com/metalagman/ddgo/actions/workflows/ci.yml/badge.svg)](https://github.com/metalagman/ddgo/actions/workflows/ci.yml)
[![version](https://img.shields.io/github/v/release/metalagman/ddgo?sort=semver)](https://github.com/metalagman/ddgo/releases)
[![license](https://img.shields.io/github/license/metalagman/ddgo)](LICENSE)

`ddgo` is a Go port of [Matomo Device Detector](https://github.com/matomo-org/device-detector).

## Library usage

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
```

Bot detection:

```go
result, _ := detector.Parse("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
// result.Bot.IsBot == true
// result.Bot.Name == "Googlebot"
// result.Device.Type == "Bot"
```

Client hints via structured input:

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

Client hints from headers:

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

Parser options:

```go
detector, _ = ddgo.New(ddgo.WithMaxUserAgentLen(7))
result, _ = detector.Parse("Mozilla/5.0")
// result.UserAgent == "Mozilla"

detector, _ = ddgo.New(ddgo.WithUserAgentTrimming(false))
result, _ = detector.Parse("  Mozilla/5.0  ")
// result.UserAgent == "  Mozilla/5.0  "
```

Parse only client hints:

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

Runnable examples are in [`example_test.go`](example_test.go) (`Example*` functions).
