package ddgo

import "strings"

// Detector parses user-agent strings into structured detection results.
//
// A Detector is safe for concurrent use by multiple goroutines.
type Detector struct {
	opts  options
	cache *resultCache
}

// New creates a detector with optional configuration overrides.
func New(opts ...Option) *Detector {
	cfg := defaultOptions()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&cfg)
	}
	return &Detector{
		opts:  cfg,
		cache: newResultCache(cfg.resultCacheSize),
	}
}

// Parse analyzes a user-agent string and returns a detection result.
//
// Parse can return cached results for identical normalized user-agent inputs.
// If d is nil, Parse behaves as if called on New().
func (d *Detector) Parse(userAgent string) Result {
	return d.parse(userAgent, ClientHints{}, true)
}

// ParseWithClientHints analyzes a user-agent string with explicit client hints.
//
// Hint-based parsing bypasses the internal Parse cache. If d is nil,
// ParseWithClientHints behaves as if called on New().
func (d *Detector) ParseWithClientHints(userAgent string, hints ClientHints) Result {
	return d.parse(userAgent, hints, false)
}

// ParseWithHeaders analyzes a user-agent string and Sec-CH-UA style headers.
//
// This helper normalizes headers via ParseClientHintsFromHeaders and then
// delegates to ParseWithClientHints. If d is nil, ParseWithHeaders behaves as
// if called on New().
func (d *Detector) ParseWithHeaders(userAgent string, headers map[string]string) Result {
	return d.parse(userAgent, ParseClientHintsFromHeaders(headers), false)
}

func (d *Detector) parse(userAgent string, hints ClientHints, allowCache bool) Result {
	if d == nil {
		d = New()
	}

	ua := normalizeUserAgent(userAgent, d.opts.trimWhitespace)
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
	}
	if allowCache && d.cache != nil {
		if cached, ok := d.cache.get(ua); ok {
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
		d.cache.set(ua, result)
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
