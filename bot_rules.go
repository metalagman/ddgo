package ddgo

import (
	"fmt"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

type botYAMLRule struct {
	Regex    string `yaml:"regex"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
	URL      string `yaml:"url"`
	Producer struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"producer"`
}

type botRule struct {
	pattern  *regexp2.Regexp
	name     string
	category string
	url      string
	producer Producer
}

var (
	botRulesOnce sync.Once
	botRules     []botRule
	botRulesErr  error
)

func loadBotRules() ([]botRule, error) {
	botRulesOnce.Do(func() {
		files, err := loadSnapshotFiles()
		if err != nil {
			botRulesErr = err
			return
		}
		content, ok := files["bots.yml"]
		if !ok {
			botRulesErr = fmt.Errorf("compiled snapshot missing bots.yml")
			return
		}

		var yamlRules []botYAMLRule
		if err := yaml.Unmarshal([]byte(content), &yamlRules); err != nil {
			botRulesErr = fmt.Errorf("decode bots.yml: %w", err)
			return
		}

		compiled := make([]botRule, 0, len(yamlRules))
		for _, item := range yamlRules {
			if item.Regex == "" {
				continue
			}
			re, err := regexp2.Compile(normalizeRulePattern(item.Regex), 0)
			if err != nil {
				// Keep parser operational even if one upstream rule is invalid for current engine mode.
				continue
			}
			re.MatchTimeout = 100 * time.Millisecond

			compiled = append(compiled, botRule{
				pattern:  re,
				name:     defaultString(item.Name, Unknown),
				category: defaultString(item.Category, Unknown),
				url:      defaultString(item.URL, Unknown),
				producer: Producer{
					Name: defaultString(item.Producer.Name, Unknown),
					URL:  defaultString(item.Producer.URL, Unknown),
				},
			})
		}
		botRules = compiled
	})

	if botRulesErr != nil {
		return nil, botRulesErr
	}
	return botRules, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
