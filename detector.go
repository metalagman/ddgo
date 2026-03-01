package ddgo

import "strings"

// Detector is the entry point for user-agent parsing.
type Detector struct {
	opts options
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
		opts: cfg,
	}
}

// Parse analyzes a user-agent string and returns a detection result.
func (d *Detector) Parse(userAgent string) Result {
	return d.parse(userAgent, ClientHints{})
}

// ParseWithClientHints analyzes a user-agent string with optional client hints.
func (d *Detector) ParseWithClientHints(userAgent string, hints ClientHints) Result {
	return d.parse(userAgent, hints)
}

// ParseWithHeaders analyzes a user-agent string and Sec-CH-UA style headers.
func (d *Detector) ParseWithHeaders(userAgent string, headers map[string]string) Result {
	return d.parse(userAgent, ParseClientHintsFromHeaders(headers))
}

func (d *Detector) parse(userAgent string, hints ClientHints) Result {
	if d == nil {
		d = New()
	}

	ua := normalizeUserAgent(userAgent, d.opts.trimWhitespace)
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
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
