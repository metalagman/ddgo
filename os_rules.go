package ddgo

import (
	"fmt"
	"strings"
	"sync"

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

var (
	osRulesOnce sync.Once
	osRules     []osRule
	osRulesErr  error
)

func parseOSSnapshot(ua string) (OS, bool) {
	rules, err := loadOSRules()
	if err != nil {
		panic("ddgo: os rules not initialized: " + err.Error())
	}

	for _, rule := range rules {
		match, ok := matchRegexp2String(rule.pattern, ua)
		if !ok {
			continue
		}

		name := normalizeRuleField(expandRuleTemplate(rule.nameTemplate, match))
		version := normalizeRuleVersion(expandRuleTemplate(rule.versionTemplate, match))
		if version == Unknown && len(rule.versions) > 0 {
			for _, nested := range rule.versions {
				nestedMatch, nestedOK := matchRegexp2String(nested.pattern, ua)
				if !nestedOK {
					continue
				}
				version = normalizeRuleVersion(expandRuleTemplate(nested.versionTemplate, nestedMatch))
				if version != Unknown {
					break
				}
			}
		}

		return OS{
			Name:     name,
			Version:  version,
			Platform: detectOSPlatform(ua, name),
		}, true
	}

	return OS{}, false
}

func loadOSRules() ([]osRule, error) {
	osRulesOnce.Do(func() {
		files, err := loadSnapshotFiles()
		if err != nil {
			osRulesErr = err
			return
		}
		content, ok := files["oss.yml"]
		if !ok {
			osRulesErr = missingSnapshotFileError("oss.yml")
			return
		}

		var yamlRules []osYAMLRule
		if err := yaml.Unmarshal([]byte(content), &yamlRules); err != nil {
			osRulesErr = fmt.Errorf("decode oss.yml: %w", err)
			return
		}

		compiled := make([]osRule, 0, len(yamlRules))
		for _, item := range yamlRules {
			if strings.TrimSpace(item.Regex) == "" {
				continue
			}
			re, err := compileRuleRegex(item.Regex)
			if err != nil {
				osRulesErr = fmt.Errorf("compile oss.yml regex %q: %w", item.Regex, err)
				return
			}

			versionRules := make([]osVersionRule, 0, len(item.Versions))
			for _, nested := range item.Versions {
				if strings.TrimSpace(nested.Regex) == "" {
					continue
				}
				nestedRegex, nestedErr := compileRuleRegex(nested.Regex)
				if nestedErr != nil {
					osRulesErr = fmt.Errorf("compile oss.yml nested regex %q: %w", nested.Regex, nestedErr)
					return
				}
				versionRules = append(versionRules, osVersionRule{
					pattern:         nestedRegex,
					versionTemplate: nested.Version,
				})
			}

			compiled = append(compiled, osRule{
				pattern:         re,
				nameTemplate:    item.Name,
				versionTemplate: item.Version,
				versions:        versionRules,
			})
		}
		osRules = compiled
	})

	if osRulesErr != nil {
		return nil, osRulesErr
	}
	return osRules, nil
}

func detectOSPlatform(ua, osName string) string {
	lowerUA := strings.ToLower(ua)

	switch {
	case strings.Contains(lowerUA, "arm64"), strings.Contains(lowerUA, "aarch64"), strings.Contains(lowerUA, "armv"), strings.Contains(lowerUA, " arm;"), strings.Contains(lowerUA, "; arm"), strings.Contains(lowerUA, "arm_64"):
		return "ARM"
	case strings.Contains(lowerUA, "x86_64"), strings.Contains(lowerUA, "amd64"), strings.Contains(lowerUA, "win64"), strings.Contains(lowerUA, "wow64"), strings.Contains(lowerUA, " x64"), strings.Contains(lowerUA, ";x64"):
		return "x64"
	case strings.Contains(lowerUA, "i686"), strings.Contains(lowerUA, "i386"), strings.Contains(lowerUA, " x86"):
		return "x86"
	}

	lowerOS := strings.ToLower(osName)
	switch {
	case strings.Contains(lowerOS, "android"), strings.Contains(lowerOS, "ios"), strings.Contains(lowerOS, "ipad"), strings.Contains(lowerOS, "watchos"), strings.Contains(lowerOS, "harmony"):
		return "ARM"
	case strings.Contains(lowerOS, "windows"):
		return windowsPlatform(ua)
	}

	return Unknown
}
