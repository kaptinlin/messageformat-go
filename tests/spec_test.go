package tests

import (
	"testing"

	"github.com/kaptinlin/messageformat-go"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/tests/utils"
)

// skipTags defines tags that should be skipped in tests
var skipTags = map[string]bool{
	"u:locale": true,
}

// tests creates test functions for a given test case (matches TypeScript structure)
func tests(tc utils.Test) func(*testing.T) {
	return func(t *testing.T) {
		// Merge built-in functions with test functions (matches TypeScript: { ...DraftFunctions, ...TestFunctions })
		allFunctions := make(map[string]functions.MessageFunction)

		// Add draft functions
		for name, fn := range functions.DraftFunctions {
			allFunctions[name] = fn
		}

		// Add test functions
		for name, fn := range TestFunctions() {
			allFunctions[name] = fn
		}

		switch utils.GetTestType(tc) {
		case utils.TestTypeSyntaxError:
			t.Run("syntax error", func(t *testing.T) {
				t.Run("MessageFormat(string)", func(t *testing.T) {
					options := &messageformat.MessageFormatOptions{
						Functions: allFunctions,
					}
					_, err := messageformat.New(tc.Locale, tc.Src, options)
					assert.Error(t, err, "Expected syntax error")
				})
			})

		case utils.TestTypeDataModelError:
			t.Run("data model error", func(t *testing.T) {
				t.Run("MessageFormat(string)", func(t *testing.T) {
					options := &messageformat.MessageFormatOptions{
						Functions: allFunctions,
					}
					_, err := messageformat.New(tc.Locale, tc.Src, options)
					assert.Error(t, err, "Expected data model error")
				})
			})

		case utils.TestTypeFormat:
			fallthrough
		default:
			t.Run("format", func(t *testing.T) {
				// Create MessageFormat options
				options := &messageformat.MessageFormatOptions{
					Functions: allFunctions,
				}

				// Set bidi isolation if specified (matches TypeScript: bidiIsolation: tc.bidiIsolation)
				if tc.BidiIsolationRaw != nil {
					if bidiStr, ok := tc.BidiIsolationRaw.(string); ok {
						switch bidiStr {
						case "default":
							options.BidiIsolation = messageformat.BidiDefault
						case "none":
							options.BidiIsolation = messageformat.BidiNone
						}
					} else if tc.BidiIsolation != nil {
						if *tc.BidiIsolation {
							options.BidiIsolation = messageformat.BidiDefault
						} else {
							options.BidiIsolation = messageformat.BidiNone
						}
					}
				}

				// Create MessageFormat instance
				mf, err := messageformat.New(tc.Locale, tc.Src, options)
				if err != nil {
					// If we expect no errors but got one, this is a failure
					if tc.ExpErrors == nil {
						t.Errorf("Unexpected error creating MessageFormat: %v", err)
					}
					return
				}

				// Collect errors using callback (matches TypeScript: let errors: any[] = [])
				var errors []error
				onError := func(err error) {
					errors = append(errors, err)
				}

				// Format the message (matches TypeScript: const msg = mf.format(tc.params, err => errors.push(err)))
				result, err := mf.Format(tc.GetParamsMap(), onError)

				// Check expected result (matches TypeScript: if (typeof tc.exp === 'string') expect(msg).toBe(tc.exp))
				if tc.Exp != nil {
					if expectedStr, ok := tc.Exp.(string); ok {
						assert.Equal(t, expectedStr, result, "Formatted result mismatch")
					}
				}

				// Check expected errors (matches TypeScript error checking logic)
				if tc.ExpErrors != nil {
					if tc.ExpErrors == false {
						assert.Empty(t, errors, "Expected no errors but got: %v", errors)
					} else {
						assert.NotEmpty(t, errors, "Expected errors but got none")
					}
				} else {
					assert.Empty(t, errors, "Unexpected errors: %v", errors)
				}

				// Format method should not return errors for runtime issues
				assert.NoError(t, err, "Format method should not return errors for runtime issues")

				// Check expected parts if specified (matches TypeScript: if (tc.expParts))
				if tc.ExpParts != nil {
					// Reset errors for parts test (matches TypeScript: errors = [])
					errors = nil
					parts, err := mf.FormatToParts(tc.GetParamsMap(), onError)

					// Convert parts to interface{} for comparison (matches TypeScript: expect(mp).toMatchObject(tc.expParts))
					var actualParts []interface{}
					for _, part := range parts {
						partMap := map[string]interface{}{
							"type": part.Type(),
						}

						// Handle different part types according to test expectations
						switch p := part.(type) {
						case *messagevalue.MarkupPart:
							// For markup parts, include kind, name, and options
							partMap["kind"] = p.Kind()
							partMap["name"] = p.Name()

							// Handle options
							options := p.Options()
							if len(options) > 0 {
								// Handle u:id option specially - it should be promoted to top level
								if id, hasID := options["u:id"]; hasID {
									partMap["id"] = id
									// Create filtered options without u:id
									if len(options) > 1 {
										filteredOptions := make(map[string]interface{})
										for k, v := range options {
											if k != "u:id" {
												filteredOptions[k] = v
											}
										}
										if len(filteredOptions) > 0 {
											partMap["options"] = filteredOptions
										}
									}
								} else {
									partMap["options"] = options
								}
							}

						case *messagevalue.TextPart:
							// For text parts, include value
							partMap["value"] = p.Value()

						case *messagevalue.BidiIsolationPart:
							// For bidi isolation parts, include value
							partMap["value"] = p.Value()

						case *messagevalue.FallbackPart:
							// For fallback parts, include source
							partMap["source"] = p.Source()

						case *messagevalue.NumberPart:
							// For number parts, include parts array if available
							if numberParts := p.Parts(); len(numberParts) > 0 {
								var parts []interface{}
								for _, np := range numberParts {
									parts = append(parts, map[string]interface{}{
										"type":  np.Type(),
										"value": np.Value(),
									})
								}
								partMap["parts"] = parts
							} else {
								// Fallback to value if no parts available
								partMap["value"] = p.Value()
							}

						default:
							// For other parts, include value
							partMap["value"] = part.Value()

							// Check if the part has ID, dir, or locale options
							if withOptions, ok := part.(interface {
								GetID() string
								GetDir() string
								GetLocale() string
							}); ok {
								if id := withOptions.GetID(); id != "" {
									partMap["id"] = id
								}
								if dir := withOptions.GetDir(); dir != "" && dir != "auto" {
									partMap["dir"] = dir
								}
								if locale := withOptions.GetLocale(); locale != "" {
									partMap["locale"] = locale
								}
							}
						}

						actualParts = append(actualParts, partMap)
					}

					assert.Equal(t, tc.ExpParts, actualParts, "Parts mismatch")

					// Check errors for parts formatting
					if tc.ExpErrors != nil {
						if tc.ExpErrors == false {
							assert.Empty(t, errors, "Expected no errors in parts but got: %v", errors)
						} else {
							assert.NotEmpty(t, errors, "Expected errors in parts but got none")
						}
					} else {
						assert.Empty(t, errors, "Unexpected errors in parts: %v", errors)
					}

					// FormatToParts method should not return errors for runtime issues
					assert.NoError(t, err, "FormatToParts method should not return errors for runtime issues")
				}
			})
		}
	}
}

// TestMessageFormatWorkingGroup runs all tests from the MessageFormat Working Group test suite
// Matches TypeScript structure: for (const scenario of testScenarios(...))
func TestMessageFormatWorkingGroup(t *testing.T) {
	testDir := "messageformat-wg/test/tests"

	scenarios, err := utils.TestScenarios(testDir)
	require.NoError(t, err, "Failed to load test scenarios")

	for _, scenario := range scenarios {
		t.Run(scenario.Scenario, func(t *testing.T) {
			testCases := utils.TestCases(scenario)

			for _, tc := range testCases {
				// Determine if test should be skipped (matches TypeScript: tc.tags?.some(tag => skipTags.has(tag)))
				shouldSkip := false
				if tc.Tags != nil {
					for _, tag := range tc.Tags {
						if skipTags[tag] {
							shouldSkip = true
							break
						}
					}
				}

				testName := utils.TestName(tc)

				// Handle test execution (matches TypeScript: const describe_ = tc.only ? describe.only : ...)
				switch {
				case tc.Only:
					// Run only this test
					t.Run(testName, tests(tc))
				case shouldSkip:
					// Skip this test
					t.Run(testName, func(t *testing.T) {
						t.Skip("Skipped due to tag")
					})
				default:
					// Run normal test
					t.Run(testName, tests(tc))
				}
			}
		})
	}
}
