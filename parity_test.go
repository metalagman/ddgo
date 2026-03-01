package ddgo

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

const (
	parityCasesPath  = "testdata/parity/cases.json"
	parityGoldenPath = "testdata/parity/golden.json"
)

type parityCase struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`
	UserAgent string            `json:"user_agent"`
	Headers   map[string]string `json:"headers,omitempty"`
}

func TestParityGoldenFixtures(t *testing.T) {
	t.Parallel()

	cases := loadParityCases(t)
	actual := runParityCases(t, cases)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		writeGolden(t, actual)
		return
	}

	expected := loadGolden(t)

	caseIDs := make([]string, 0, len(cases))
	for _, tc := range cases {
		caseIDs = append(caseIDs, tc.ID)
		expectedResult, ok := expected[tc.ID]
		if !ok {
			t.Fatalf("missing golden fixture for case %q", tc.ID)
		}
		actualResult := actual[tc.ID]
		if reflect.DeepEqual(expectedResult, actualResult) {
			continue
		}
		t.Fatalf("parity mismatch for case %q (%s)\nexpected:\n%s\nactual:\n%s", tc.ID, tc.Source, prettyJSON(t, expectedResult), prettyJSON(t, actualResult))
	}

	sort.Strings(caseIDs)
	for id := range expected {
		idx := sort.SearchStrings(caseIDs, id)
		if idx < len(caseIDs) && caseIDs[idx] == id {
			continue
		}
		t.Fatalf("golden fixture %q has no matching case in %s", id, parityCasesPath)
	}
}

func TestParityHeaderHintEquivalence(t *testing.T) {
	t.Parallel()

	cases := loadParityCases(t)
	detector := New()
	for _, tc := range cases {
		if len(tc.Headers) == 0 {
			continue
		}
		fromHeaders := detector.ParseWithHeaders(tc.UserAgent, tc.Headers)
		hints := ParseClientHintsFromHeaders(tc.Headers)
		fromHints := detector.ParseWithClientHints(tc.UserAgent, hints)
		if reflect.DeepEqual(fromHeaders, fromHints) {
			continue
		}
		t.Fatalf("header/hints differential mismatch for case %q\nfrom headers:\n%s\nfrom hints:\n%s", tc.ID, prettyJSON(t, fromHeaders), prettyJSON(t, fromHints))
	}
}

func loadParityCases(t *testing.T) []parityCase {
	t.Helper()

	raw, err := os.ReadFile(parityCasesPath)
	if err != nil {
		t.Fatalf("read parity cases: %v", err)
	}
	var cases []parityCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode parity cases: %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("no parity cases found in %s", parityCasesPath)
	}
	for _, tc := range cases {
		if tc.ID == "" {
			t.Fatalf("parity case with empty id in %s", parityCasesPath)
		}
		if tc.Source == "" {
			t.Fatalf("parity case %q is missing source metadata", tc.ID)
		}
	}
	return cases
}

func runParityCases(t *testing.T, cases []parityCase) map[string]Result {
	t.Helper()

	detector := New()
	results := make(map[string]Result, len(cases))
	for _, tc := range cases {
		if tc.ID == "" {
			t.Fatal("unexpected empty fixture id")
		}
		if _, exists := results[tc.ID]; exists {
			t.Fatalf("duplicate parity fixture id %q", tc.ID)
		}
		if len(tc.Headers) == 0 {
			results[tc.ID] = detector.Parse(tc.UserAgent)
			continue
		}
		results[tc.ID] = detector.ParseWithHeaders(tc.UserAgent, tc.Headers)
	}
	return results
}

func loadGolden(t *testing.T) map[string]Result {
	t.Helper()

	raw, err := os.ReadFile(parityGoldenPath)
	if err != nil {
		t.Fatalf("read parity golden file: %v", err)
	}
	var golden map[string]Result
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode parity golden file: %v", err)
	}
	return golden
}

func writeGolden(t *testing.T, golden map[string]Result) {
	t.Helper()

	payload, err := json.MarshalIndent(golden, "", "  ")
	if err != nil {
		t.Fatalf("encode parity golden file: %v", err)
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(parityGoldenPath, payload, 0o644); err != nil {
		t.Fatalf("write parity golden file: %v", err)
	}
}

func prettyJSON(t *testing.T, v any) string {
	t.Helper()

	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("encode pretty json: %v", err)
	}
	return string(payload)
}
