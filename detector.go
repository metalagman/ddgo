package ddgo

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNilDetector = errors.New("ddgo: nil detector")

// Detector parses user-agent strings into structured detection results.
//
// A Detector is safe for concurrent use by multiple goroutines.
type Detector struct {
	opts  options
	cache ResultCache
}

// New creates a detector with optional configuration overrides.
func New(opts ...Option) (*Detector, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&cfg)
	}
	if err := initParserRuntime(); err != nil {
		return nil, err
	}
	return &Detector{
		opts:  cfg,
		cache: cfg.cache(),
	}, nil
}

// Parse analyzes a user-agent string and returns a detection result.
//
// Parse can return cached results for identical normalized user-agent inputs.
func (d *Detector) Parse(userAgent string) (Result, error) {
	return d.parse(userAgent, ClientHints{}, true)
}

// ParseWithClientHints analyzes a user-agent string with explicit client hints.
//
// Hint-based parsing bypasses the internal Parse cache.
func (d *Detector) ParseWithClientHints(userAgent string, hints ClientHints) (Result, error) {
	return d.parse(userAgent, hints, false)
}

// ParseWithHeaders analyzes a user-agent string and Sec-CH-UA style headers.
//
// This helper normalizes headers via ParseClientHintsFromHeaders and then
// delegates to ParseWithClientHints.
func (d *Detector) ParseWithHeaders(userAgent string, headers map[string]string) (Result, error) {
	return d.parse(userAgent, ParseClientHintsFromHeaders(headers), false)
}

func (d *Detector) parse(userAgent string, hints ClientHints, allowCache bool) (Result, error) {
	if d == nil {
		return Result{}, ErrNilDetector
	}

	ua := normalizeUserAgent(userAgent, d.opts.trimWhitespace)
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
	}
	if allowCache && d.cache != nil {
		if cached, ok := d.cache.Get(ua); ok {
			return cached, nil
		}
	}

	bot, err := parseBot(ua)
	if err != nil {
		return Result{}, fmt.Errorf("parse bot: %w", err)
	}
	client, err := parseClient(ua, bot.IsBot)
	if err != nil {
		return Result{}, fmt.Errorf("parse client: %w", err)
	}
	osInfo, err := parseOS(ua)
	if err != nil {
		return Result{}, fmt.Errorf("parse os: %w", err)
	}
	device, err := parseDevice(ua, bot.IsBot)
	if err != nil {
		return Result{}, fmt.Errorf("parse device: %w", err)
	}

	result := Result{
		UserAgent: ua,
		Bot:       bot,
		Client:    client,
		OS:        osInfo,
		Device:    device,
	}
	if !bot.IsBot {
		applyClientHints(&result, hints)
	}
	if allowCache && d.cache != nil {
		d.cache.Set(ua, result)
	}
	return result, nil
}

func normalizeUserAgent(userAgent string, trimWhitespace bool) string {
	normalized := strings.NewReplacer(
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
		"\t", " ",
	).Replace(userAgent)
	if trimWhitespace {
		return strings.Join(strings.Fields(normalized), " ")
	}
	return normalized
}
