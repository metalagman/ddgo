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
	if d == nil {
		d = New()
	}

	ua := userAgent
	if d.opts.trimWhitespace {
		ua = strings.TrimSpace(ua)
	}
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
	}

	bot := parseBot(ua)

	return Result{
		UserAgent: ua,
		Bot:       bot,
		Client:    parseClient(ua, bot.IsBot),
		OS:        parseOS(ua),
		Device:    parseDevice(ua, bot.IsBot),
	}
}
