package ddgo

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

type deviceYAMLModelRule struct {
	Regex  string `yaml:"regex"`
	Model  string `yaml:"model"`
	Device string `yaml:"device"`
}

type deviceYAMLBrandRule struct {
	Regex  string                `yaml:"regex"`
	Device string                `yaml:"device"`
	Model  string                `yaml:"model"`
	Models []deviceYAMLModelRule `yaml:"models"`
}

type deviceModelRule struct {
	pattern       *regexp2.Regexp
	modelTemplate string
	deviceType    string
}

type deviceRule struct {
	brand         string
	pattern       *regexp2.Regexp
	deviceType    string
	modelTemplate string
	models        []deviceModelRule
}

var deviceRuleSources = []string{
	"device/cameras.yml",
	"device/car_browsers.yml",
	"device/consoles.yml",
	"device/mobiles.yml",
	"device/notebooks.yml",
	"device/portable_media_player.yml",
	"device/shell_tv.yml",
	"device/televisions.yml",
}

var (
	deviceRulesOnce sync.Once
	deviceRules     []deviceRule
	deviceRulesErr  error
)

func parseDeviceSnapshot(ua string) (Device, bool) {
	rules, err := loadDeviceRules()
	if err != nil {
		return Device{}, false
	}

	for _, rule := range rules {
		topMatch, ok := matchRegexp2String(rule.pattern, ua)
		if !ok {
			continue
		}

		deviceType := normalizeDeviceType(rule.deviceType)
		model := normalizeModel(expandRuleTemplate(rule.modelTemplate, topMatch))
		for _, nested := range rule.models {
			nestedMatch, nestedOK := matchRegexp2String(nested.pattern, ua)
			if !nestedOK {
				continue
			}
			if nested.modelTemplate != "" {
				model = normalizeModel(expandRuleTemplate(nested.modelTemplate, nestedMatch))
			}
			if nested.deviceType != "" {
				deviceType = normalizeDeviceType(nested.deviceType)
			}
			break
		}

		return Device{
			Type:  deviceType,
			Brand: normalizeRuleField(rule.brand),
			Model: model,
		}, true
	}

	return Device{}, false
}

func loadDeviceRules() ([]deviceRule, error) {
	deviceRulesOnce.Do(func() {
		files, err := loadSnapshotFiles()
		if err != nil {
			deviceRulesErr = err
			return
		}

		compiled := make([]deviceRule, 0, 2048)
		for _, path := range deviceRuleSources {
			content, ok := files[path]
			if !ok {
				deviceRulesErr = missingSnapshotFileError(path)
				return
			}

			var root yaml.Node
			if err := yaml.Unmarshal([]byte(content), &root); err != nil {
				deviceRulesErr = fmt.Errorf("decode %s: %w", path, err)
				return
			}
			if len(root.Content) == 0 {
				continue
			}

			mapping := root.Content[0]
			if mapping.Kind != yaml.MappingNode {
				continue
			}

			for i := 0; i+1 < len(mapping.Content); i += 2 {
				brandNode := mapping.Content[i]
				ruleNode := mapping.Content[i+1]

				var yamlRule deviceYAMLBrandRule
				if err := ruleNode.Decode(&yamlRule); err != nil {
					continue
				}
				if strings.TrimSpace(yamlRule.Regex) == "" {
					continue
				}

				re, err := compileRuleRegex(yamlRule.Regex)
				if err != nil {
					continue
				}

				modelRules := make([]deviceModelRule, 0, len(yamlRule.Models))
				for _, modelRule := range yamlRule.Models {
					if strings.TrimSpace(modelRule.Regex) == "" {
						continue
					}
					modelRegex, modelErr := compileRuleRegex(modelRule.Regex)
					if modelErr != nil {
						continue
					}
					modelRules = append(modelRules, deviceModelRule{
						pattern:       modelRegex,
						modelTemplate: modelRule.Model,
						deviceType:    modelRule.Device,
					})
				}

				compiled = append(compiled, deviceRule{
					brand:         strings.TrimSpace(brandNode.Value),
					pattern:       re,
					deviceType:    yamlRule.Device,
					modelTemplate: yamlRule.Model,
					models:        modelRules,
				})
			}
		}

		deviceRules = compiled
	})

	if deviceRulesErr != nil {
		return nil, deviceRulesErr
	}
	return deviceRules, nil
}

func normalizeDeviceType(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "_", " ")))
	if normalized == "" {
		return Unknown
	}

	switch normalized {
	case "smartphone":
		return "Smartphone"
	case "feature phone":
		return "Feature Phone"
	case "phablet":
		return "Phablet"
	case "tablet":
		return "Tablet"
	case "desktop":
		return "Desktop"
	case "console":
		return "Console"
	case "tv":
		return "TV"
	case "camera":
		return "Camera"
	case "car browser":
		return "Car Browser"
	case "portable media player":
		return "Portable Media Player"
	case "smart display":
		return "Smart Display"
	case "smart speaker":
		return "Smart Speaker"
	case "peripheral":
		return "Peripheral"
	case "wearable":
		return "Wearable"
	default:
		return titleCaseWords(normalized)
	}
}

func normalizeModel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return Unknown
	}
	return strings.Join(strings.Fields(value), " ")
}

func titleCaseWords(value string) string {
	parts := strings.Fields(value)
	for i, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
