package ddgo

import (
	"errors"
	"fmt"
	"strings"

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

var errBotsSnapshotMissing = errors.New("compiled snapshot missing bots.yml")

func loadBotRules(files map[string]string) ([]botRule, error) {
	content, ok := files["bots.yml"]
	if !ok {
		return nil, errBotsSnapshotMissing
	}

	var yamlRules []botYAMLRule
	err := yaml.Unmarshal([]byte(content), &yamlRules)
	if err != nil {
		return nil, fmt.Errorf("decode bots.yml: %w", err)
	}

	compiled := make([]botRule, 0, len(yamlRules))
	for _, item := range yamlRules {
		if item.Regex == "" {
			continue
		}
		re, compileErr := compileRuleRegex(item.Regex)
		if compileErr != nil {
			return nil, fmt.Errorf("compile bots.yml regex %q: %w", item.Regex, compileErr)
		}

		compiled = append(compiled, botRule{
			pattern:  re,
			name:     normalizeUnknownString(item.Name),
			category: normalizeUnknownString(item.Category),
			url:      normalizeUnknownString(item.URL),
			producer: Producer{
				Name: normalizeUnknownString(item.Producer.Name),
				URL:  normalizeUnknownString(item.Producer.URL),
			},
		})
	}
	return compiled, nil
}

func normalizeUnknownString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return Unknown
	}
	return value
}
