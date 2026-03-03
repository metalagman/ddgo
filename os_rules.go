package ddgo

import (
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

type osYAMLVersionRule struct {
	Regex   string `yaml:"regex"`
	Version string `yaml:"version"`
}

type osYAMLRule struct {
	Regex    string              `yaml:"regex"`
	Name     string              `yaml:"name"`
	Version  string              `yaml:"version"`
	Versions []osYAMLVersionRule `yaml:"versions"`
}

type osVersionRule struct {
	pattern         *regexp2.Regexp
	versionTemplate string
}

type osRule struct {
	pattern         *regexp2.Regexp
	nameTemplate    string
	versionTemplate string
	versions        []osVersionRule
}

func parseOSSnapshot(runtime *parserRuntime, ua string, uaRunes []rune) (OS, bool, error) {
	for _, rule := range runtime.osRules {
		match, ok, matchErr := matchRegexp2Runes(rule.pattern, uaRunes)
		if matchErr != nil {
			return OS{}, false, fmt.Errorf("match os rule: %w", matchErr)
		}
		if !ok {
			continue
		}

		osInfo, err := buildOSFromRule(rule, ua, uaRunes, match)
		if err != nil {
			return OS{}, false, err
		}
		return osInfo, true, nil
	}

	return OS{}, false, nil
}

func buildOSFromRule(rule osRule, ua string, uaRunes []rune, match *regexp2.Match) (OS, error) {
	name := normalizeRuleField(expandRuleTemplate(rule.nameTemplate, match))
	version := normalizeRuleVersion(expandRuleTemplate(rule.versionTemplate, match))
	if version == Unknown {
		resolvedVersion, err := resolveOSNestedVersion(rule.versions, uaRunes)
		if err != nil {
			return OS{}, err
		}
		if resolvedVersion != "" {
			version = normalizeRuleVersion(resolvedVersion)
		}
	}

	return OS{
		Name:     name,
		Version:  version,
		Platform: detectOSPlatform(ua, name),
	}, nil
}

func resolveOSNestedVersion(versionRules []osVersionRule, uaRunes []rune) (string, error) {
	for _, nested := range versionRules {
		nestedMatch, nestedOK, nestedErr := matchRegexp2Runes(nested.pattern, uaRunes)
		if nestedErr != nil {
			return "", fmt.Errorf("match os version rule: %w", nestedErr)
		}
		if !nestedOK {
			continue
		}

		version := expandRuleTemplate(nested.versionTemplate, nestedMatch)
		if normalizeRuleVersion(version) != Unknown {
			return version, nil
		}
	}
	return "", nil
}

func loadOSRules(files map[string]string) ([]osRule, error) {
	content, ok := files["oss.yml"]
	if !ok {
		return nil, missingSnapshotFileError("oss.yml")
	}

	var yamlRules []osYAMLRule
	err := yaml.Unmarshal([]byte(content), &yamlRules)
	if err != nil {
		return nil, fmt.Errorf("decode oss.yml: %w", err)
	}

	compiled := make([]osRule, 0, len(yamlRules))
	for _, item := range yamlRules {
		if strings.TrimSpace(item.Regex) == "" {
			continue
		}

		re, compileErr := compileRuleRegex(item.Regex)
		if compileErr != nil {
			return nil, fmt.Errorf("compile oss.yml regex %q: %w", item.Regex, compileErr)
		}

		versionRules, versionErr := compileOSVersionRules(item.Versions)
		if versionErr != nil {
			return nil, versionErr
		}

		compiled = append(compiled, osRule{
			pattern:         re,
			nameTemplate:    item.Name,
			versionTemplate: item.Version,
			versions:        versionRules,
		})
	}
	return compiled, nil
}

func compileOSVersionRules(rawRules []osYAMLVersionRule) ([]osVersionRule, error) {
	versionRules := make([]osVersionRule, 0, len(rawRules))
	for _, nested := range rawRules {
		if strings.TrimSpace(nested.Regex) == "" {
			continue
		}
		nestedRegex, nestedErr := compileRuleRegex(nested.Regex)
		if nestedErr != nil {
			return nil, fmt.Errorf("compile oss.yml nested regex %q: %w", nested.Regex, nestedErr)
		}
		versionRules = append(versionRules, osVersionRule{
			pattern:         nestedRegex,
			versionTemplate: nested.Version,
		})
	}
	return versionRules, nil
}

func detectOSPlatform(ua, osName string) string {
	lowerUA := strings.ToLower(ua)
	if platform := platformFromUserAgent(lowerUA); platform != "" {
		return platform
	}

	lowerOS := strings.ToLower(osName)
	switch {
	case strings.Contains(lowerOS, "android"),
		strings.Contains(lowerOS, "ios"),
		strings.Contains(lowerOS, "ipad"),
		strings.Contains(lowerOS, "watchos"),
		strings.Contains(lowerOS, "harmony"):
		return "ARM"
	case strings.Contains(lowerOS, "windows"):
		return windowsPlatform(ua)
	default:
		return Unknown
	}
}

func platformFromUserAgent(lowerUA string) string {
	switch {
	case hasARMMarker(lowerUA):
		return "ARM"
	case hasX64Marker(lowerUA):
		return "x64"
	case hasX86Marker(lowerUA):
		return "x86"
	default:
		return ""
	}
}

func hasARMMarker(lowerUA string) bool {
	return strings.Contains(lowerUA, "arm64") ||
		strings.Contains(lowerUA, "aarch64") ||
		strings.Contains(lowerUA, "armv") ||
		strings.Contains(lowerUA, " arm;") ||
		strings.Contains(lowerUA, "; arm") ||
		strings.Contains(lowerUA, "arm_64")
}

func hasX64Marker(lowerUA string) bool {
	return strings.Contains(lowerUA, "x86_64") ||
		strings.Contains(lowerUA, "amd64") ||
		strings.Contains(lowerUA, "win64") ||
		strings.Contains(lowerUA, "wow64") ||
		strings.Contains(lowerUA, " x64") ||
		strings.Contains(lowerUA, ";x64")
}

func hasX86Marker(lowerUA string) bool {
	return strings.Contains(lowerUA, "i686") ||
		strings.Contains(lowerUA, "i386") ||
		strings.Contains(lowerUA, " x86")
}
