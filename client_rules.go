package ddgo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

type clientYAMLRule struct {
	Regex   string `yaml:"regex"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Engine  struct {
		Default  string            `yaml:"default"`
		Versions map[string]string `yaml:"versions"`
	} `yaml:"engine"`
}

type clientRule struct {
	pattern        *regexp2.Regexp
	nameTemplate   string
	versionPattern string
	engineDefault  string
	engineVersions map[string]string
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

type clientRuleSource struct {
	path       string
	clientType string
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

func parseClientSnapshot(runtime *parserRuntime, ua string, uaRunes []rune) (Client, bool, error) {
	for _, set := range runtime.clientRules {
		client, ok, err := parseClientFromRuleSet(set, ua, uaRunes, runtime)
		if err != nil {
			return Client{}, false, err
		}
		if ok {
			return client, true, nil
		}
	}

	return Client{}, false, nil
}

func parseClientFromRuleSet(set clientRuleSet, ua string, uaRunes []rune, runtime *parserRuntime) (Client, bool, error) {
	for _, rule := range set.rules {
		match, ok, matchErr := matchRegexp2Runes(rule.pattern, uaRunes)
		if matchErr != nil {
			return Client{}, false, fmt.Errorf("match client rule: %w", matchErr)
		}
		if !ok {
			continue
		}

		client, buildErr := buildClientFromMatch(set.clientType, rule, ua, uaRunes, match, runtime)
		if buildErr != nil {
			return Client{}, false, buildErr
		}
		return client, true, nil
	}
	return Client{}, false, nil
}

func buildClientFromMatch(
	clientType string,
	rule clientRule,
	ua string,
	uaRunes []rune,
	match *regexp2.Match,
	runtime *parserRuntime,
) (Client, error) {
	name := normalizeRuleField(expandRuleTemplate(rule.nameTemplate, match))
	version := normalizeRuleVersion(expandRuleTemplate(rule.versionPattern, match))

	engine := rule.engineDefault
	if len(rule.engineVersions) > 0 && version != Unknown {
		bestV := ""
		for v, e := range rule.engineVersions {
			if compareVersions(version, v) >= 0 {
				if bestV == "" || compareVersions(v, bestV) > 0 {
					bestV = v
					engine = e
				}
			}
		}
	}

	var err error
	if engine == "" {
		engine, err = detectClientEngine(ua, uaRunes, runtime.clientEngines)
		if err != nil {
			return Client{}, fmt.Errorf("detect client engine: %w", err)
		}
	}
	engine = normalizeRuleField(engine)

	engineVersion := Unknown
	if engine != Unknown {
		engineVersion = normalizeRuleVersion(extractEngineVersion(ua, engine, version))
	}

	return Client{
		Type:          clientType,
		Name:          name,
		Version:       version,
		Engine:        engine,
		EngineVersion: engineVersion,
	}, nil
}

func loadClientRules(files map[string]string) ([]clientRuleSet, error) {
	sources := clientRuleSources()
	compiled := make([]clientRuleSet, 0, len(sources))

	for _, source := range sources {
		content, ok := files[source.path]
		if !ok {
			return nil, missingSnapshotFileError(source.path)
		}

		rules, err := decodeClientRules(content, source.path)
		if err != nil {
			return nil, err
		}

		compiled = append(compiled, clientRuleSet{
			clientType: source.clientType,
			rules:      rules,
		})
	}

	return compiled, nil
}

func decodeClientRules(content, sourcePath string) ([]clientRule, error) {
	var yamlRules []clientYAMLRule
	err := yaml.Unmarshal([]byte(content), &yamlRules)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", sourcePath, err)
	}

	compiled := make([]clientRule, 0, len(yamlRules))
	for _, item := range yamlRules {
		if strings.TrimSpace(item.Regex) == "" {
			continue
		}
		re, compileErr := compileRuleRegex(item.Regex)
		if compileErr != nil {
			return nil, fmt.Errorf("compile %s regex %q: %w", sourcePath, item.Regex, compileErr)
		}
		compiled = append(compiled, clientRule{
			pattern:        re,
			nameTemplate:   item.Name,
			versionPattern: item.Version,
			engineDefault:  item.Engine.Default,
			engineVersions: item.Engine.Versions,
		})
	}

	return compiled, nil
}

func clientRuleSources() []clientRuleSource {
	return []clientRuleSource{
		{path: "client/feed_readers.yml", clientType: "Feed Reader"},
		{path: "client/mobile_apps.yml", clientType: "Mobile App"},
		{path: "client/mediaplayers.yml", clientType: "Media Player"},
		{path: "client/pim.yml", clientType: "PIM"},
		{path: "client/browsers.yml", clientType: "Browser"},
		{path: "client/libraries.yml", clientType: "Library"},
	}
}

func loadClientEngineRules(files map[string]string) ([]clientEngineRule, error) {
	const enginePath = "client/browser_engine.yml"

	content, ok := files[enginePath]
	if !ok {
		return nil, missingSnapshotFileError(enginePath)
	}

	var yamlRules []clientEngineYAMLRule
	err := yaml.Unmarshal([]byte(content), &yamlRules)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", enginePath, err)
	}

	compiled := make([]clientEngineRule, 0, len(yamlRules))
	for _, item := range yamlRules {
		if strings.TrimSpace(item.Regex) == "" || strings.TrimSpace(item.Name) == "" {
			continue
		}
		re, compileErr := compileRuleRegex(item.Regex)
		if compileErr != nil {
			return nil, fmt.Errorf("compile %s regex %q: %w", enginePath, item.Regex, compileErr)
		}
		compiled = append(compiled, clientEngineRule{
			pattern: re,
			name:    item.Name,
		})
	}
	return compiled, nil
}

func detectClientEngine(ua string, uaRunes []rune, rules []clientEngineRule) (string, error) {
	for _, rule := range rules {
		ok, matchErr := matchRegexp2RunesBool(rule.pattern, uaRunes)
		if matchErr != nil {
			return Unknown, fmt.Errorf("match client engine rule: %w", matchErr)
		}
		if ok {
			return rule.name, nil
		}
	}

	switch {
	case reClientEdge.MatchString(ua),
		reClientEdgeAlt.MatchString(ua),
		reClientOpera.MatchString(ua),
		reClientChrome.MatchString(ua),
		reClientChromeOS.MatchString(ua):
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
		version := firstMatch(reEngineBlinkPrimary, ua, "")
		if version != "" {
			return version
		}
		version = firstMatch(reEngineBlinkAlt, ua, "")
		if version != "" {
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

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) || i < len(parts2); i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}
	return 0
}
