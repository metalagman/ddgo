package ddgo

import (
	"strings"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

func TestEmbeddedSnapshotIncludesMirroredRuleFiles(t *testing.T) {
	t.Parallel()

	files, err := loadSnapshotFiles()
	if err != nil {
		t.Fatalf("loadSnapshotFiles() failed: %v", err)
	}

	required := []string{
		"bots.yml",
		"client/browsers.yml",
		"client/feed_readers.yml",
		"client/hints/apps.yml",
		"client/hints/browsers.yml",
		"client/libraries.yml",
		"client/mediaplayers.yml",
		"client/mobile_apps.yml",
		"client/pim.yml",
		"device/cameras.yml",
		"device/car_browsers.yml",
		"device/consoles.yml",
		"device/mobiles.yml",
		"device/notebooks.yml",
		"device/portable_media_player.yml",
		"device/shell_tv.yml",
		"device/televisions.yml",
		"oss.yml",
		"vendorfragments.yml",
	}
	for _, path := range required {
		content, ok := files[path]
		if !ok {
			t.Fatalf("missing required snapshot file %q", path)
		}
		if strings.TrimSpace(content) == "" {
			t.Fatalf("required snapshot file %q is empty", path)
		}
	}
}

func TestParseBotUsesSnapshotRules(t *testing.T) {
	t.Parallel()

	ua := "monitoring360bot/1.1"
	result := mustParse(t, newTestDetector(t), ua)

	if !result.Bot.IsBot {
		t.Fatalf("expected bot detection for %q", ua)
	}
	if result.Bot.Name != "360 Monitoring" {
		t.Fatalf("unexpected bot name %q", result.Bot.Name)
	}
	if result.Bot.Category != "Site Monitor" {
		t.Fatalf("unexpected bot category %q", result.Bot.Category)
	}
	if result.Bot.Producer.Name != "Plesk International GmbH" {
		t.Fatalf("unexpected producer %+v", result.Bot.Producer)
	}
}

func TestSnapshotRegexEntriesCompile(t *testing.T) {
	t.Parallel()

	files, err := loadSnapshotFiles()
	if err != nil {
		t.Fatalf("loadSnapshotFiles() failed: %v", err)
	}

	totalRegex := 0
	for path, content := range files {
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(content), &node); err != nil {
			t.Fatalf("yaml unmarshal failed for %q: %v", path, err)
		}

		var regexes []string
		collectRegexValues(&node, &regexes)
		for _, pattern := range regexes {
			re, err := regexp2.Compile(normalizeRulePattern(pattern), 0)
			if err != nil {
				t.Fatalf("regex compile failed for %q pattern %q: %v", path, pattern, err)
			}
			re.MatchTimeout = 25 * time.Millisecond
		}
		totalRegex += len(regexes)
	}

	if totalRegex < 100 {
		t.Fatalf("expected substantial regex coverage from snapshot files, got %d patterns", totalRegex)
	}
}

func collectRegexValues(node *yaml.Node, out *[]string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			collectRegexValues(child, out)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			if key.Kind == yaml.ScalarNode && key.Value == "regex" && value.Kind == yaml.ScalarNode {
				*out = append(*out, value.Value)
			}
			collectRegexValues(value, out)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			collectRegexValues(child, out)
		}
	}
}
