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

	return Result{
		UserAgent: ua,
		Bot: Bot{
			Name:     Unknown,
			Category: Unknown,
			URL:      Unknown,
			Producer: Producer{
				Name: Unknown,
				URL:  Unknown,
			},
		},
		Client: Client{
			Type:          Unknown,
			Name:          Unknown,
			Version:       Unknown,
			Engine:        Unknown,
			EngineVersion: Unknown,
		},
		OS: OS{
			Name:     Unknown,
			Version:  Unknown,
			Platform: Unknown,
		},
		Device: Device{
			Type:  Unknown,
			Brand: Unknown,
			Model: Unknown,
		},
	}
}
