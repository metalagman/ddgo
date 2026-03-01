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

func TestParseGooglebot(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	result := New().Parse(ua)

	if !result.Bot.IsBot {
		t.Fatal("expected bot detection")
	}
	if result.Bot.Name != "Googlebot" {
		t.Fatalf("unexpected bot name %q", result.Bot.Name)
	}
	if result.Client.Name != Unknown {
		t.Fatalf("expected unknown client for bot, got %q", result.Client.Name)
	}
	if result.Device.Type != "Bot" {
		t.Fatalf("expected bot device type, got %q", result.Device.Type)
	}
}

func TestParseFirefoxWindowsDesktop(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0"
	result := New().Parse(ua)

	if result.Bot.IsBot {
		t.Fatal("did not expect bot")
	}
	if result.Client.Name != "Firefox" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.Client.Version != "124.0" {
		t.Fatalf("unexpected client version %q", result.Client.Version)
	}
	if result.OS.Name != "Windows" || result.OS.Version != "10" {
		t.Fatalf("unexpected os %+v", result.OS)
	}
	if result.Device.Type != "Desktop" {
		t.Fatalf("unexpected device type %q", result.Device.Type)
	}
}

func TestParseAndroidChromeSamsung(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (Linux; Android 14; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36"
	result := New().Parse(ua)

	if result.Client.Name != "Chrome" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.OS.Name != "Android" || result.OS.Version != "14" {
		t.Fatalf("unexpected os %+v", result.OS)
	}
	if result.Device.Type != "Smartphone" || result.Device.Brand != "Samsung" || result.Device.Model != "SM-G991B" {
		t.Fatalf("unexpected device %+v", result.Device)
	}
}

func TestParseIPhoneSafari(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Mobile/15E148 Safari/604.1"
	result := New().Parse(ua)

	if result.Client.Name != "Safari" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.OS.Name != "iOS" || result.OS.Version != "17.3.1" {
		t.Fatalf("unexpected os %+v", result.OS)
	}
	if result.Device.Type != "Smartphone" || result.Device.Brand != "Apple" || result.Device.Model != "iPhone" {
		t.Fatalf("unexpected device %+v", result.Device)
	}
}

func TestParseCurl(t *testing.T) {
	t.Parallel()

	result := New().Parse("curl/8.7.1")
	if result.Client.Type != "Library" || result.Client.Name != "curl" || result.Client.Version != "8.7.1" {
		t.Fatalf("unexpected curl client %+v", result.Client)
	}
}
