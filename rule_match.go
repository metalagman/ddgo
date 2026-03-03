package ddgo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

var (
	reTemplateGroup         = regexp.MustCompile(`\$(\d+)`)
	errCompiledSnapshotMiss = errors.New("compiled snapshot missing file")
)

const (
	regexpMatchTimeout       = 100 * time.Millisecond
	templateSubmatchExpected = 2
)

func compileRuleRegex(pattern string) (*regexp2.Regexp, error) {
	re, err := regexp2.Compile(normalizeRulePattern(pattern), 0)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = regexpMatchTimeout
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
		if len(raw) != templateSubmatchExpected {
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

func matchRegexp2Runes(pattern *regexp2.Regexp, input []rune) (*regexp2.Match, bool, error) {
	if pattern == nil {
		return nil, false, nil
	}
	match, err := pattern.FindRunesMatch(input)
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
	return fmt.Errorf("%w: %s", errCompiledSnapshotMiss, path)
}
