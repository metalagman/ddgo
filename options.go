package ddgo

const (
	defaultMaxUserAgentLen = 2048
	defaultResultCacheSize = 256
)

type options struct {
	maxUserAgentLen int
	trimWhitespace  bool
	resultCache     ResultCache
}

// Option configures Detector behavior.
type Option func(*options)

func defaultOptions() options {
	return options{
		maxUserAgentLen: defaultMaxUserAgentLen,
		trimWhitespace:  true,
		resultCache:     NewLRUResultCache(defaultResultCacheSize),
	}
}

// WithMaxUserAgentLen limits how many bytes of a user-agent string are parsed.
// Values below 1 are ignored.
func WithMaxUserAgentLen(limit int) Option {
	return func(cfg *options) {
		if limit < 1 {
			return
		}
		cfg.maxUserAgentLen = limit
	}
}

// WithUserAgentTrimming toggles normalization of user-agent whitespace.
//
// When enabled, Parse collapses repeated whitespace and trims leading/trailing
// space before matching.
func WithUserAgentTrimming(enabled bool) Option {
	return func(cfg *options) {
		cfg.trimWhitespace = enabled
	}
}

// WithResultCache configures a custom parse result cache implementation.
//
// Passing nil explicitly disables caching.
func WithResultCache(cache ResultCache) Option {
	return func(cfg *options) {
		cfg.resultCache = cache
	}
}

func (cfg options) cache() ResultCache {
	return cfg.resultCache
}
