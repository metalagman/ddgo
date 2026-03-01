package ddgo

import (
	"sync"
	"testing"
)

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

func TestWithResultCacheSizeZeroDisablesCache(t *testing.T) {
	t.Parallel()

	detector := New(WithResultCacheSize(0))
	if detector.cache != nil {
		t.Fatal("expected cache to be disabled")
	}

	result := detector.Parse("Mozilla/5.0")
	if result.UserAgent != "Mozilla/5.0" {
		t.Fatalf("unexpected parse result %+v", result)
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

func TestParseEdgeAndroidAlias(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36 EdgA/122.0.2365.66"
	result := New().Parse(ua)

	if result.Client.Name != "Microsoft Edge" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.Client.Version != "122.0.2365.66" {
		t.Fatalf("unexpected client version %q", result.Client.Version)
	}
	if result.Client.Engine != "Blink" {
		t.Fatalf("unexpected engine %q", result.Client.Engine)
	}
}

func TestParseChromeIOSAlias(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/122.0.6261.69 Mobile/15E148 Safari/604.1"
	result := New().Parse(ua)

	if result.Client.Name != "Chrome" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
	if result.Client.Version != "122.0.6261.69" {
		t.Fatalf("unexpected client version %q", result.Client.Version)
	}
	if result.Client.Engine != "WebKit" {
		t.Fatalf("unexpected engine %q", result.Client.Engine)
	}
}

func TestParseWithClientHints(t *testing.T) {
	t.Parallel()

	hints := ClientHints{
		Brands: []ClientHintBrand{
			{Name: "Not A;Brand", Version: "24"},
			{Name: "Chromium", Version: "122.0.6261.128"},
			{Name: "Google Chrome", Version: "122.0.6261.128"},
		},
		Platform:        "Android",
		PlatformVersion: "14.0.0",
		Model:           "SM-G991B",
		Mobile:          boolPtr(true),
	}

	result := New().ParseWithClientHints("Mozilla/5.0", hints)
	if result.Client.Name != "Chrome" || result.Client.Version != "122.0.6261.128" {
		t.Fatalf("unexpected client %+v", result.Client)
	}
	if result.OS.Name != "Android" || result.OS.Version != "14" || result.OS.Platform != "ARM" {
		t.Fatalf("unexpected os %+v", result.OS)
	}
	if result.Device.Type != "Smartphone" || result.Device.Brand != "Samsung" || result.Device.Model != "SM-G991B" {
		t.Fatalf("unexpected device %+v", result.Device)
	}
}

func TestParseWithHeaders(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"Sec-CH-UA":                  "\"Not(A:Brand\";v=\"99\", \"Microsoft Edge\";v=\"123.0.0.0\", \"Chromium\";v=\"123.0.0.0\"",
		"Sec-CH-UA-Platform":         "\"Windows\"",
		"Sec-CH-UA-Platform-Version": "\"15.0.0\"",
		"Sec-CH-UA-Mobile":           "?0",
	}
	result := New().ParseWithHeaders("Mozilla/5.0", headers)
	if result.Client.Name != "Microsoft Edge" || result.Client.Version != "123.0.0.0" {
		t.Fatalf("unexpected client %+v", result.Client)
	}
	if result.OS.Name != "Windows" || result.OS.Version != "15" || result.OS.Platform != "x64" {
		t.Fatalf("unexpected os %+v", result.OS)
	}
	if result.Device.Type != "Desktop" {
		t.Fatalf("unexpected device type %q", result.Device.Type)
	}
}

func TestParseWithClientHintsBotPrecedence(t *testing.T) {
	t.Parallel()

	hints := ClientHints{
		Brands: []ClientHintBrand{
			{Name: "Google Chrome", Version: "122.0.6261.128"},
		},
		Platform: "Android",
		Mobile:   boolPtr(true),
	}
	result := New().ParseWithClientHints("Googlebot/2.1", hints)
	if !result.Bot.IsBot {
		t.Fatal("expected bot detection")
	}
	if result.Client.Name != Unknown {
		t.Fatalf("expected unknown client for bot, got %q", result.Client.Name)
	}
	if result.Device.Type != "Bot" {
		t.Fatalf("expected bot device, got %q", result.Device.Type)
	}
}

func TestParseNormalizesWhitespace(t *testing.T) {
	t.Parallel()

	ua := "\t Mozilla/5.0 \n (Windows NT 10.0; Win64; x64; rv:124.0)\r\nGecko/20100101 Firefox/124.0 \t"
	result := New().Parse(ua)
	if result.UserAgent != "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0" {
		t.Fatalf("unexpected normalized user agent %q", result.UserAgent)
	}
	if result.Client.Name != "Firefox" {
		t.Fatalf("unexpected client name %q", result.Client.Name)
	}
}

func TestParseConcurrentAccess(t *testing.T) {
	t.Parallel()

	detector := New(WithResultCacheSize(32))
	userAgents := []string{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
		"Mozilla/5.0 (Linux; Android 14; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/122.0.6261.69 Mobile/15E148 Safari/604.1",
	}

	headers := map[string]string{
		"Sec-CH-UA":                  "\"Not(A:Brand\";v=\"99\", \"Microsoft Edge\";v=\"123.0.0.0\", \"Chromium\";v=\"123.0.0.0\"",
		"Sec-CH-UA-Platform":         "\"Windows\"",
		"Sec-CH-UA-Platform-Version": "\"15.0.0\"",
		"Sec-CH-UA-Mobile":           "?0",
	}

	const workers = 64
	const iterations = 200

	var wg sync.WaitGroup
	errors := make(chan string, workers*iterations)
	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				ua := userAgents[(worker+i)%len(userAgents)]
				result := detector.Parse(ua)
				if result.UserAgent == "" {
					errors <- "Parse returned empty user agent"
					return
				}

				headerResult := detector.ParseWithHeaders("Mozilla/5.0", headers)
				if headerResult.Client.Name != "Microsoft Edge" {
					errors <- "ParseWithHeaders returned unexpected client"
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
