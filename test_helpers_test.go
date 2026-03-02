package ddgo

import "testing"

func newTestDetector(tb testing.TB, opts ...Option) *Detector {
	tb.Helper()

	detector, err := New(opts...)
	if err != nil {
		tb.Fatalf("New() failed: %v", err)
	}
	return detector
}

func mustParse(tb testing.TB, detector *Detector, userAgent string) Result {
	tb.Helper()

	result, err := detector.Parse(userAgent)
	if err != nil {
		tb.Fatalf("Parse(%q) failed: %v", userAgent, err)
	}
	return result
}

func mustParseWithHeaders(tb testing.TB, detector *Detector, userAgent string, headers map[string]string) Result {
	tb.Helper()

	result, err := detector.ParseWithHeaders(userAgent, headers)
	if err != nil {
		tb.Fatalf("ParseWithHeaders(%q) failed: %v", userAgent, err)
	}
	return result
}

func mustParseWithHints(tb testing.TB, detector *Detector, userAgent string, hints ClientHints) Result {
	tb.Helper()

	result, err := detector.ParseWithClientHints(userAgent, hints)
	if err != nil {
		tb.Fatalf("ParseWithClientHints(%q) failed: %v", userAgent, err)
	}
	return result
}
