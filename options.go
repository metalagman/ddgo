package ddgo

const defaultMaxUserAgentLen = 2048

type options struct {
	maxUserAgentLen int
	trimWhitespace  bool
}

// Option configures detector behavior.
type Option func(*options)

func defaultOptions() options {
	return options{
		maxUserAgentLen: defaultMaxUserAgentLen,
		trimWhitespace:  true,
	}
}

// WithMaxUserAgentLen limits how many bytes of a user-agent string are parsed.
// Values below 1 are ignored.
func WithMaxUserAgentLen(max int) Option {
	return func(cfg *options) {
		if max < 1 {
			return
		}
		cfg.maxUserAgentLen = max
	}
}

// WithUserAgentTrimming toggles trimming of leading/trailing whitespace.
func WithUserAgentTrimming(enabled bool) Option {
	return func(cfg *options) {
		cfg.trimWhitespace = enabled
	}
}
