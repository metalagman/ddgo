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
