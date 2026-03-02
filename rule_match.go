package ddgo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

var reTemplateGroup = regexp.MustCompile(`\$(\d+)`)

func compileRuleRegex(pattern string) (*regexp2.Regexp, error) {
	re, err := regexp2.Compile(normalizeRulePattern(pattern), 0)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = 100 * time.Millisecond
	return re, nil
}

func expandRuleTemplate(template string, match *regexp2.Match) string {
	template = strings.TrimSpace(template)
	if template == "" {
		return ""
	}
	if match == nil {
		return template
	}
	expanded := reTemplateGroup.ReplaceAllStringFunc(template, func(token string) string {
		raw := reTemplateGroup.FindStringSubmatch(token)
		if len(raw) != 2 {
			return ""
		}
		groupIndex, err := strconv.Atoi(raw[1])
		if err != nil {
			return ""
		}
		if match.GroupCount() < groupIndex+1 {
			return ""
		}
		group := match.GroupByNumber(groupIndex)
		if group == nil {
			return ""
		}
		return group.String()
	})
	return strings.TrimSpace(expanded)
}

func normalizeRuleVersion(version string) string {
	version = strings.TrimSpace(strings.ReplaceAll(version, "_", "."))
	version = strings.Trim(version, " .")
	if version == "" {
		return Unknown
	}
	return version
}

func normalizeRuleField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return Unknown
	}
	return value
}

func matchRegexp2String(pattern *regexp2.Regexp, input string) (*regexp2.Match, bool, error) {
	if pattern == nil {
		return nil, false, nil
	}
	match, err := pattern.FindStringMatch(input)
	if err != nil {
		if isRegexp2MatchTimeout(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if match == nil {
		return nil, false, nil
	}
	return match, true, nil
}

func isRegexp2MatchTimeout(err error) bool {
	return err != nil && strings.Contains(err.Error(), "match timeout after")
}

func missingSnapshotFileError(path string) error {
	return fmt.Errorf("compiled snapshot missing %s", path)
}
