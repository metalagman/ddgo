package ddgo

import (
	"log"
	"strings"
)

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

// MustNew creates a detector and exits the process if initialization fails.
func MustNew(opts ...Option) *Detector {
	detector, err := New(opts...)
	if err != nil {
		log.Fatalf("ddgo: detector initialization failed: %v", err)
	}
	return detector
}

// Parse analyzes a user-agent string and returns a detection result.
//
// Parse can return cached results for identical normalized user-agent inputs.
// Parse panics if called on a nil Detector.
func (d *Detector) Parse(userAgent string) Result {
	return d.parse(userAgent, ClientHints{}, true)
}

// ParseWithClientHints analyzes a user-agent string with explicit client hints.
//
// Hint-based parsing bypasses the internal Parse cache. ParseWithClientHints
// panics if called on a nil Detector.
func (d *Detector) ParseWithClientHints(userAgent string, hints ClientHints) Result {
	return d.parse(userAgent, hints, false)
}

// ParseWithHeaders analyzes a user-agent string and Sec-CH-UA style headers.
//
// This helper normalizes headers via ParseClientHintsFromHeaders and then
// delegates to ParseWithClientHints. ParseWithHeaders panics if called on a nil
// Detector.
func (d *Detector) ParseWithHeaders(userAgent string, headers map[string]string) Result {
	return d.parse(userAgent, ParseClientHintsFromHeaders(headers), false)
}

func (d *Detector) parse(userAgent string, hints ClientHints, allowCache bool) Result {
	if d == nil {
		d = MustNew()
	}

	ua := normalizeUserAgent(userAgent, d.opts.trimWhitespace)
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
	}
	if allowCache && d.cache != nil {
		if cached, ok := d.cache.Get(ua); ok {
			return cached
		}
	}

	bot := parseBot(ua)
	result := Result{
		UserAgent: ua,
		Bot:       bot,
		Client:    parseClient(ua, bot.IsBot),
		OS:        parseOS(ua),
		Device:    parseDevice(ua, bot.IsBot),
	}
	if !bot.IsBot {
		applyClientHints(&result, hints)
	}
	if allowCache && d.cache != nil {
		d.cache.Set(ua, result)
	}
	return result
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
