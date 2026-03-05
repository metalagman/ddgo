package ddgo

import (
	"strings"
	"testing"

	"github.com/dlclark/regexp2"
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
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/124.0.0.0 Safari/604.1")
	if !ok || client.Name != "Microsoft Edge" || client.Engine != "WebKit" {
		t.Errorf("parseLegacyBrowserClient(EdgiOS) = %+v, %v; want Name:Microsoft Edge, Engine:WebKit", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64) OPR/110.0.0.0")
	if !ok || client.Name != "Opera" || client.Engine != "Blink" {
		t.Errorf("parseLegacyBrowserClient(Opera) = %+v, %v; want Name:Opera, Engine:Blink", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Android 11; Mobile; rv:94.0) Gecko/94.0 Firefox/94.0")
	if !ok || client.Name != "Firefox" || client.Engine != "Gecko" {
		t.Errorf("parseLegacyBrowserClient(Firefox) = %+v, %v; want Name:Firefox, Engine:Gecko", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) FxiOS/94.0")
	if !ok || client.Name != "Firefox" || client.Engine != "WebKit" {
		t.Errorf("parseLegacyBrowserClient(FxiOS) = %+v, %v; want Name:Firefox, Engine:WebKit", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Mobile/15E148 Safari/604.1")
	if !ok || client.Name != "Safari" || client.Engine != "WebKit" {
		t.Errorf("parseLegacyBrowserClient(Safari) = %+v, %v; want Name:Safari, Engine:WebKit", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	if !ok || client.Name != "Chrome" || client.Engine != "Blink" {
		t.Errorf("parseLegacyBrowserClient(Chrome) = %+v, %v; want Name:Chrome, Engine:Blink", client, ok)
	}
	client, ok = parseLegacyBrowserClient("Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) CriOS/124.0.0.0")
	if !ok || client.Name != "Chrome" || client.Engine != "WebKit" {
		t.Errorf("parseLegacyBrowserClient(CriOS) = %+v, %v; want Name:Chrome, Engine:WebKit", client, ok)
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

func TestNormalizeDeviceType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"smartphone", "Smartphone"},
		{"FEATURE_PHONE", "Feature Phone"},
		{"phablet", "Phablet"},
		{"tablet", "Tablet"},
		{"desktop", "Desktop"},
		{"console", "Console"},
		{"tv", "TV"},
		{"camera", "Camera"},
		{"car browser", "Car Browser"},
		{"portable media player", "Portable Media Player"},
		{"smart display", "Smart Display"},
		{"smart speaker", "Smart Speaker"},
		{"peripheral", "Peripheral"},
		{"wearable", "Wearable"},
		{"unknown", "Unknown"},
		{"", "Unknown"},
		{"custom type", "Custom Type"},
	}

	for _, tc := range cases {
		got := normalizeDeviceType(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeDeviceType(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestInferDeviceType(t *testing.T) {
	btrue := true
	bfalse := false

	cases := []struct {
		mobile   *bool
		osName   string
		expected string
	}{
		{&btrue, "Android", "Smartphone"},
		{&bfalse, "Android", "Tablet"},
		{&bfalse, "iOS", "Tablet"},
		{&btrue, "Windows", "Smartphone"},
		{&bfalse, "Windows", "Desktop"},
		{&bfalse, "macOS", "Desktop"},
		{&bfalse, "Linux", "Desktop"},
		{&bfalse, "Chrome OS", "Desktop"},
		{&bfalse, "Other", "Unknown"},
		{nil, "Android", "Smartphone"},
		{nil, "iOS", "Smartphone"},
		{nil, "Windows", "Desktop"},
		{nil, "macOS", "Desktop"},
		{nil, "Linux", "Desktop"},
		{nil, "Chrome OS", "Desktop"},
		{nil, "Other", "Unknown"},
	}

	for _, tc := range cases {
		got := inferDeviceType(tc.mobile, tc.osName)
		if got != tc.expected {
			t.Errorf("inferDeviceType(%v, %q) = %q; want %q", tc.mobile, tc.osName, got, tc.expected)
		}
	}
}

func TestProfileForBrand(t *testing.T) {
	cases := []struct {
		brand string
		ok    bool
		name  string
	}{
		{"Edge", true, "Microsoft Edge"},
		{"Opera", true, "Opera"},
		{"Chrome", true, "Chrome"},
		{"Chromium", true, "Chrome"},
		{"Firefox", true, "Firefox"},
		{"Safari", true, "Safari"},
		{"Unknown", false, ""},
	}

	for _, tc := range cases {
		got, ok := profileForBrand(tc.brand)
		if ok != tc.ok {
			t.Errorf("profileForBrand(%q) ok = %v; want %v", tc.brand, ok, tc.ok)
		}
		if ok && got.Name != tc.name {
			t.Errorf("profileForBrand(%q) name = %q; want %q", tc.brand, got.Name, tc.name)
		}
	}
}

func TestPlatformFromUserAgent(t *testing.T) {
	cases := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (Linux; arm64; Android 14; SM-G991B)", "ARM"},
		{"Mozilla/5.0 (Linux; aarch64; Android 14)", "ARM"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "x64"},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64)", "x64"},
		{"Mozilla/5.0 (Windows NT 6.1; x86)", "x86"},
		{"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; i686)", "x86"},
		{"Unknown", ""},
	}

	for _, tc := range cases {
		got := platformFromUserAgent(strings.ToLower(tc.ua))
		if got != tc.expected {
			t.Errorf("platformFromUserAgent(%q) = %q; want %q", tc.ua, got, tc.expected)
		}
	}
}

func TestNewWithNilOption(t *testing.T) {
	d, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}
	if d == nil {
		t.Fatal("New(nil) returned nil detector")
	}
}

func TestDetectClientEngineLegacy(t *testing.T) {
	// Test the fallback switch in detectClientEngine
	cases := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36", "Blink"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edg/124.0.0.0 Safari/537.36", "Blink"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) EdgA/124.0.0.0", "Blink"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Mobile/15E148 Safari/604.1", "WebKit"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0", "Gecko"},
		{"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.7.5) Gecko/20041107 Firefox/1.0", "Gecko"},
		{"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)", "Trident"},
		{"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Trident/4.0)", "Trident"},
		{"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.10", "Presto"},
		{"Unknown", Unknown},
	}

	for _, tc := range cases {
		// Pass empty rules to trigger fallback switch
		got, err := detectClientEngine(tc.ua, []rune(tc.ua), nil)
		if err != nil {
			t.Errorf("detectClientEngine(%q) error = %v", tc.ua, err)
			continue
		}
		if got != tc.expected {
			t.Errorf("detectClientEngine(%q) = %q; want %q", tc.ua, got, tc.expected)
		}
	}
}

func TestExpandRuleTemplate(t *testing.T) {
	re := regexp2.MustCompile("(a)(b)", 0)
	match, _ := re.FindStringMatch("ab")

	cases := []struct {
		template string
		match    *regexp2.Match
		expected string
	}{
		{"$1$2", match, "ab"},
		{"$1 $2", match, "a b"},
		{"$1-$3", match, "a-"},
		{"$1-$x", match, "a-$x"},
		{"", match, ""},
		{"static", nil, "static"},
		{"$1", nil, "$1"},
	}

	for _, tc := range cases {
		got := expandRuleTemplate(tc.template, tc.match)
		if got != tc.expected {
			t.Errorf("expandRuleTemplate(%q) = %q; want %q", tc.template, got, tc.expected)
		}
	}
}

func TestNormalizeRuleVersion(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"1_2_3", "1.2.3"},
		{" 1.2.3 ", "1.2.3"},
		{"", Unknown},
		{" ", Unknown},
		{".", Unknown},
	}

	for _, tc := range cases {
		got := normalizeRuleVersion(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeRuleVersion(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestNormalizeRuleField(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{" value ", "value"},
		{"", Unknown},
		{" ", Unknown},
	}

	for _, tc := range cases {
		got := normalizeRuleField(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeRuleField(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestLoadRulesInvalidYAMLAll(t *testing.T) {
	invalidYAML := "[ invalid yaml"
	mockMap := map[string]string{
		"bots.yml":                   invalidYAML,
		"client/browsers.yml":        invalidYAML,
		"client/browser_engine.yml":  invalidYAML,
		"oss.yml":                    invalidYAML,
		"vendorfragments.yml":        invalidYAML,
		"client/feed_readers.yml":    invalidYAML,
		"client/mobile_apps.yml":     invalidYAML,
		"client/mediaplayers.yml":    invalidYAML,
		"client/pim.yml":             invalidYAML,
		"client/libraries.yml":       invalidYAML,
		"device/cameras.yml":         invalidYAML,
		"device/car_browsers.yml":    invalidYAML,
		"device/consoles.yml":        invalidYAML,
		"device/mobiles.yml":         invalidYAML,
		"device/notebooks.yml":       invalidYAML,
		"device/portable_media_player.yml": invalidYAML,
		"device/shell_tv.yml":        invalidYAML,
		"device/televisions.yml":     invalidYAML,
	}

	if _, err := loadBotRules(mockMap); err == nil {
		t.Error("loadBotRules() expected error")
	}
	if _, err := loadClientRules(mockMap); err == nil {
		t.Error("loadClientRules() expected error")
	}
	if _, err := loadClientEngineRules(mockMap); err == nil {
		t.Error("loadClientEngineRules() expected error")
	}
	if _, err := loadOSRules(mockMap); err == nil {
		t.Error("loadOSRules() expected error")
	}
	if _, err := loadDeviceRules(mockMap); err == nil {
		t.Error("loadDeviceRules() expected error")
	}
}

func TestParseBotMatch(t *testing.T) {
	detector, _ := New()
	// Googlebot
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	res, err := detector.Parse(ua)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !res.Bot.IsBot || res.Bot.Name != "Googlebot" {
		t.Errorf("Parse(Googlebot) = %+v; want Googlebot", res.Bot)
	}
}

func TestCacheEdgeCases(t *testing.T) {
	// Test NewLRUResultCache with invalid size
	c := NewLRUResultCache(0)
	if c != nil {
		t.Fatal("NewLRUResultCache(0) should return nil")
	}

	c = NewLRUResultCache(1)
	if c == nil {
		t.Fatal("NewLRUResultCache(1) should return cache")
	}
	c.Set("a", Result{})
	c.Set("b", Result{}) // Should evict "a"
	_, ok := c.Get("a")
	if ok {
		t.Error("expected 'a' to be evicted")
	}
}


