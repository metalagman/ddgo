package ddgo

import "strings"

func normalizeRulePattern(pattern string) string {
	// Upstream rules occasionally escape underscores (for PCRE), which Go engines
	// treat as an invalid escape sequence. Underscore does not require escaping.
	pattern = strings.ReplaceAll(pattern, `\\_`, `_`)
	return strings.ReplaceAll(pattern, `\_`, `_`)
}
