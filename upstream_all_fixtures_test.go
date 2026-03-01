package ddgo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"gopkg.in/yaml.v3"
)

const upstreamFixturesDir = "testdata/upstream/fixtures"

type upstreamFixtureCase struct {
	UserAgent string         `yaml:"user_agent"`
	Headers   map[string]any `yaml:"headers"`
	Bot       map[string]any `yaml:"bot"`
	Client    map[string]any `yaml:"client"`
	OS        any            `yaml:"os"`
	Device    map[string]any `yaml:"device"`
}

type parityMetric struct {
	label     string
	threshold float64
	expected  int
	matched   int
	misses    []string
}

func (m *parityMetric) compare(path string, index int, field, expected, actual string, normalizer func(string) string) {
	normalizedExpected := normalizer(expected)
	if normalizedExpected == "" {
		return
	}
	m.expected++

	normalizedActual := normalizer(actual)
	if normalizedExpected == normalizedActual {
		m.matched++
		return
	}
	if len(m.misses) < 8 {
		m.misses = append(
			m.misses,
			fmt.Sprintf("%s case %d %s: expected %q got %q", path, index, field, expected, actual),
		)
	}
}

func (m *parityMetric) compareBool(path string, index int, field string, expected, actual bool) {
	m.expected++
	if expected == actual {
		m.matched++
		return
	}
	if len(m.misses) < 8 {
		m.misses = append(
			m.misses,
			fmt.Sprintf("%s case %d %s: expected %t got %t", path, index, field, expected, actual),
		)
	}
}

func (m *parityMetric) rate() float64 {
	if m.expected == 0 {
		return 1
	}
	return float64(m.matched) / float64(m.expected)
}

func TestUpstreamAllFixturesSmoke(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping full upstream corpus test in short mode")
	}

	paths := listUpstreamFixturePaths(t)
	if len(paths) < 80 {
		t.Fatalf("unexpected fixture count: got %d, want at least 80", len(paths))
	}

	detector := New()
	totalCases := 0
	headerCases := 0
	expectedBotCases := 0
	detectedBotCases := 0

	for _, path := range paths {
		fixtures := loadUpstreamFixtureFile(t, path)
		totalCases += len(fixtures)

		for i, fixture := range fixtures {
			ua := strings.TrimSpace(fixture.UserAgent)
			headers := normalizeFixtureHeaders(fixture.Headers)
			if ua == "" && len(headers) == 0 {
				t.Fatalf("%s case %d has no user_agent and no headers", path, i)
			}

			var result Result
			if len(headers) > 0 {
				headerCases++
				result = detector.ParseWithHeaders(fixture.UserAgent, headers)
			} else {
				result = detector.Parse(fixture.UserAgent)
			}

			if ua != "" && strings.TrimSpace(result.UserAgent) == "" {
				t.Fatalf("%s case %d parsed empty user-agent result", path, i)
			}

			expectedBotName := mapStringValue(fixture.Bot, "name")
			if expectedBotName != "" {
				expectedBotCases++
				if result.Bot.IsBot {
					detectedBotCases++
				}
			}
		}
	}

	if totalCases < 5000 {
		t.Fatalf("unexpected corpus size: got %d cases, want at least 5000", totalCases)
	}
	if headerCases == 0 {
		t.Fatal("expected client-hints header fixture cases, found none")
	}

	botDetectionRate := 1.0
	if expectedBotCases > 0 {
		botDetectionRate = float64(detectedBotCases) / float64(expectedBotCases)
	}
	if botDetectionRate < 0.95 {
		t.Fatalf(
			"bot detection rate too low on full upstream corpus: %.2f%% (%d/%d)",
			botDetectionRate*100,
			detectedBotCases,
			expectedBotCases,
		)
	}
}

func TestUpstreamFixtureDifferentialParity(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping upstream parity test in short mode")
	}

	metrics := []*parityMetric{
		{label: "bot detection", threshold: 0.96},
		{label: "bot name", threshold: 0.85},
		{label: "client type", threshold: 0.80},
		{label: "client name", threshold: 0.60},
		{label: "client version", threshold: 0.45},
		{label: "os name", threshold: 0.75},
		{label: "os version", threshold: 0.40},
		{label: "device type", threshold: 0.70},
		{label: "device brand", threshold: 0.55},
		{label: "device model", threshold: 0.35},
	}

	metricByLabel := map[string]*parityMetric{}
	for _, metric := range metrics {
		metricByLabel[metric.label] = metric
	}

	detector := New()
	totalCases := 0
	for _, path := range listUpstreamFixturePaths(t) {
		fixtures := loadUpstreamFixtureFile(t, path)
		totalCases += len(fixtures)

		for i, fixture := range fixtures {
			headers := normalizeFixtureHeaders(fixture.Headers)
			result := detector.Parse(fixture.UserAgent)
			if len(headers) > 0 {
				result = detector.ParseWithHeaders(fixture.UserAgent, headers)
			}

			expectedBotName := mapStringValue(fixture.Bot, "name")
			if expectedBotName != "" {
				metricByLabel["bot detection"].compareBool(path, i, "bot.is_bot", true, result.Bot.IsBot)
				metricByLabel["bot name"].compare(path, i, "bot.name", expectedBotName, result.Bot.Name, normalizeNameValue)
			}

			metricByLabel["client type"].compare(
				path, i, "client.type", mapStringValue(fixture.Client, "type"), result.Client.Type, normalizeClientType,
			)
			metricByLabel["client name"].compare(
				path, i, "client.name", mapStringValue(fixture.Client, "name"), result.Client.Name, normalizeNameValue,
			)
			metricByLabel["client version"].compare(
				path, i, "client.version", mapStringValue(fixture.Client, "version"), result.Client.Version, normalizeVersionValue,
			)

			expectedOS := normalizeFixtureMap(fixture.OS)
			metricByLabel["os name"].compare(
				path, i, "os.name", mapStringValue(expectedOS, "name"), result.OS.Name, normalizeOSName,
			)
			metricByLabel["os version"].compare(
				path, i, "os.version", mapStringValue(expectedOS, "version"), result.OS.Version, normalizeVersionValue,
			)

			metricByLabel["device type"].compare(
				path, i, "device.type", mapStringValue(fixture.Device, "type"), result.Device.Type, normalizeDeviceTypeValue,
			)
			metricByLabel["device brand"].compare(
				path, i, "device.brand", mapStringValue(fixture.Device, "brand"), result.Device.Brand, normalizeNameValue,
			)
			metricByLabel["device model"].compare(
				path, i, "device.model", mapStringValue(fixture.Device, "model"), result.Device.Model, normalizeModelValue,
			)
		}
	}

	if totalCases < 5000 {
		t.Fatalf("unexpected corpus size: got %d cases, want at least 5000", totalCases)
	}

	var failures []string
	for _, metric := range metrics {
		rate := metric.rate()
		if rate >= metric.threshold {
			continue
		}
		failures = append(
			failures,
			fmt.Sprintf(
				"%s parity too low: %.2f%% (%d/%d), threshold %.2f%%",
				metric.label,
				rate*100,
				metric.matched,
				metric.expected,
				metric.threshold*100,
			),
		)
		if len(metric.misses) > 0 {
			failures = append(failures, strings.Join(metric.misses, "\n"))
		}
	}

	if len(failures) > 0 {
		t.Fatalf("upstream differential parity failed:\n%s", strings.Join(failures, "\n"))
	}
}

func listUpstreamFixturePaths(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(upstreamFixturesDir)
	if err != nil {
		t.Fatalf("read upstream fixtures dir: %v", err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		paths = append(paths, filepath.Join(upstreamFixturesDir, entry.Name()))
	}
	sort.Strings(paths)
	return paths
}

func loadUpstreamFixtureFile(t *testing.T, path string) []upstreamFixtureCase {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file %s: %v", path, err)
	}
	var fixtures []upstreamFixtureCase
	if err := yaml.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode fixture file %s: %v", path, err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("no fixtures in %s", path)
	}
	return fixtures
}

func mapStringValue(m map[string]any, key string) string {
	if len(m) == 0 {
		return ""
	}
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", typed)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", typed)
	case float32, float64:
		return fmt.Sprintf("%v", typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func normalizeFixtureMap(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case map[any]any:
		normalized := make(map[string]any, len(typed))
		for key, val := range typed {
			normalized[fmt.Sprintf("%v", key)] = val
		}
		return normalized
	default:
		return nil
	}
}

func normalizeFixtureHeaders(raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	headers := make(map[string]string, len(raw))
	for key, value := range raw {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			headers[key] = typed
		case []any:
			parts := make([]string, 0, len(typed))
			for _, item := range typed {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			headers[key] = strings.Join(parts, ", ")
		default:
			headers[key] = fmt.Sprintf("%v", typed)
		}
	}
	return headers
}

func normalizeNameValue(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == strings.ToLower(Unknown) {
		return ""
	}
	value = strings.ReplaceAll(value, "_", " ")
	return strings.Join(strings.Fields(value), " ")
}

func normalizeModelValue(value string) string {
	return normalizeNameValue(value)
}

func normalizeVersionValue(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == strings.ToLower(Unknown) {
		return ""
	}
	value = strings.TrimPrefix(value, "v")
	value = strings.ReplaceAll(value, "_", ".")
	value = strings.Trim(value, ". ")
	for strings.HasSuffix(value, ".0") {
		value = strings.TrimSuffix(value, ".0")
	}
	return value
}

func normalizeClientType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == strings.ToLower(Unknown) {
		return ""
	}
	value = strings.ReplaceAll(value, "_", " ")
	switch strings.Join(strings.Fields(value), " ") {
	case "browser":
		return "browser"
	case "feed reader":
		return "feed reader"
	case "mobile app":
		return "mobile app"
	case "media player", "mediaplayer":
		return "media player"
	case "pim":
		return "pim"
	case "library":
		return "library"
	default:
		return value
	}
}

func normalizeOSName(value string) string {
	switch normalizeNameValue(value) {
	case "gnu/linux":
		return "linux"
	case "mac os", "macos", "mac":
		return "mac"
	default:
		return normalizeNameValue(value)
	}
}

func normalizeDeviceTypeValue(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == strings.ToLower(Unknown) {
		return ""
	}
	value = strings.ReplaceAll(value, "_", " ")
	switch strings.Join(strings.Fields(value), " ") {
	case "smartphone":
		return "smartphone"
	case "feature phone":
		return "feature phone"
	case "phablet":
		return "phablet"
	case "tablet":
		return "tablet"
	case "desktop":
		return "desktop"
	case "console":
		return "console"
	case "tv", "television":
		return "tv"
	case "camera":
		return "camera"
	case "car browser":
		return "car browser"
	case "portable media player":
		return "portable media player"
	case "smart display":
		return "smart display"
	case "smart speaker":
		return "smart speaker"
	case "wearable":
		return "wearable"
	case "peripheral":
		return "peripheral"
	case "bot":
		return "bot"
	default:
		return value
	}
}

func TestUpstreamFixtureFilesAreReadable(t *testing.T) {
	t.Parallel()

	var unreadable []string
	err := filepath.WalkDir(upstreamFixturesDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			unreadable = append(unreadable, fmt.Sprintf("%s: %v", path, walkErr))
			return nil
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".yml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			unreadable = append(unreadable, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			unreadable = append(unreadable, fmt.Sprintf("%s: empty", path))
			return nil
		}
		if unicode.IsSpace(rune(data[0])) {
			// no-op; explicit read/inspection to avoid accidental binary files
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk upstream fixtures: %v", err)
	}
	if len(unreadable) > 0 {
		t.Fatalf("unreadable upstream fixture files:\n%s", strings.Join(unreadable, "\n"))
	}
}
