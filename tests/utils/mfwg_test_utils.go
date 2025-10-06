package utils

import (
	"fmt"
	"github.com/go-json-experiment/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TestParam represents a parameter in the test
type TestParam struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Test represents a single test case from the MessageFormat Working Group test suite
type Test struct {
	Src           string        `json:"src"`
	Locale        string        `json:"locale,omitempty"`
	Params        []TestParam   `json:"params,omitempty"`
	Exp           interface{}   `json:"exp,omitempty"`
	ExpErrors     interface{}   `json:"expErrors,omitempty"`
	ExpParts      []interface{} `json:"expParts,omitempty"`
	BidiIsolation *bool         `json:"bidiIsolation,omitempty"`
	Only          bool          `json:"only,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
	Description   string        `json:"description,omitempty"`

	// Internal field to store the original bidi isolation value
	BidiIsolationRaw interface{} `json:"-"`
}

// GetParamsMap converts the params array to a map for easier use
func (t Test) GetParamsMap() map[string]interface{} {
	if t.Params == nil {
		return nil
	}

	result := make(map[string]interface{})
	for _, param := range t.Params {
		result[param.Name] = param.Value
	}
	return result
}

// TestScenario represents a collection of related tests
type TestScenario struct {
	Scenario string `json:"scenario"`
	Tests    []Test `json:"tests"`
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
func GetTestType(tc Test) TestType {
	if tc.ExpErrors != nil {
		// Check if it's specifically a syntax error or data model error
		if errSlice, ok := tc.ExpErrors.([]interface{}); ok {
			for _, err := range errSlice {
				if errMap, ok := err.(map[string]interface{}); ok {
					if errType, exists := errMap["type"]; exists {
						if errType == "syntax-error" {
							return TestTypeSyntaxError
						}
						if errType == "data-model-error" {
							return TestTypeDataModelError
						}
						// Runtime errors like bad-operand, bad-option should be format tests
						if errType == "bad-operand" || errType == "bad-option" ||
							errType == "bad-selector" || errType == "unresolved-variable" ||
							errType == "unsupported-operation" || errType == "bad-function-result" {
							return TestTypeFormat
						}
					}
				}
			}
		} else if errMap, ok := tc.ExpErrors.(map[string]interface{}); ok {
			if errType, exists := errMap["type"]; exists {
				if errType == "syntax-error" {
					return TestTypeSyntaxError
				}
				if errType == "data-model-error" {
					return TestTypeDataModelError
				}
				// Runtime errors like bad-operand, bad-option should be format tests
				if errType == "bad-operand" || errType == "bad-option" ||
					errType == "bad-selector" || errType == "unresolved-variable" ||
					errType == "unsupported-operation" || errType == "bad-function-result" {
					return TestTypeFormat
				}
			}
		} else if tc.ExpErrors == true {
			// If expErrors is just true, assume it's a format test with runtime errors
			return TestTypeFormat
		} else if tc.ExpErrors == false {
			// If expErrors is false, it's a normal format test
			return TestTypeFormat
		}

		// If there are expected errors but no specific type, assume format tests with runtime errors
		return TestTypeFormat
	}

	return TestTypeFormat
}

// TestFile represents the structure of a test JSON file
type TestFile struct {
	Schema                string                 `json:"$schema,omitempty"`
	Scenario              string                 `json:"scenario"`
	Description           string                 `json:"description,omitempty"`
	DefaultTestProperties map[string]interface{} `json:"defaultTestProperties,omitempty"`
	Tests                 []Test                 `json:"tests"`
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
		var testFile TestFile
		if err := json.Unmarshal(data, &testFile); err != nil {
			return fmt.Errorf("failed to parse JSON file %s: %w", path, err)
		}

		// Apply default test properties to each test
		for i := range testFile.Tests {
			test := &testFile.Tests[i]

			// Apply defaults if not already set
			if testFile.DefaultTestProperties != nil {
				if test.Locale == "" {
					if locale, ok := testFile.DefaultTestProperties["locale"].(string); ok {
						test.Locale = locale
					}
				}
				if test.BidiIsolation == nil {
					if bidiValue, exists := testFile.DefaultTestProperties["bidiIsolation"]; exists {
						test.BidiIsolationRaw = bidiValue
						if bidi, ok := bidiValue.(bool); ok {
							test.BidiIsolation = &bidi
						} else if bidiStr, ok := bidiValue.(string); ok {
							// Convert string to bool for compatibility
							boolValue := bidiStr == "default"
							test.BidiIsolation = &boolValue
						}
					}
				}
				if test.ExpErrors == nil {
					if expErrors, ok := testFile.DefaultTestProperties["expErrors"]; ok {
						test.ExpErrors = expErrors
					}
				}
			}
		}

		// Create the scenario
		scenario := TestScenario{
			Scenario: testFile.Scenario,
			Tests:    testFile.Tests,
		}

		scenarios = append(scenarios, scenario)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk test directory: %w", err)
	}

	return scenarios, nil
}

// TestCases extracts all test cases from a scenario
func TestCases(scenario TestScenario) []Test {
	return scenario.Tests
}

// HasTag checks if a test has a specific tag
func HasTag(tc Test, tag string) bool {
	for _, t := range tc.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// ShouldSkip determines if a test should be skipped based on tags
func ShouldSkip(tc Test, skipTags map[string]bool) bool {
	for _, tag := range tc.Tags {
		if skipTags[tag] {
			return true
		}
	}
	return false
}

// IsOnlyTest checks if this is an "only" test that should run exclusively
func IsOnlyTest(tc Test) bool {
	return tc.Only
}

// HasOnlyTests checks if any test in the scenario is marked as "only"
func HasOnlyTests(scenario TestScenario) bool {
	for _, test := range scenario.Tests {
		if test.Only {
			return true
		}
	}
	return false
}

// FilterOnlyTests returns only the tests marked with "only"
func FilterOnlyTests(scenario TestScenario) []Test {
	var onlyTests []Test
	for _, test := range scenario.Tests {
		if test.Only {
			onlyTests = append(onlyTests, test)
		}
	}
	return onlyTests
}

// ExpectedString returns the expected result as a string if possible
func ExpectedString(tc Test) (string, bool) {
	if str, ok := tc.Exp.(string); ok {
		return str, true
	}
	return "", false
}

// ExpectedErrors returns the expected errors in a normalized format
func ExpectedErrors(tc Test) []map[string]interface{} {
	if tc.ExpErrors == nil {
		return nil
	}

	var errors []map[string]interface{}

	switch v := tc.ExpErrors.(type) {
	case []interface{}:
		for _, err := range v {
			if errMap, ok := err.(map[string]interface{}); ok {
				errors = append(errors, errMap)
			}
		}
	case map[string]interface{}:
		errors = append(errors, v)
	case bool:
		if v {
			// If expErrors is true, we expect some error but don't know the details
			errors = append(errors, map[string]interface{}{"type": "unknown"})
		}
	}

	return errors
}

// ExpectedParts returns the expected parts if available
func ExpectedParts(tc Test) []interface{} {
	return tc.ExpParts
}

// GetBidiIsolation returns the bidi isolation setting, defaulting to true if not specified
func GetBidiIsolation(tc Test) bool {
	if tc.BidiIsolation != nil {
		return *tc.BidiIsolation
	}
	return true // Default to true as per MessageFormat spec
}
