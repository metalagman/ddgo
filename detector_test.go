package ddgo

import "testing"

func TestNewReturnsDetector(t *testing.T) {
	t.Parallel()

	d := New()
	if d == nil {
		t.Fatal("New() returned nil detector")
	}
}

func TestParseDefaults(t *testing.T) {
	t.Parallel()

	result := New().Parse("Mozilla/5.0")

	if result.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected user agent %q", result.UserAgent)
	}
	if result.Bot.IsBot {
		t.Fatal("expected default IsBot=false")
	}
	if result.Bot.Name != Unknown {
		t.Fatalf("unexpected bot name %q", result.Bot.Name)
	}
	if result.Client.Name != Unknown {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.OS.Name != Unknown {
		t.Fatalf("unexpected os name %q", result.OS.Name)
	}
	if result.Device.Type != Unknown {
		t.Fatalf("unexpected device type %q", result.Device.Type)
	}
}

func TestParseAppliesWhitespaceTrimmingByDefault(t *testing.T) {
	t.Parallel()

	result := New().Parse("  Mozilla/5.0  ")
	if result.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected user agent after trim %q", result.UserAgent)
	}
}

func TestWithUserAgentTrimmingFalse(t *testing.T) {
	t.Parallel()

	result := New(WithUserAgentTrimming(false)).Parse("  Mozilla/5.0  ")
	if result.UserAgent != "  Mozilla/5.0  " {
		t.Fatalf("unexpected user agent without trim %q", result.UserAgent)
	}
}

func TestWithMaxUserAgentLen(t *testing.T) {
	t.Parallel()

	result := New(WithMaxUserAgentLen(3)).Parse("Mozilla/5.0")
	if result.UserAgent != "Moz" {
		t.Fatalf("unexpected capped user agent %q", result.UserAgent)
	}
}

func TestWithMaxUserAgentLenIgnoresInvalidValue(t *testing.T) {
	t.Parallel()

	result := New(WithMaxUserAgentLen(0)).Parse("Mozilla/5.0")
	if result.UserAgent != "Mozilla/5.0" {
		t.Fatalf("expected uncapped user agent, got %q", result.UserAgent)
	}
}

func TestParseOnNilDetector(t *testing.T) {
	t.Parallel()

	var d *Detector
	result := d.Parse("Mozilla/5.0")
	if result.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected user agent %q", result.UserAgent)
	}
}
