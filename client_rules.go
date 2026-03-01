package ddgo

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

type clientYAMLRule struct {
	Regex   string `yaml:"regex"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Engine  struct {
		Default string `yaml:"default"`
	} `yaml:"engine"`
}

type clientRule struct {
	pattern        *regexp2.Regexp
	nameTemplate   string
	versionPattern string
	engineDefault  string
}

type clientRuleSet struct {
	clientType string
	rules      []clientRule
}

type clientEngineYAMLRule struct {
	Regex string `yaml:"regex"`
	Name  string `yaml:"name"`
}

type clientEngineRule struct {
	pattern *regexp2.Regexp
	name    string
}

var (
	clientRulesOnce sync.Once
	clientRules     []clientRuleSet
	clientRulesErr  error

	clientEngineOnce sync.Once
	clientEngines    []clientEngineRule
	clientEngineErr  error
)

var clientRuleSources = []struct {
	path       string
	clientType string
}{
	{path: "client/feed_readers.yml", clientType: "Feed Reader"},
	{path: "client/mobile_apps.yml", clientType: "Mobile App"},
	{path: "client/mediaplayers.yml", clientType: "Media Player"},
	{path: "client/pim.yml", clientType: "PIM"},
	{path: "client/browsers.yml", clientType: "Browser"},
	{path: "client/libraries.yml", clientType: "Library"},
}

var (
	reEngineBlinkPrimary = regexp.MustCompile(`\b(?:Chrome|Chromium|HeadlessChrome|CriOS|CrMo)/([0-9.]+)\b`)
	reEngineBlinkAlt     = regexp.MustCompile(`\b(?:OPR|Edg|EdgA|EdgiOS|YaBrowser|SamsungBrowser|Brave|Vivaldi)/([0-9.]+)\b`)
	reEngineWebKit       = regexp.MustCompile(`\bAppleWebKit/([0-9.]+)\b`)
	reEngineGecko        = regexp.MustCompile(`\brv:([0-9.]+)\b`)
	reEngineTrident      = regexp.MustCompile(`\bTrident/([0-9.]+)\b`)
	reEnginePresto       = regexp.MustCompile(`\bPresto/([0-9.]+)\b`)
	reEngineNetFront     = regexp.MustCompile(`\bNetFront/([0-9.]+)\b`)
	reEngineGoanna       = regexp.MustCompile(`\bGoanna/([0-9.]+)\b`)
	reEngineServo        = regexp.MustCompile(`\bServo/([0-9.]+)\b`)
	reEngineEkiohFlow    = regexp.MustCompile(`\bEkioh(?:Flow)?/([0-9.]+)\b`)
	reEngineArachne      = regexp.MustCompile(`\bxChaos_Arachne/([0-9.]+)\b`)
)

func parseClientSnapshot(ua string) (Client, bool) {
	ruleSets, err := loadClientRules()
	if err != nil {
		panic("ddgo: client rules not initialized: " + err.Error())
	}

	for _, set := range ruleSets {
		for _, rule := range set.rules {
			match, ok := matchRegexp2String(rule.pattern, ua)
			if !ok {
				continue
			}

			name := normalizeRuleField(expandRuleTemplate(rule.nameTemplate, match))
			version := normalizeRuleVersion(expandRuleTemplate(rule.versionPattern, match))

			engine := strings.TrimSpace(rule.engineDefault)
			if engine == "" {
				engine, err = detectClientEngine(ua)
				if err != nil {
					panic("ddgo: client engine rules not initialized: " + err.Error())
				}
			}
			engine = normalizeRuleField(engine)

			engineVersion := Unknown
			if engine != Unknown {
				engineVersion = normalizeRuleVersion(extractEngineVersion(ua, engine, version))
			}

			return Client{
				Type:          set.clientType,
				Name:          name,
				Version:       version,
				Engine:        engine,
				EngineVersion: engineVersion,
			}, true
		}
	}

	return Client{}, false
}

func loadClientRules() ([]clientRuleSet, error) {
	clientRulesOnce.Do(func() {
		files, err := loadSnapshotFiles()
		if err != nil {
			clientRulesErr = err
			return
		}

		compiled := make([]clientRuleSet, 0, len(clientRuleSources))
		for _, source := range clientRuleSources {
			content, ok := files[source.path]
			if !ok {
				clientRulesErr = missingSnapshotFileError(source.path)
				return
			}

			var yamlRules []clientYAMLRule
			if err := yaml.Unmarshal([]byte(content), &yamlRules); err != nil {
				clientRulesErr = fmt.Errorf("decode %s: %w", source.path, err)
				return
			}

			rules := make([]clientRule, 0, len(yamlRules))
			for _, item := range yamlRules {
				if strings.TrimSpace(item.Regex) == "" {
					continue
				}
				re, err := compileRuleRegex(item.Regex)
				if err != nil {
					clientRulesErr = fmt.Errorf("compile %s regex %q: %w", source.path, item.Regex, err)
					return
				}
				rules = append(rules, clientRule{
					pattern:        re,
					nameTemplate:   item.Name,
					versionPattern: item.Version,
					engineDefault:  item.Engine.Default,
				})
			}

			compiled = append(compiled, clientRuleSet{
				clientType: source.clientType,
				rules:      rules,
			})
		}

		clientRules = compiled
	})

	if clientRulesErr != nil {
		return nil, clientRulesErr
	}
	return clientRules, nil
}

func loadClientEngineRules() ([]clientEngineRule, error) {
	clientEngineOnce.Do(func() {
		files, err := loadSnapshotFiles()
		if err != nil {
			clientEngineErr = err
			return
		}
		content, ok := files["client/browser_engine.yml"]
		if !ok {
			clientEngineErr = missingSnapshotFileError("client/browser_engine.yml")
			return
		}

		var yamlRules []clientEngineYAMLRule
		if err := yaml.Unmarshal([]byte(content), &yamlRules); err != nil {
			clientEngineErr = fmt.Errorf("decode client/browser_engine.yml: %w", err)
			return
		}

		compiled := make([]clientEngineRule, 0, len(yamlRules))
		for _, item := range yamlRules {
			if strings.TrimSpace(item.Regex) == "" || strings.TrimSpace(item.Name) == "" {
				continue
			}
			re, err := compileRuleRegex(item.Regex)
			if err != nil {
				clientEngineErr = fmt.Errorf("compile client/browser_engine.yml regex %q: %w", item.Regex, err)
				return
			}
			compiled = append(compiled, clientEngineRule{
				pattern: re,
				name:    item.Name,
			})
		}
		clientEngines = compiled
	})

	if clientEngineErr != nil {
		return nil, clientEngineErr
	}
	return clientEngines, nil
}

func detectClientEngine(ua string) (string, error) {
	rules, err := loadClientEngineRules()
	if err != nil {
		return Unknown, err
	}
	for _, rule := range rules {
		if _, ok := matchRegexp2String(rule.pattern, ua); ok {
			return rule.name, nil
		}
	}

	switch {
	case reClientEdge.MatchString(ua), reClientEdgeAlt.MatchString(ua), reClientOpera.MatchString(ua), reClientChrome.MatchString(ua), reClientChromeOS.MatchString(ua):
		return "Blink", nil
	case reWebKit.MatchString(ua):
		return "WebKit", nil
	case reGeckoRV.MatchString(ua), strings.Contains(ua, "Gecko/"):
		return "Gecko", nil
	case strings.Contains(ua, "Trident/"):
		return "Trident", nil
	case strings.Contains(ua, "Presto/"):
		return "Presto", nil
	default:
		return Unknown, nil
	}
}

func extractEngineVersion(ua, engineName, clientVersion string) string {
	switch strings.ToLower(engineName) {
	case "blink":
		if version := firstMatch(reEngineBlinkPrimary, ua, ""); version != "" {
			return version
		}
		if version := firstMatch(reEngineBlinkAlt, ua, ""); version != "" {
			return version
		}
		if clientVersion != Unknown {
			return clientVersion
		}
	case "webkit":
		return firstMatch(reEngineWebKit, ua, "")
	case "gecko":
		return firstMatch(reEngineGecko, ua, "")
	case "trident":
		return firstMatch(reEngineTrident, ua, "")
	case "presto":
		return firstMatch(reEnginePresto, ua, "")
	case "netfront":
		return firstMatch(reEngineNetFront, ua, "")
	case "goanna":
		return firstMatch(reEngineGoanna, ua, "")
	case "servo":
		return firstMatch(reEngineServo, ua, "")
	case "ekiohflow":
		return firstMatch(reEngineEkiohFlow, ua, "")
	case "arachne":
		return firstMatch(reEngineArachne, ua, "")
	}
	return ""
}
