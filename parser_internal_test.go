package ddgo

import (
	"strings"
	"testing"
)

func TestMapWindowsVersion(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"10.0", "10"},
		{"6.3", "8.1"},
		{"6.2", "8"},
		{"6.1", "7"},
		{"5.1", "5.1"},
		{"unknown", "unknown"},
	}

	for _, tc := range cases {
		got := mapWindowsVersion(tc.input)
		if got != tc.expected {
			t.Errorf("mapWindowsVersion(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestWindowsPlatform(t *testing.T) {
	cases := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "x64"},
		{"Mozilla/5.0 (Windows NT 10.0; WOW64)", "x64"},
		{"Mozilla/5.0 (Windows NT 6.1; Win32; x86)", "x86"},
		{"Mozilla/5.0 (Windows NT 6.1)", "x86"},
	}

	for _, tc := range cases {
		got := windowsPlatform(tc.ua)
		if got != tc.expected {
			t.Errorf("windowsPlatform(%q) = %q; want %q", tc.ua, got, tc.expected)
		}
	}
}

func TestTitleCaseWords(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"FEATURE phone", "FEATURE Phone"},
		{"smart speaker", "Smart Speaker"},
		{"", ""},
		{"a b c", "A B C"},
	}

	for _, tc := range cases {
		got := titleCaseWords(tc.input)
		if got != tc.expected {
			t.Errorf("titleCaseWords(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestCanonicalClientName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"microsoft edge", "Microsoft Edge"},
		{"EDGE", "Microsoft Edge"},
		{"Opera", "Opera"},
		{"Google Chrome", "Chrome"},
		{"chrome", "Chrome"},
		{"chromium", "Chrome"},
		{"Mozilla Firefox", "Firefox"},
		{"firefox", "Firefox"},
		{"Safari", "Safari"},
		{"Unknown Browser", "Unknown Browser"},
	}

	for _, tc := range cases {
		got := canonicalClientName(tc.input)
		if got != tc.expected {
			t.Errorf("canonicalClientName(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestCanonicalOSName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"windows", "Windows"},
		{"Android", "Android"},
		{"iOS", "iOS"},
		{"macos", "macOS"},
		{"Mac OS X", "macOS"},
		{"linux", "Linux"},
		{"Chrome OS", "Chrome OS"},
		{"cros", "Chrome OS"},
		{"Unknown OS", "Unknown OS"},
	}

	for _, tc := range cases {
		got := canonicalOSName(tc.input)
		if got != tc.expected {
			t.Errorf("canonicalOSName(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestInferBrandFromModel(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"SM-G991B", "Samsung"},
		{"Pixel 6", "Google"},
		{"pixel 5", "Google"},
		{"iPhone 13", "Apple"},
		{"iPad Air", "Apple"},
		{"Unknown Model", "Unknown"},
		{"", "Unknown"},
		{Unknown, "Unknown"},
	}

	for _, tc := range cases {
		got := inferBrandFromModel(tc.input)
		if got != tc.expected {
			t.Errorf("inferBrandFromModel(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestLegacyParsers(t *testing.T) {
	// Test parseBotLegacy
	bot := parseBotLegacy("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	if !bot.IsBot || bot.Name != "Googlebot" {
		t.Errorf("parseBotLegacy(Googlebot) = %+v; want IsBot:true, Name:Googlebot", bot)
	}
	bot = parseBotLegacy("Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)")
	if !bot.IsBot || bot.Name != "bingbot" {
		t.Errorf("parseBotLegacy(bingbot) = %+v; want IsBot:true, Name:bingbot", bot)
	}
	bot = parseBotLegacy("DuckDuckBot/1.1; (+http://duckduckgo.com/duckduckbot.html)")
	if !bot.IsBot || bot.Name != "DuckDuckBot" {
		t.Errorf("parseBotLegacy(DuckDuckBot) = %+v; want IsBot:true, Name:DuckDuckBot", bot)
	}
	bot = parseBotLegacy("Mozilla/5.0 (compatible; Generic bot/1.0)")
	if !bot.IsBot || bot.Name != "Generic Bot" {
		t.Errorf("parseBotLegacy(Generic Bot) = %+v; want IsBot:true, Name:Generic Bot", bot)
	}
	bot = parseBotLegacy("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if bot.IsBot {
		t.Errorf("parseBotLegacy(Normal UA) = %+v; want IsBot:false", bot)
	}

	// Test parseClientLegacy
	client := parseClientLegacy("curl/7.64.1")
	if client.Name != "curl" || client.Type != "Library" {
		t.Errorf("parseClientLegacy(curl) = %+v; want Name:curl, Type:Library", client)
	}
	client = parseClientLegacy("Go-http-client/1.1")
	if client.Name != "Go HTTP Client" || client.Type != "Library" {
		t.Errorf("parseClientLegacy(Go HTTP) = %+v; want Name:Go HTTP Client, Type:Library", client)
	}

	// Test parseLegacyBrowserClient
	client, ok := parseLegacyBrowserClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edg/124.0.0.0")
	if !ok || client.Name != "Microsoft Edge" || client.Engine != "Blink" {
		t.Errorf("parseLegacyBrowserClient(Edge) = %+v, %v; want Name:Microsoft Edge, Engine:Blink", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64) OPR/110.0.0.0")
	if !ok || client.Name != "Opera" || client.Engine != "Blink" {
		t.Errorf("parseLegacyBrowserClient(Opera) = %+v, %v; want Name:Opera, Engine:Blink", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Android 11; Mobile; rv:94.0) Gecko/94.0 Firefox/94.0")
	if !ok || client.Name != "Firefox" || client.Engine != "Gecko" {
		t.Errorf("parseLegacyBrowserClient(Firefox) = %+v, %v; want Name:Firefox, Engine:Gecko", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Mobile/15E148 Safari/604.1")
	if !ok || client.Name != "Safari" || client.Engine != "WebKit" {
		t.Errorf("parseLegacyBrowserClient(Safari) = %+v, %v; want Name:Safari, Engine:WebKit", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	if !ok || client.Name != "Chrome" || client.Engine != "Blink" {
		t.Errorf("parseLegacyBrowserClient(Chrome) = %+v, %v; want Name:Chrome, Engine:Blink", client, ok)
	}

	// Test parseOSLegacy
	os := parseOSLegacy("Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	if os.Name != "Windows" || os.Version != "10" {
		t.Errorf("parseOSLegacy(Windows) = %+v; want Name:Windows, Version:10", os)
	}
	os = parseOSLegacy("Mozilla/5.0 (Android 11; Mobile; rv:94.0) Gecko/94.0 Firefox/94.0")
	if os.Name != "Android" || os.Version != "11" {
		t.Errorf("parseOSLegacy(Android) = %+v; want Name:Android, Version:11", os)
	}
	os = parseOSLegacy("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X)")
	if os.Name != "iOS" || os.Version != "15.1" {
		t.Errorf("parseOSLegacy(iOS) = %+v; want Name:iOS, Version:15.1", os)
	}
	os = parseOSLegacy("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	if os.Name != "macOS" || os.Version != "10.15.7" {
		t.Errorf("parseOSLegacy(macOS) = %+v; want Name:macOS, Version:10.15.7", os)
	}
	os = parseOSLegacy("Mozilla/5.0 (X11; Linux x86_64)")
	if os.Name != "Linux" {
		t.Errorf("parseOSLegacy(Linux) = %+v; want Name:Linux", os)
	}

	// Test parseDeviceLegacy
	device := parseDeviceLegacy("Mozilla/5.0 (iPad; CPU OS 15_1 like Mac OS X)")
	if device.Type != "Tablet" || device.Brand != "Apple" {
		t.Errorf("parseDeviceLegacy(iPad) = %+v; want Type:Tablet, Brand:Apple", device)
	}
	device = parseDeviceLegacy("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X)")
	if device.Type != "Smartphone" || device.Brand != "Apple" {
		t.Errorf("parseDeviceLegacy(iPhone) = %+v; want Type:Smartphone, Brand:Apple", device)
	}
	device = parseDeviceLegacy("Mozilla/5.0 (Linux; Android 11; Mobile; SM-G991B)")
	if device.Type != "Smartphone" || device.Brand != "Samsung" {
		t.Errorf("parseDeviceLegacy(Samsung) = %+v; want Type:Smartphone, Brand:Samsung", device)
	}
	device = parseDeviceLegacy("Mozilla/5.0 (Linux; Android 11; Mobile; Pixel 6)")
	if device.Type != "Smartphone" || device.Brand != "Google" {
		t.Errorf("parseDeviceLegacy(Pixel) = %+v; want Type:Smartphone, Brand:Google", device)
	}
	device = parseDeviceLegacy("Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	if device.Type != "Desktop" {
		t.Errorf("parseDeviceLegacy(Desktop) = %+v; want Type:Desktop", device)
	}
}

func TestNestedOSVersion(t *testing.T) {
	detector, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Smartisan OS SM801 -> 2.5
	ua := "Mozilla/5.0 (Linux; Android 11; SM801) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36"
	result, err := detector.Parse(ua)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.OS.Name != "Smartisan OS" || result.OS.Version != "2.5" {
		t.Errorf("Parse(Smartisan SM801) = %+v; want OS Name:Smartisan OS, Version:2.5", result.OS)
	}
}

func TestClientEngineDetection(t *testing.T) {
	detector, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cases := []struct {
		ua      string
		engine  string
		version string
	}{
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			engine:  "Blink",
			version: "124.0.0.0",
		},
		{
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
			engine:  "Gecko",
			version: "125.0",
		},
		{
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
			engine:  "WebKit",
			version: "605.1.15",
		},
		{
			ua:      "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
			engine:  "Trident",
			version: "6.0",
		},
		{
			ua:      "Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.10",
			engine:  "Presto",
			version: "2.8.131",
		},
	}

	for _, tc := range cases {
		result, err := detector.Parse(tc.ua)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", tc.ua, err)
			continue
		}
		if result.Client.Engine != tc.engine {
			t.Errorf("Parse(%q) engine = %q; want %q", tc.ua, result.Client.Engine, tc.engine)
		}
		if result.Client.EngineVersion != tc.version {
			t.Errorf("Parse(%q) engine version = %q; want %q", tc.ua, result.Client.EngineVersion, tc.version)
		}
	}
}

func TestExtractEngineVersionAll(t *testing.T) {
	cases := []struct {
		ua      string
		engine  string
		version string
		want    string
	}{
		{"Chrome/124.0.0.0", "Blink", "124", "124.0.0.0"},
		{"OPR/110.0.0.0", "Blink", "110", "110.0.0.0"},
		{"AppleWebKit/605.1.15", "WebKit", "", "605.1.15"},
		{"rv:125.0", "Gecko", "", "125.0"},
		{"Trident/7.0", "Trident", "", "7.0"},
		{"Presto/2.12.388", "Presto", "", "2.12.388"},
		{"NetFront/3.5", "NetFront", "", "3.5"},
		{"Goanna/4.1.0", "Goanna", "", "4.1.0"},
		{"Servo/0.0.1", "Servo", "", "0.0.1"},
		{"EkiohFlow/2.1.0", "EkiohFlow", "", "2.1.0"},
		{"xChaos_Arachne/1.0", "Arachne", "", "1.0"},
		{"Unknown/1.0", "Unknown", "", ""},
	}

	for _, tc := range cases {
		got := extractEngineVersion(tc.ua, tc.engine, tc.version)
		if got != tc.want {
			t.Errorf("extractEngineVersion(%q, %q, %q) = %q; want %q", tc.ua, tc.engine, tc.version, got, tc.want)
		}
	}
}

func TestMissingSnapshotFileError(t *testing.T) {
	err := missingSnapshotFileError("test.yml")
	if err == nil || !strings.Contains(err.Error(), "test.yml") {
		t.Errorf("missingSnapshotFileError() = %v; want error containing test.yml", err)
	}
}

func TestLoadRulesMissingFiles(t *testing.T) {
	mockMap := map[string]string{
		"other.yml": "[]",
	}
	_, err := loadBotRules(mockMap)
	if err == nil {
		t.Error("loadBotRules() expected error for missing bots.yml")
	}
	_, err = loadClientRules(mockMap)
	if err == nil {
		t.Error("loadClientRules() expected error for missing files")
	}
	_, err = loadClientEngineRules(mockMap)
	if err == nil {
		t.Error("loadClientEngineRules() expected error for missing files")
	}
	_, err = loadOSRules(mockMap)
	if err == nil {
		t.Error("loadOSRules() expected error for missing files")
	}
	_, err = loadDeviceRules(mockMap)
	if err == nil {
		t.Error("loadDeviceRules() expected error for missing files")
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.1", "1.0.9", 1},
		{"1.0.9", "1.1", -1},
		{"2", "1.9.9", 1},
		{"1.2.3", "1.2.3.4", -1},
		{"1.2.3.4", "1.2.3", 1},
		{"invalid", "1.0.0", -1},
	}

	for _, tc := range cases {
		got := compareVersions(tc.v1, tc.v2)
		if got != tc.want {
			t.Errorf("compareVersions(%q, %q) = %d; want %d", tc.v1, tc.v2, got, tc.want)
		}
	}
}
