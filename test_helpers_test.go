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
