package ddgo

const (
	defaultMaxUserAgentLen = 2048
	defaultResultCacheSize = 256
)

type options struct {
	maxUserAgentLen int
	trimWhitespace  bool
	resultCacheSize int
	resultCache     ResultCache
	cacheSet        bool
}

// Option configures Detector behavior.
type Option func(*options)

func defaultOptions() options {
	return options{
		maxUserAgentLen: defaultMaxUserAgentLen,
		trimWhitespace:  true,
		resultCacheSize: defaultResultCacheSize,
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

// WithResultCacheSize configures a bounded in-memory parse result cache.
// Set to 0 to disable caching. Negative values are ignored.
func WithResultCacheSize(size int) Option {
	return func(cfg *options) {
		if size < 0 {
			return
		}
		cfg.resultCacheSize = size
	}
}

// WithResultCache configures a custom parse result cache implementation.
//
// Passing nil explicitly disables caching.
func WithResultCache(cache ResultCache) Option {
	return func(cfg *options) {
		cfg.resultCache = cache
		cfg.cacheSet = true
	}
}

func (cfg options) cache() ResultCache {
	if cfg.cacheSet {
		return cfg.resultCache
	}
	return newResultCache(cfg.resultCacheSize)
}
