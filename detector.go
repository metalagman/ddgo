package ddgo

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNilDetector is returned when Parse is called on a nil *Detector.
var ErrNilDetector = errors.New("ddgo: nil detector")

// Detector parses user-agent strings into structured detection results.
//
// A Detector is safe for concurrent use by multiple goroutines.
type Detector struct {
	opts    options
	cache   ResultCache
	runtime *parserRuntime
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
	runtime, err := newParserRuntime()
	if err != nil {
		return nil, err
	}
	return &Detector{
		opts:    cfg,
		cache:   cfg.cache(),
		runtime: runtime,
	}, nil
}

// Parse analyzes a user-agent string and returns a detection result.
//
// Parse can return cached results for identical normalized user-agent inputs.
func (d *Detector) Parse(userAgent string) (Result, error) {
	return d.parseWithCache(userAgent)
}

// ParseWithClientHints analyzes a user-agent string with explicit client hints.
//
// Hint-based parsing bypasses the internal Parse cache.
func (d *Detector) ParseWithClientHints(userAgent string, hints ClientHints) (Result, error) {
	return d.parseWithClientHints(userAgent, hints)
}

// ParseWithHeaders analyzes a user-agent string and Sec-CH-UA style headers.
//
// This helper normalizes headers via ParseClientHintsFromHeaders and then
// delegates to ParseWithClientHints.
func (d *Detector) ParseWithHeaders(userAgent string, headers map[string]string) (Result, error) {
	return d.parseWithClientHints(userAgent, ParseClientHintsFromHeaders(headers))
}

func (d *Detector) parseWithCache(userAgent string) (Result, error) {
	if d == nil {
		return Result{}, ErrNilDetector
	}

	ua := d.prepareUserAgent(userAgent)
	if d.cache != nil {
		cached, ok := d.cache.Get(ua)
		if ok {
			return cached, nil
		}
	}

	result, err := d.parseUserAgent(ua, ClientHints{})
	if err != nil {
		return Result{}, err
	}
	if d.cache != nil {
		d.cache.Set(ua, result)
	}
	return result, nil
}

func (d *Detector) parseWithClientHints(userAgent string, hints ClientHints) (Result, error) {
	if d == nil {
		return Result{}, ErrNilDetector
	}

	ua := d.prepareUserAgent(userAgent)
	return d.parseUserAgent(ua, hints)
}

func (d *Detector) prepareUserAgent(userAgent string) string {
	ua := normalizeUserAgent(userAgent)
	if d.opts.trimWhitespace {
		ua = trimUserAgentWhitespace(ua)
	}
	if d.opts.maxUserAgentLen > 0 && len(ua) > d.opts.maxUserAgentLen {
		ua = ua[:d.opts.maxUserAgentLen]
	}
	return ua
}

func (d *Detector) parseUserAgent(ua string, hints ClientHints) (Result, error) {
	bot, err := parseBot(d.runtime, ua)
	if err != nil {
		return Result{}, fmt.Errorf("parse bot: %w", err)
	}
	client := unknownClient()
	if !bot.IsBot {
		client, err = parseClient(d.runtime, ua)
		if err != nil {
			return Result{}, fmt.Errorf("parse client: %w", err)
		}
	}
	osInfo, err := parseOS(d.runtime, ua)
	if err != nil {
		return Result{}, fmt.Errorf("parse os: %w", err)
	}
	device := deviceForBot(bot)
	if !bot.IsBot {
		device, err = parseDevice(d.runtime, ua)
		if err != nil {
			return Result{}, fmt.Errorf("parse device: %w", err)
		}
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
	return result, nil
}

func normalizeUserAgent(userAgent string) string {
	return strings.NewReplacer(
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
		"\t", " ",
	).Replace(userAgent)
}

func trimUserAgentWhitespace(userAgent string) string {
	return strings.Join(strings.Fields(userAgent), " ")
}

func deviceForBot(bot Bot) Device {
	if bot.IsBot {
		return Device{
			Type:  "Bot",
			Brand: Unknown,
			Model: Unknown,
		}
	}
	return Device{
		Type:  Unknown,
		Brand: Unknown,
		Model: Unknown,
	}
}
