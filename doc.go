// Package ddgo parses user-agent strings into bot, client, operating system,
// and device metadata.
//
// The parser is intentionally deterministic and returns Unknown for fields that
// cannot be derived from available inputs. For browser client-hint enrichment,
// ParseWithHeaders and ParseWithClientHints can be used alongside Parse.
package ddgo
