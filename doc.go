// Package ddgo parses user-agent strings into bot, client, operating system,
// and device metadata.
//
// The parser is intentionally deterministic: enum-like fields fall back to
// typed Unknown values while open text fields remain empty when unavailable.
// For browser client-hint enrichment, ParseWithHeaders and ParseWithClientHints
// can be used alongside Parse.
package ddgo
