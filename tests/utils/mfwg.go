package mfwg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
)

// TestParam represents a parameter in the test
type TestParam struct {
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
	Value any    `json:"value"`
}

// Test represents a single test case from the MessageFormat Working Group test suite
type Test struct {
	Src           string      `json:"src"`
	Locale        string      `json:"locale,omitempty"`
	Params        []TestParam `json:"params,omitempty"`
	Exp           any         `json:"exp,omitempty"`
	ExpCleanSrc   *string     `json:"expCleanSrc,omitempty"`
	ExpErrors     any         `json:"expErrors,omitempty"`
	ExpParts      []any       `json:"expParts,omitempty"`
	BidiIsolation *bool       `json:"bidiIsolation,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	Description   string      `json:"description,omitempty"`
	id            string
}

type rawTest struct {
	Src           *string      `json:"src"`
	Locale        *string      `json:"locale,omitempty"`
	Params        *[]TestParam `json:"params,omitempty"`
	Exp           any          `json:"exp,omitempty"`
	ExpCleanSrc   *string      `json:"expCleanSrc,omitempty"`
	ExpErrors     any          `json:"expErrors,omitempty"`
	ExpParts      *[]any       `json:"expParts,omitempty"`
	BidiIsolation any          `json:"bidiIsolation,omitempty"`
	Only          bool         `json:"only,omitempty"`
	Tags          *[]string    `json:"tags,omitempty"`
	Description   string       `json:"description,omitempty"`
}

// GetParamsMap converts the params array to a map for easier use
func (t Test) GetParamsMap() map[string]any {
	if t.Params == nil {
		return nil
	}

	result := make(map[string]any)
	for _, param := range t.Params {
		result[param.Name] = param.Value
	}
	return result
}

// TestScenario represents a collection of related tests
type TestScenario struct {
	Scenario string `json:"scenario"`
	Path     string `json:"-"`
	Tests    []Test `json:"tests"`
}

// ID returns the fixture-relative identity of the test case.
// TypeScript original code:
// // No direct equivalent; reference test names are not guaranteed unique.
func (t Test) ID() string {
	return t.id
}

// TestType represents the type of test based on expected behavior
type TestType string

const (
	TestTypeSyntaxError    TestType = "syntax-error"
	TestTypeDataModelError TestType = "data-model-error"
	TestTypeFormat         TestType = "format"
)

// TestName generates a descriptive name for a test case
func TestName(tc Test) string {
	if tc.Description != "" {
		return tc.Description
	}

	name := fmt.Sprintf("src: %s", tc.Src)
	if tc.Locale != "" {
		name += fmt.Sprintf(", locale: %s", tc.Locale)
	}
	if len(tc.Params) > 0 {
		name += fmt.Sprintf(", params: %v", tc.Params)
	}

	return name
}

// GetTestType determines the type of test based on expected errors
// TypeScript original code:
// if (!tc.expErrors) return 'valid';
// if (Array.isArray(tc.expErrors)) { ... }
func GetTestType(tc Test) TestType {
	expected := ExpectedErrors(tc)
	for _, err := range expected {
		if err["type"] == "syntax-error" {
			return TestTypeSyntaxError
		}
	}
	for _, err := range expected {
		errorType, _ := err["type"].(string)
		if isDataModelErrorType(errorType) {
			return TestTypeDataModelError
		}
	}
	return TestTypeFormat
}

// isDataModelErrorType reports whether an official error belongs to model validation.
// TypeScript original code:
// const dataModelErrors = ['duplicate-attribute', ...];
func isDataModelErrorType(errorType string) bool {
	switch errorType {
	case "duplicate-attribute",
		"duplicate-declaration",
		"duplicate-option-name",
		"duplicate-variant",
		"missing-fallback-variant",
		"missing-selector-annotation",
		"variant-key-mismatch":
		return true
	default:
		return false
	}
}

type testFile struct {
	Schema                string    `json:"$schema,omitempty"`
	Scenario              string    `json:"scenario"`
	Description           string    `json:"description,omitempty"`
	DefaultTestProperties *rawTest  `json:"defaultTestProperties,omitempty"`
	Tests                 []rawTest `json:"tests"`
}

// TestScenarios loads all test scenarios from a directory
func TestScenarios(testDir string) ([]TestScenario, error) {
	var scenarios []TestScenario

	// Walk through the test directory
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Load the JSON file
		// #nosec G304 - path is safely constructed from filepath.WalkDir
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Parse the JSON content as TestFile
		var testFile testFile
		if err := json.Unmarshal(data, &testFile); err != nil {
			return fmt.Errorf("failed to parse JSON file %s: %w", path, err)
		}
		relativePath, err := filepath.Rel(testDir, path)
		if err != nil {
			return fmt.Errorf("failed to resolve fixture path %s: %w", path, err)
		}
		relativePath = filepath.ToSlash(relativePath)

		tests := make([]Test, len(testFile.Tests))
		for i, raw := range testFile.Tests {
			test, err := normalizeTest(testFile.DefaultTestProperties, raw)
			if err != nil {
				return fmt.Errorf("failed to normalize JSON file %s test %d: %w", path, i, err)
			}
			test.id = fmt.Sprintf("%s#%d", relativePath, i)
			if raw.Only {
				return fmt.Errorf("fixture %s uses only: true", test.id)
			}
			tests[i] = test
		}

		// Create the scenario
		scenarioName := testFile.Scenario
		if scenarioName == "" {
			scenarioName = relativePath
		}
		scenario := TestScenario{
			Scenario: scenarioName,
			Path:     relativePath,
			Tests:    tests,
		}

		scenarios = append(scenarios, scenario)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk test directory: %w", err)
	}

	return scenarios, nil
}

// normalizeTest applies scenario defaults with explicit test properties taking precedence.
// TypeScript original code:
// const td = Object.assign({}, defaults, test);
// if ('type' in p && p.type === 'datetime') pr[p.name] = new Date(p.value);
func normalizeTest(defaults *rawTest, raw rawTest) (Test, error) {
	if defaults == nil {
		defaults = &rawTest{}
	}

	src := defaults.Src
	if raw.Src != nil {
		src = raw.Src
	}
	locale := defaults.Locale
	if raw.Locale != nil {
		locale = raw.Locale
	}
	params := defaults.Params
	if raw.Params != nil {
		params = raw.Params
	}
	tags := defaults.Tags
	if raw.Tags != nil {
		tags = raw.Tags
	}
	exp := defaults.Exp
	if raw.Exp != nil {
		exp = raw.Exp
	}
	expCleanSrc := defaults.ExpCleanSrc
	if raw.ExpCleanSrc != nil {
		expCleanSrc = raw.ExpCleanSrc
	}
	expErrors := defaults.ExpErrors
	if raw.ExpErrors != nil {
		expErrors = raw.ExpErrors
	}
	expParts := defaults.ExpParts
	if raw.ExpParts != nil {
		expParts = raw.ExpParts
	}
	bidiIsolation := defaults.BidiIsolation
	if raw.BidiIsolation != nil {
		bidiIsolation = raw.BidiIsolation
	}

	test := Test{
		Exp:         exp,
		ExpCleanSrc: expCleanSrc,
		ExpErrors:   expErrors,
		Description: raw.Description,
	}
	if src != nil {
		test.Src = *src
	}
	if locale != nil {
		test.Locale = *locale
	}
	if params != nil {
		test.Params = slices.Clone(*params)
		for i := range test.Params {
			param := &test.Params[i]
			if param.Type != "datetime" {
				continue
			}
			value, ok := param.Value.(string)
			if !ok {
				return Test{}, fmt.Errorf("datetime param %q must be a string", param.Name)
			}
			parsed, err := time.Parse(time.RFC3339, value)
			if err != nil {
				parsed, err = time.ParseInLocation("2006-01-02T15:04:05", value, time.UTC)
			}
			if err != nil {
				return Test{}, fmt.Errorf("invalid datetime param %q: %w", param.Name, err)
			}
			param.Value = parsed
		}
	}
	if tags != nil {
		test.Tags = slices.Clone(*tags)
	}
	if expParts != nil {
		test.ExpParts = slices.Clone(*expParts)
	}
	switch value := bidiIsolation.(type) {
	case bool:
		test.BidiIsolation = &value
	case string:
		isolate := value == "default"
		test.BidiIsolation = &isolate
	}
	return test, nil
}

// TestCases extracts all test cases from a scenario
func TestCases(scenario TestScenario) []Test {
	return slices.Clone(scenario.Tests)
}

// ExpectedString returns the expected result as a string if possible
func ExpectedString(tc Test) (string, bool) {
	if str, ok := tc.Exp.(string); ok {
		return str, true
	}
	return "", false
}

// ExpectedErrors returns the expected errors in a normalized format
func ExpectedErrors(tc Test) []map[string]any {
	if tc.ExpErrors == nil {
		return nil
	}

	var errors []map[string]any

	switch v := tc.ExpErrors.(type) {
	case []any:
		for _, err := range v {
			if errMap, ok := err.(map[string]any); ok {
				errors = append(errors, errMap)
			}
		}
	case map[string]any:
		errors = append(errors, v)
	case bool:
		if v {
			// If expErrors is true, we expect some error but don't know the details
			errors = append(errors, map[string]any{"type": "unknown"})
		}
	}

	return errors
}

// ExpectedParts returns the expected parts if available
func ExpectedParts(tc Test) []any {
	return tc.ExpParts
}

// GetBidiIsolation returns the bidi isolation setting, defaulting to true if not specified
func GetBidiIsolation(tc Test) bool {
	if tc.BidiIsolation != nil {
		return *tc.BidiIsolation
	}
	return true // Default to true as per MessageFormat spec
}
