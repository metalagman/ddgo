package ddgo

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

const (
	deviceTypeSmartphone        = "Smartphone"
	deviceTypeFeaturePhone      = "Feature Phone"
	deviceTypePhablet           = "Phablet"
	deviceTypeTablet            = "Tablet"
	deviceTypeDesktop           = "Desktop"
	deviceTypeConsole           = "Console"
	deviceTypeTV                = "TV"
	deviceTypeCamera            = "Camera"
	deviceTypeCarBrowser        = "Car Browser"
	deviceTypePortableMedia     = "Portable Media Player"
	deviceTypeSmartDisplay      = "Smart Display"
	deviceTypeSmartSpeaker      = "Smart Speaker"
	deviceTypePeripheral        = "Peripheral"
	deviceTypeWearable          = "Wearable"
	estimatedDeviceRuleCapacity = 2048
	yamlMappingPairStride       = 2
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

func parseDeviceSnapshot(runtime *parserRuntime, ua string) (Device, bool, error) {
	for _, rule := range runtime.deviceRules {
		device, ok, err := parseDeviceFromRule(ua, rule)
		if err != nil {
			return Device{}, false, err
		}
		if ok {
			return device, true, nil
		}
	}

	return Device{}, false, nil
}

func parseDeviceFromRule(ua string, rule deviceRule) (Device, bool, error) {
	topMatch, ok, matchErr := matchRegexp2String(rule.pattern, ua)
	if matchErr != nil {
		return Device{}, false, fmt.Errorf("match device rule for brand %q: %w", rule.brand, matchErr)
	}
	if !ok {
		return Device{}, false, nil
	}

	deviceType := normalizeDeviceType(rule.deviceType)
	model := normalizeModel(expandRuleTemplate(rule.modelTemplate, topMatch))
	var err error
	deviceType, model, err = applyNestedDeviceRules(ua, rule.models, rule.brand, deviceType, model)
	if err != nil {
		return Device{}, false, err
	}

	return Device{
		Type:  deviceType,
		Brand: normalizeRuleField(rule.brand),
		Model: model,
	}, true, nil
}

func applyNestedDeviceRules(
	ua string,
	modelRules []deviceModelRule,
	brand, currentType, currentModel string,
) (deviceType, model string, err error) {
	deviceType = currentType
	model = currentModel

	for _, nested := range modelRules {
		nestedMatch, nestedOK, nestedErr := matchRegexp2String(nested.pattern, ua)
		if nestedErr != nil {
			return Unknown, Unknown, fmt.Errorf("match nested device rule for brand %q: %w", brand, nestedErr)
		}
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

	return deviceType, model, nil
}

func loadDeviceRules(files map[string]string) ([]deviceRule, error) {
	sources := deviceRuleSources()
	compiled := make([]deviceRule, 0, estimatedDeviceRuleCapacity)

	for _, path := range sources {
		content, ok := files[path]
		if !ok {
			return nil, missingSnapshotFileError(path)
		}

		rules, err := decodeDeviceRuleSource(content, path)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, rules...)
	}

	return compiled, nil
}

func decodeDeviceRuleSource(content, sourcePath string) ([]deviceRule, error) {
	var root yaml.Node
	err := yaml.Unmarshal([]byte(content), &root)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", sourcePath, err)
	}
	if len(root.Content) == 0 {
		return nil, nil
	}

	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return nil, nil
	}

	compiled := make([]deviceRule, 0, len(mapping.Content)/yamlMappingPairStride)
	for i := 0; i+1 < len(mapping.Content); i += yamlMappingPairStride {
		rule, ok, buildErr := compileDeviceRulePair(mapping.Content[i], mapping.Content[i+1], sourcePath)
		if buildErr != nil {
			return nil, buildErr
		}
		if ok {
			compiled = append(compiled, rule)
		}
	}

	return compiled, nil
}

func compileDeviceRulePair(brandNode, ruleNode *yaml.Node, sourcePath string) (deviceRule, bool, error) {
	var yamlRule deviceYAMLBrandRule
	err := ruleNode.Decode(&yamlRule)
	if err != nil {
		return deviceRule{}, false, fmt.Errorf(
			"decode %s rule for brand %q: %w",
			sourcePath,
			strings.TrimSpace(brandNode.Value),
			err,
		)
	}
	if strings.TrimSpace(yamlRule.Regex) == "" {
		return deviceRule{}, false, nil
	}

	re, err := compileRuleRegex(yamlRule.Regex)
	if err != nil {
		return deviceRule{}, false, fmt.Errorf("compile %s regex %q: %w", sourcePath, yamlRule.Regex, err)
	}

	modelRules, err := compileDeviceModelRules(yamlRule.Models, sourcePath)
	if err != nil {
		return deviceRule{}, false, err
	}

	return deviceRule{
		brand:         strings.TrimSpace(brandNode.Value),
		pattern:       re,
		deviceType:    yamlRule.Device,
		modelTemplate: yamlRule.Model,
		models:        modelRules,
	}, true, nil
}

func compileDeviceModelRules(models []deviceYAMLModelRule, sourcePath string) ([]deviceModelRule, error) {
	compiled := make([]deviceModelRule, 0, len(models))
	for _, modelRule := range models {
		if strings.TrimSpace(modelRule.Regex) == "" {
			continue
		}
		modelRegex, err := compileRuleRegex(modelRule.Regex)
		if err != nil {
			return nil, fmt.Errorf("compile %s model regex %q: %w", sourcePath, modelRule.Regex, err)
		}
		compiled = append(compiled, deviceModelRule{
			pattern:       modelRegex,
			modelTemplate: modelRule.Model,
			deviceType:    modelRule.Device,
		})
	}
	return compiled, nil
}

func deviceRuleSources() []string {
	return []string{
		"device/cameras.yml",
		"device/car_browsers.yml",
		"device/consoles.yml",
		"device/mobiles.yml",
		"device/notebooks.yml",
		"device/portable_media_player.yml",
		"device/shell_tv.yml",
		"device/televisions.yml",
	}
}

func normalizeDeviceType(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(raw, "_", " ")))
	if normalized == "" {
		return Unknown
	}

	switch normalized {
	case "smartphone":
		return deviceTypeSmartphone
	case "feature phone":
		return deviceTypeFeaturePhone
	case "phablet":
		return deviceTypePhablet
	case "tablet":
		return deviceTypeTablet
	case "desktop":
		return deviceTypeDesktop
	case "console":
		return deviceTypeConsole
	case "tv":
		return deviceTypeTV
	case "camera":
		return deviceTypeCamera
	case "car browser":
		return deviceTypeCarBrowser
	case "portable media player":
		return deviceTypePortableMedia
	case "smart display":
		return deviceTypeSmartDisplay
	case "smart speaker":
		return deviceTypeSmartSpeaker
	case "peripheral":
		return deviceTypePeripheral
	case "wearable":
		return deviceTypeWearable
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
