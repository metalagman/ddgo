package ddgo

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const upstreamBotFixturePath = "testdata/upstream/fixtures/bots.yml"

type upstreamBotFixture struct {
	UserAgent string `yaml:"user_agent"`
	Bot       struct {
		Name     string `yaml:"name"`
		Category string `yaml:"category"`
		URL      string `yaml:"url"`
		Producer struct {
			Name string `yaml:"name"`
			URL  string `yaml:"url"`
		} `yaml:"producer"`
	} `yaml:"bot"`
}

func TestUpstreamBotFixtures(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping large upstream bot fixture test in short mode")
	}

	fixtures := loadUpstreamBotFixtures(t)
	detector := New()

	var detectionMisses []string
	var nameMismatches []string
	detectedBots := 0
	nameMatches := 0
	for i, fixture := range fixtures {
		result := detector.Parse(fixture.UserAgent)
		if !result.Bot.IsBot {
			detectionMisses = append(detectionMisses, fmt.Sprintf("#%d: expected bot detection for %q", i, fixture.UserAgent))
			continue
		}
		detectedBots++

		if fixture.Bot.Name == "" {
			continue
		}
		if result.Bot.Name == fixture.Bot.Name {
			nameMatches++
			continue
		}
		nameMismatches = append(nameMismatches, fmt.Sprintf("#%d: expected bot name %q got %q for %q", i, fixture.Bot.Name, result.Bot.Name, fixture.UserAgent))
	}

	total := len(fixtures)
	if total == 0 {
		t.Fatal("no fixtures loaded")
	}
	detectionRate := float64(detectedBots) / float64(total)
	nameMatchRate := float64(nameMatches) / float64(total)

	const maxShown = 20
	if detectionRate < 0.95 {
		shown := detectionMisses
		if len(shown) > maxShown {
			shown = shown[:maxShown]
		}
		t.Fatalf("bot detection rate too low: %.2f%% (%d/%d)\n%s", detectionRate*100, detectedBots, total, strings.Join(shown, "\n"))
	}
	if nameMatchRate < 0.80 {
		shown := nameMismatches
		if len(shown) > maxShown {
			shown = shown[:maxShown]
		}
		t.Fatalf("bot name match rate too low: %.2f%% (%d/%d)\n%s", nameMatchRate*100, nameMatches, total, strings.Join(shown, "\n"))
	}
}

func loadUpstreamBotFixtures(t *testing.T) []upstreamBotFixture {
	t.Helper()

	raw, err := os.ReadFile(upstreamBotFixturePath)
	if err != nil {
		t.Fatalf("read upstream bot fixtures: %v", err)
	}

	var fixtures []upstreamBotFixture
	if err := yaml.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode upstream bot fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("no fixtures in %s", upstreamBotFixturePath)
	}

	for i, fixture := range fixtures {
		if strings.TrimSpace(fixture.UserAgent) == "" {
			t.Fatalf("fixture %d has empty user_agent", i)
		}
		if strings.TrimSpace(fixture.Bot.Name) == "" {
			t.Fatalf("fixture %d has empty bot name", i)
		}
	}
	return fixtures
}
