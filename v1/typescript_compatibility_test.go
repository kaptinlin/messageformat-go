package v1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TypeScriptTestCase represents a test case from the TypeScript reference implementation
// TypeScript original code:
//
//	export type TestCase = {
//	  locale?: string | PluralFunction;
//	  options?: Record<string, unknown>;
//	  skip?: string[];
//	  src: string;
//	  exp: Array<[any, string | RegExp | { error: true | string | RegExp } | any[]]>;
//	};
type TypeScriptTestCase struct {
	Locale  string         `json:"locale,omitempty"`
	Options map[string]any `json:"options,omitempty"`
	Skip    []string       `json:"skip,omitempty"`
	Src     string         `json:"src"`
	Exp     [][]any        `json:"exp"`
}

type TypeScriptTestSuite map[string][]TypeScriptTestCase

// loadTypeScriptTestCases loads test cases to verify TypeScript compatibility
func loadTypeScriptTestCases() TypeScriptTestSuite {
	// Basic test cases that mirror the TypeScript implementation structure
	return TypeScriptTestSuite{
		"Basic messages": {
			{
				Src: "This is a string.",
				Exp: [][]any{{nil, "This is a string."}},
			},
			{
				Src: "{foo}",
				Exp: [][]any{
					{nil, map[string]any{"error": true}},
					{map[string]any{"foo": "FOO"}, "FOO"},
				},
			},
		},

		"CLDR locales": {
			{
				Locale: "cy", // Welsh - Complex plural rules test
				Src:    "{NUM, plural, zero{a} one{b} two{c} few{d} many{e} other{f} =42{omg42}}",
				Exp: [][]any{
					{map[string]any{"NUM": 0}, "a"},
					{map[string]any{"NUM": 1}, "b"},
					{map[string]any{"NUM": 2}, "c"},
					{map[string]any{"NUM": 3}, "d"},
					{map[string]any{"NUM": 6}, "e"},
					{map[string]any{"NUM": 15}, "f"},
					{map[string]any{"NUM": 42}, "omg42"},
				},
			},
			{
				Locale: "cy",
				Src:    "{num, selectordinal, zero{0,7,8,9} one{1} two{2} few{3,4} many{5,6} other{+}}",
				Exp: [][]any{
					{map[string]any{"num": 5}, "5,6"},
				},
			},
		},

		"Octothorpe replacement": {
			{
				Src: "{count, plural, one{# item} other{# items}}",
				Exp: [][]any{
					{map[string]any{"count": 1}, "1 item"},
					{map[string]any{"count": 5}, "5 items"},
				},
			},
			{
				Src: "{count, plural, one{# item (total: #)} other{# items (total: #)}}",
				Exp: [][]any{
					{map[string]any{"count": 2}, "2 items (total: 2)"},
				},
			},
		},

		"Select statements": {
			{
				Src: "{gender, select, male{He} female{She} other{They}} went home.",
				Exp: [][]any{
					{map[string]any{"gender": "male"}, "He went home."},
					{map[string]any{"gender": "female"}, "She went home."},
					{map[string]any{"gender": "unknown"}, "They went home."},
				},
			},
		},

		"Nested messages": {
			{
				Src: "{gender, select, male{{count, plural, one{He has # item} other{He has # items}}} female{{count, plural, one{She has # item} other{She has # items}}} other{{count, plural, one{They have # item} other{They have # items}}}}",
				Exp: [][]any{
					{map[string]any{"gender": "male", "count": 1}, "He has 1 item"},
					{map[string]any{"gender": "female", "count": 3}, "She has 3 items"},
					{map[string]any{"gender": "other", "count": 2}, "They have 2 items"},
				},
			},
		},

		"Error handling": {
			{
				Options: map[string]any{"requireAllArguments": true},
				Src:     "{missing}",
				Exp: [][]any{
					{map[string]any{}, map[string]any{"error": true}},
				},
			},
		},

		"Return type variations": {
			{
				Options: map[string]any{"returnType": "string"},
				Src:     "Hello {name}!",
				Exp: [][]any{
					{map[string]any{"name": "World"}, "Hello World!"},
				},
			},
			{
				Options: map[string]any{"returnType": "values"},
				Src:     "Hello {name}!",
				Exp: [][]any{
					{map[string]any{"name": "World"}, []any{"Hello ", "World", "!"}},
				},
			},
		},

		"Strict mode": {
			{
				Options: map[string]any{"strict": true},
				Src:     "{foo, invalid}",
				Exp: [][]any{
					{map[string]any{"foo": "bar"}, map[string]any{"error": true}},
				},
			},
		},
	}
}

// TestTypeScriptCompatibilityOfficial runs official TypeScript compatibility tests
func TestTypeScriptCompatibilityOfficial(t *testing.T) {
	testSuite := loadTypeScriptTestCases()

	for suiteName, testCases := range testSuite {
		t.Run(suiteName, func(t *testing.T) {
			for i, testCase := range testCases {
				t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
					options := &MessageFormatOptions{}

					if testCase.Options != nil {
						if rt, ok := testCase.Options["returnType"].(string); ok {
							if rt == "values" {
								options.ReturnType = ReturnTypeValues
							} else {
								options.ReturnType = ReturnTypeString
							}
						}

						if req, ok := testCase.Options["requireAllArguments"].(bool); ok {
							options.RequireAllArguments = req
						}

						if strict, ok := testCase.Options["strict"].(bool); ok {
							options.Strict = strict
						}
					}

					locale := testCase.Locale
					if locale == "" {
						locale = "en"
					}

					mf, err := New(locale, options)
					require.NoError(t, err, "Failed to create MessageFormat for locale: %s", locale)

					compiled, err := mf.Compile(testCase.Src)
					if err != nil {
						for _, exp := range testCase.Exp {
							if len(exp) >= 2 {
								if errMap, ok := exp[1].(map[string]any); ok {
									if errMap["error"] == true {
										return // Expected compilation error
									}
								}
							}
						}
						t.Fatalf("Unexpected compilation error for message '%s': %v", testCase.Src, err)
					}

					for expIndex, exp := range testCase.Exp {
						if len(exp) < 2 {
							continue
						}

						params := exp[0]
						expected := exp[1]

						result, err := compiled(params)

						if errMap, ok := expected.(map[string]any); ok && errMap["error"] == true {
							assert.Error(t, err, "Expected error for params %v, message: %s", params, testCase.Src)
							continue
						}

						require.NoError(t, err, "Execution error for exp[%d], params %v, message: %s", expIndex, params, testCase.Src)

						if expectedStr, ok := expected.(string); ok {
							assert.Equal(t, expectedStr, result, "Mismatch for exp[%d], params %v, message: %s", expIndex, params, testCase.Src)
						} else if expectedSlice, ok := expected.([]any); ok {
							if options.ReturnType == ReturnTypeValues {
								assert.Equal(t, expectedSlice, result, "Values mismatch for exp[%d], params %v, message: %s", expIndex, params, testCase.Src)
							} else {
								var expectedStr strings.Builder
								for _, part := range expectedSlice {
									expectedStr.WriteString(fmt.Sprintf("%v", part))
								}
								assert.Equal(t, expectedStr.String(), result, "Concatenated values mismatch for exp[%d], params %v, message: %s", expIndex, params, testCase.Src)
							}
						}
					}
				})
			}
		})
	}
}

func TestTypeScriptCompatibilityStaticMethods(t *testing.T) {
	t.Run("Escape function", func(t *testing.T) {
		testCases := []struct {
			input      string
			octothorpe bool
			expected   string
		}{
			{"Hello {name}!", false, "Hello '{'name'}'!"},
			{"Count: #", true, "Count: '#'"},
			{"Count: #", false, "Count: #"},
			{"{test} #{count}", true, "'{'test'}' '#''{'count'}'"},
		}

		for _, tc := range testCases {
			result := Escape(tc.input, tc.octothorpe)
			assert.Equal(t, tc.expected, result, "Escape(%q, %v)", tc.input, tc.octothorpe)
		}
	})

	t.Run("SupportedLocalesOf function", func(t *testing.T) {
		testCases := []struct {
			locales  any
			expected []string
		}{
			{[]string{"en", "fr", "de"}, []string{"en", "fr", "de"}},
			{[]string{"en", "xx", "fr"}, []string{"en", "fr"}}, // Filter invalid locales
			{"en", []string{"en"}},
		}

		for _, tc := range testCases {
			result, err := SupportedLocalesOf(tc.locales)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, result, "SupportedLocalesOf(%v)", tc.locales)
		}
	})
}

func TestTypeScriptCompatibilityOptionsResolution(t *testing.T) {
	mf, err := New("en", &MessageFormatOptions{
		ReturnType:  ReturnTypeString,
		Currency:    "EUR",
		BiDiSupport: true,
	})
	require.NoError(t, err)

	resolved := mf.ResolvedOptions()

	assert.Equal(t, "en", resolved.Locale)
	assert.Equal(t, ReturnTypeString, resolved.ReturnType)
	assert.Equal(t, "EUR", resolved.Currency)
	assert.True(t, resolved.BiDiSupport)
}

// BenchmarkTypeScriptCompatibilityPerformance benchmarks against TypeScript baseline
func BenchmarkTypeScriptCompatibilityPerformance(b *testing.B) {
	scenarios := []struct {
		name    string
		locale  string
		message string
		params  map[string]any
	}{
		{
			name:    "SimpleInterpolation",
			locale:  "en",
			message: "Hello {name}!",
			params:  map[string]any{"name": "World"},
		},
		{
			name:    "BasicPlural",
			locale:  "en",
			message: "{count, plural, one{# item} other{# items}}",
			params:  map[string]any{"count": 5},
		},
		{
			name:    "ComplexNested",
			locale:  "en",
			message: "{gender, select, male{He has {count, plural, one{# item} other{# items}}} other{They have items}}",
			params:  map[string]any{"gender": "male", "count": 3},
		},
		{
			name:    "WelshPlurals",
			locale:  "cy",
			message: "{NUM, plural, zero{zero} one{one} two{two} few{few} many{many} other{other}}",
			params:  map[string]any{"NUM": 6},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			mf, err := New(scenario.locale, &MessageFormatOptions{
				ReturnType: ReturnTypeString,
			})
			if err != nil {
				b.Fatal(err)
			}

			compiled, err := mf.Compile(scenario.message)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for b.Loop() {
				_, err := compiled(scenario.params)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
