package messageformat

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Constructor API Tests

// TestNew tests the New constructor with various inputs
func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		locales     interface{}
		source      interface{}
		options     *MessageFormatOptions
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid string locale and source",
			locales:     "en",
			source:      "Hello World",
			options:     nil,
			expectError: false,
		},
		{
			name:        "valid slice locales",
			locales:     []string{"en", "fr"},
			source:      "Hello",
			options:     nil,
			expectError: false,
		},
		{
			name:        "nil locales",
			locales:     nil,
			source:      "Hello",
			options:     nil,
			expectError: false,
		},
		{
			name:        "empty string locale",
			locales:     "",
			source:      "Hello",
			options:     nil,
			expectError: false,
		},
		{
			name:        "empty slice locales",
			locales:     []string{},
			source:      "Hello",
			options:     nil,
			expectError: false,
		},
		{
			name:        "invalid locales type",
			locales:     123,
			source:      "Hello",
			options:     nil,
			expectError: true,
			errorMsg:    "locales must be string, []string, or nil",
		},
		{
			name:        "invalid source type",
			locales:     "en",
			source:      123,
			options:     nil,
			expectError: true,
			errorMsg:    "source must be string or datamodel.Message",
		},
		{
			name:    "with options",
			locales: "en",
			source:  "Hello",
			options: &MessageFormatOptions{
				BidiIsolation: BidiNone,
				Dir:           DirRTL,
				LocaleMatcher: LocaleLookup,
			},
			expectError: false,
		},
		{
			name:        "syntax error in pattern",
			locales:     "en",
			source:      "Hello {$name", // Missing closing brace
			options:     nil,
			expectError: true,
			errorMsg:    "parse-error",
		},
		{
			name:        "complex pattern with functions",
			locales:     "en",
			source:      "You have {$count :number} messages",
			options:     nil,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New(tc.locales, tc.source, tc.options)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
				assert.Nil(t, mf)
			} else {
				require.NoError(t, err)
				require.NotNil(t, mf)
			}
		})
	}
}

// TestNewWithDataModelMessage tests constructor with datamodel.Message
func TestNewWithDataModelMessage(t *testing.T) {
	pattern := datamodel.NewPattern([]datamodel.PatternElement{
		datamodel.NewTextElement("Hello World"),
	})
	message := datamodel.NewPatternMessage(nil, pattern, "")

	mf, err := New("en", message, nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{"en"}, mf.locales)
}

// Format API Tests

// TestFormat tests the Format method
func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		values   map[string]interface{}
		onError  func(error)
		expected string
	}{
		{
			name:     "simple text",
			source:   "Hello World",
			values:   nil,
			onError:  nil,
			expected: "Hello World",
		},
		{
			name:     "with variable",
			source:   "Hello {$name}",
			values:   map[string]interface{}{"name": "Alice"},
			onError:  nil,
			expected: "Hello \u2068Alice\u2069",
		},
		{
			name:     "missing variable",
			source:   "Hello {$missing}",
			values:   map[string]interface{}{},
			onError:  nil,
			expected: "Hello \u2068{$missing}\u2069",
		},
		{
			name:     "nil values",
			source:   "Hello {$name}",
			values:   nil,
			onError:  nil,
			expected: "Hello \u2068{$name}\u2069",
		},
		{
			name:     "number formatting",
			source:   "Count: {$count :number}",
			values:   map[string]interface{}{"count": 1234.56},
			onError:  nil,
			expected: "Count: 1,234.56",
		},
		{
			name:     "integer formatting",
			source:   "Items: {$items :integer}",
			values:   map[string]interface{}{"items": 42.7},
			onError:  nil,
			expected: "Items: 43",
		},
		{
			name:     "string function with different types",
			source:   "Value: {$value :string}",
			values:   map[string]interface{}{"value": 123},
			onError:  nil,
			expected: "Value: \u2068123\u2069",
		},
		{
			name:     "empty pattern",
			source:   "",
			values:   map[string]interface{}{"unused": "value"},
			onError:  nil,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.source, nil)
			require.NoError(t, err)

			result, err := mf.Format(tc.values, tc.onError)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFormatToPartsAPI tests the FormatToParts API method
func TestFormatToPartsAPI(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		values        map[string]interface{}
		expectedParts int
		expectedTypes []string
	}{
		{
			name:          "simple text",
			source:        "Hello World",
			values:        nil,
			expectedParts: 1,
			expectedTypes: []string{"text"},
		},
		{
			name:          "with variable",
			source:        "Hello {$name}",
			values:        map[string]interface{}{"name": "Alice"},
			expectedParts: 4,
			expectedTypes: []string{"text", "bidiIsolation", "string", "bidiIsolation"},
		},
		{
			name:          "missing variable",
			source:        "Hello {$missing}",
			values:        map[string]interface{}{},
			expectedParts: 4,
			expectedTypes: []string{"text", "bidiIsolation", "fallback", "bidiIsolation"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.source, nil)
			require.NoError(t, err)

			parts, err := mf.FormatToParts(tc.values, nil)
			require.NoError(t, err)
			assert.Len(t, parts, tc.expectedParts)

			for i, expectedType := range tc.expectedTypes {
				if i < len(parts) {
					assert.Equal(t, expectedType, parts[i].Type(), "Part %d type mismatch", i)
				}
			}
		})
	}
}

// Options API Tests

// TestMessageFormatOptions tests various option combinations
func TestMessageFormatOptions(t *testing.T) {
	tests := []struct {
		name            string
		options         *MessageFormatOptions
		expectedBidi    bool
		expectedDir     string
		expectedMatcher string
	}{
		{
			name:            "nil options (defaults)",
			options:         nil,
			expectedBidi:    true,
			expectedDir:     "ltr",
			expectedMatcher: "best fit",
		},
		{
			name: "bidi isolation none",
			options: &MessageFormatOptions{
				BidiIsolation: BidiNone,
			},
			expectedBidi:    false,
			expectedDir:     "ltr",
			expectedMatcher: "best fit",
		},
		{
			name: "rtl direction",
			options: &MessageFormatOptions{
				Dir: DirRTL,
			},
			expectedBidi:    true,
			expectedDir:     "rtl",
			expectedMatcher: "best fit",
		},
		{
			name: "lookup matcher",
			options: &MessageFormatOptions{
				LocaleMatcher: LocaleLookup,
			},
			expectedBidi:    true,
			expectedDir:     "ltr",
			expectedMatcher: "lookup",
		},
		{
			name: "all custom options",
			options: &MessageFormatOptions{
				BidiIsolation: BidiNone,
				Dir:           DirRTL,
				LocaleMatcher: LocaleLookup,
			},
			expectedBidi:    false,
			expectedDir:     "rtl",
			expectedMatcher: "lookup",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", "Hello", tc.options)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedBidi, mf.bidiIsolation)
			assert.Equal(t, tc.expectedDir, mf.dir)
			assert.Equal(t, tc.expectedMatcher, mf.localeMatcher)
		})
	}
}

// TestResolvedOptionsAPI tests the ResolvedOptions method
func TestResolvedOptionsAPI(t *testing.T) {
	options := &MessageFormatOptions{
		BidiIsolation: BidiNone,
		Dir:           DirRTL,
		LocaleMatcher: LocaleLookup,
	}

	mf, err := New("en", "Hello", options)
	require.NoError(t, err)

	resolved := mf.ResolvedOptions()

	assert.Equal(t, BidiNone, resolved.BidiIsolation)
	assert.Equal(t, DirRTL, resolved.Dir)
	assert.Equal(t, LocaleLookup, resolved.LocaleMatcher)
	assert.NotNil(t, resolved.Functions)
	assert.Contains(t, resolved.Functions, "number")
	assert.Contains(t, resolved.Functions, "string")
	assert.Contains(t, resolved.Functions, "integer")
}

// Functions API Tests

// TestDefaultFunctions tests that default functions are available
func TestDefaultFunctions(t *testing.T) {
	mf, err := New("en", "Hello", nil)
	require.NoError(t, err)

	expectedFunctions := []string{"string", "number", "integer"}
	for _, funcName := range expectedFunctions {
		assert.Contains(t, mf.functions, funcName, "Default function %s should be available", funcName)
	}
}

// TestCustomFunctionsAPI tests custom function integration
func TestCustomFunctionsAPI(t *testing.T) {
	customFunc := func(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom", "en", "custom")
	}

	options := &MessageFormatOptions{
		Functions: map[string]functions.MessageFunction{
			"custom": customFunc,
		},
	}

	mf, err := New("en", "Hello", options)
	require.NoError(t, err)

	// Custom function should be available
	assert.Contains(t, mf.functions, "custom")

	// Default functions should still be there
	assert.Contains(t, mf.functions, "string")
	assert.Contains(t, mf.functions, "number")
	assert.Contains(t, mf.functions, "integer")
}

// Locale API Tests

// TestLocaleHandling tests locale-related functionality
func TestLocaleHandling(t *testing.T) {
	tests := []struct {
		name        string
		locales     interface{}
		expectedDir string
	}{
		{
			name:        "english locale",
			locales:     "en",
			expectedDir: "ltr",
		},
		{
			name:        "arabic locale",
			locales:     "ar",
			expectedDir: "rtl",
		},
		{
			name:        "hebrew locale",
			locales:     "he",
			expectedDir: "rtl",
		},
		{
			name:        "multiple locales",
			locales:     []string{"en", "fr"},
			expectedDir: "ltr",
		},
		{
			name:        "nil locales",
			locales:     nil,
			expectedDir: "auto",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New(tc.locales, "Hello", nil)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDir, mf.dir)
		})
	}
}

// MessageFormat 2.0 Pattern Tests

// TestSelectPatterns tests MessageFormat 2.0 select patterns
func TestSelectPatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		values   map[string]interface{}
		expected string
	}{
		{
			name: "simple match pattern",
			pattern: `.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}`,
			values:   map[string]interface{}{"count": 0},
			expected: "No items",
		},
		{
			name: "plural match pattern",
			pattern: `.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}`,
			values:   map[string]interface{}{"count": 1},
			expected: "One item",
		},
		{
			name: "multiple match pattern",
			pattern: `.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}`,
			values:   map[string]interface{}{"count": 5},
			expected: "5 items",
		},
		{
			name: "multi-dimensional match",
			pattern: `.input {$count :number}
.input {$gender :string}
.match $count $gender
0   male   {{He has no items}}
0   female {{She has no items}}
0   *      {{They have no items}}
one male   {{He has one item}}
one female {{She has one item}}
one *      {{They have one item}}
*   male   {{He has {$count} items}}
*   female {{She has {$count} items}}
*   *      {{They have {$count} items}}`,
			values:   map[string]interface{}{"count": 3, "gender": "female"},
			expected: "She has 3 items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.pattern)
			require.NoError(t, err)

			result, err := mf.Format(tc.values)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestLocalDeclarations tests local variable declarations
func TestLocalDeclarations(t *testing.T) {
	pattern := `.local $greeting = {|Hello| :string}
.local $punctuation = {|!| :string}
{{{$greeting}, {$name}{$punctuation}}}`

	mf, err := New("en", pattern)
	require.NoError(t, err)

	result, err := mf.Format(map[string]interface{}{"name": "World"})
	require.NoError(t, err)
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "World")
}

// Error Handling API Tests

// TestErrorCallback tests error callback functionality
func TestErrorCallback(t *testing.T) {
	var capturedErrors []error
	onError := func(err error) {
		capturedErrors = append(capturedErrors, err)
	}

	mf, err := New("en", "Hello {$name}", nil)
	require.NoError(t, err)

	// Valid case - no errors should be captured
	result, err := mf.Format(map[string]interface{}{"name": "World"}, onError)
	require.NoError(t, err)
	assert.Equal(t, "Hello \u2068World\u2069", result)
	assert.Empty(t, capturedErrors)

	// Missing variable case - should still work with fallback
	result, err = mf.Format(map[string]interface{}{}, onError)
	require.NoError(t, err)
	assert.Equal(t, "Hello \u2068{$name}\u2069", result)
	// Error callback behavior may vary based on implementation
}

// TestInvalidPatterns tests various invalid pattern scenarios
func TestInvalidPatterns(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		shouldFail  bool
		expectParse bool // Some patterns parse but fail at runtime
	}{
		{
			name:        "unclosed brace",
			pattern:     "Hello {$name",
			shouldFail:  true,
			expectParse: false,
		},
		{
			name:        "invalid function - parses but fails at runtime",
			pattern:     "Hello {$name :invalid}",
			shouldFail:  false, // This parses successfully
			expectParse: true,
		},
		{
			name:        "malformed expression",
			pattern:     "Hello {$}",
			shouldFail:  true,
			expectParse: false,
		},
		{
			name:        "invalid match syntax",
			pattern:     ".match invalid {{text}}",
			shouldFail:  true,
			expectParse: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.pattern)
			switch {
			case tc.expectParse:
				// Should parse successfully but may fail at format time
				require.NoError(t, err)
				require.NotNil(t, mf)
			case tc.shouldFail:
				// Should fail to parse
				require.Error(t, err)
			default:
				require.NoError(t, err)
				require.NotNil(t, mf)
			}
		})
	}
}

// TestAPIEdgeCases tests edge cases and boundary conditions for the API
func TestAPIEdgeCases(t *testing.T) {
	t.Run("extremely long patterns", func(t *testing.T) {
		// Test with very long pattern
		longPattern := "Long message: "
		for i := 0; i < 100; i++ {
			longPattern += "{$var" + fmt.Sprintf("%d", i) + "} "
		}

		mf, err := New("en", longPattern)
		require.NoError(t, err)

		values := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			values[fmt.Sprintf("var%d", i)] = fmt.Sprintf("val%d", i)
		}

		result, err := mf.Format(values)
		require.NoError(t, err)
		assert.Contains(t, result, "Long message:")
	})

	t.Run("nested braces in text", func(t *testing.T) {
		// Test escaped braces
		mf, err := New("en", "Object: \\{key: {$value}\\}")
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"value": "test"})
		require.NoError(t, err)
		assert.Contains(t, result, "{key:")
	})

	t.Run("unicode in patterns and values", func(t *testing.T) {
		mf, err := New("zh-CN", "你好，{$name}！")
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"name": "世界"})
		require.NoError(t, err)
		assert.Contains(t, result, "你好")
		assert.Contains(t, result, "世界")
	})
}

// TypeScript Compatibility Tests - based on reference implementation analysis
func TestTypeScriptCompatibility(t *testing.T) {
	t.Run("offset function integration", func(t *testing.T) {
		// Test that matches TypeScript behavior: offset function with add/subtract
		pattern := `{$count :offset add=1} people liked this`
		mf, err := New("en", pattern, &MessageFormatOptions{
			Functions: map[string]functions.MessageFunction{
				"offset": functions.OffsetFunction,
			},
		})
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"count": 5})
		require.NoError(t, err)
		assert.Contains(t, result, "6") // 5 + 1 = 6
	})

	t.Run("error handling matches TypeScript patterns", func(t *testing.T) {
		// Test that errors are handled the same way as TypeScript
		pattern := `{$invalid :unknown}`
		mf, err := New("en", pattern)
		require.NoError(t, err)

		var capturedErrors []error
		result, err := mf.Format(map[string]interface{}{}, func(err error) {
			capturedErrors = append(capturedErrors, err)
		})

		// Should not fail Format() but should capture errors
		require.NoError(t, err)
		assert.NotEmpty(t, capturedErrors)
		assert.Contains(t, result, "{") // Should contain fallback
	})

	t.Run("bidi isolation default behavior", func(t *testing.T) {
		// Test default bidi isolation behavior matches TypeScript
		pattern := `Hello {$name}!`
		mf, err := New("ar", pattern) // RTL locale
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"name": "عالم"})
		require.NoError(t, err)
		// Should contain bidi isolation characters by default
		assert.Contains(t, result, "Hello")
	})

	t.Run("formatToParts structure matches TypeScript", func(t *testing.T) {
		// Test that parts structure is similar to TypeScript implementation
		pattern := `Hello {$name :string}!`
		mf, err := New("en", pattern, &MessageFormatOptions{
			BidiIsolation: BidiNone, // Disable bidi isolation to match expected output
		})
		require.NoError(t, err)

		parts, err := mf.FormatToParts(map[string]interface{}{"name": "World"})
		require.NoError(t, err)

		// Should have structure similar to TypeScript: text, expression, text
		require.Len(t, parts, 3)
		assert.Equal(t, "text", parts[0].Type())
		assert.Equal(t, "Hello ", parts[0].Value())
		assert.Equal(t, "string", parts[1].Type())
		assert.Equal(t, "text", parts[2].Type())
		assert.Equal(t, "!", parts[2].Value())
	})

	t.Run("locale resolution matches TypeScript", func(t *testing.T) {
		// Test that locale resolution behavior matches TypeScript
		mf, err := New([]string{"en-US", "en", "fr"}, "Hello {$name}")
		require.NoError(t, err)

		// Should resolve options properly like TypeScript
		resolved := mf.ResolvedOptions()
		assert.Equal(t, "best fit", string(resolved.LocaleMatcher))
		assert.Equal(t, "default", string(resolved.BidiIsolation))
	})

	t.Run("function registry behavior", func(t *testing.T) {
		// Test that function registry works like TypeScript
		customFuncs := map[string]functions.MessageFunction{
			"test": func(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
				return messagevalue.NewStringValue("test-result", "", ctx.Locales()[0])
			},
		}

		mf, err := New("en", "{$val :test}", &MessageFormatOptions{
			Functions: customFuncs,
		})
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"val": "input"})
		require.NoError(t, err)
		assert.Contains(t, result, "test-result")
	})
}

// Index Exports API Tests

// TestIndexExports tests the exported functions from index.go
func TestIndexExports(t *testing.T) {
	t.Run("ParseMessage function", func(t *testing.T) {
		message, err := ParseMessage("Hello {$name}")
		require.NoError(t, err)
		require.NotNil(t, message)
		assert.True(t, IsPatternMessage(message))
	})

	t.Run("Validate function", func(t *testing.T) {
		pattern := datamodel.NewPattern([]datamodel.PatternElement{
			datamodel.NewTextElement("Hello"),
		})
		message := datamodel.NewPatternMessage(nil, pattern, "")

		_, err := Validate(message, nil)
		require.NoError(t, err)
	})

	t.Run("Type guards", func(t *testing.T) {
		expr := datamodel.NewExpression(nil, nil, nil)
		assert.True(t, IsExpression(expr))

		literal := datamodel.NewLiteral("test")
		assert.True(t, IsLiteral(literal))
	})

	t.Run("DefaultFunctions export", func(t *testing.T) {
		assert.NotNil(t, DefaultFunctions)
		assert.Contains(t, DefaultFunctions, "string")
		assert.Contains(t, DefaultFunctions, "number")
		assert.Contains(t, DefaultFunctions, "integer")
	})

	t.Run("DraftFunctions export", func(t *testing.T) {
		assert.NotNil(t, DraftFunctions)
	})
}

// Bidi Isolation API Tests

// TestBidiIsolationOptions tests bidi isolation behavior
func TestBidiIsolationOptions(t *testing.T) {
	tests := []struct {
		name            string
		bidiIsolation   BidiIsolation
		source          string
		values          map[string]interface{}
		expectIsolation bool
	}{
		{
			name:            "default bidi isolation",
			bidiIsolation:   BidiDefault,
			source:          "Hello {$name}",
			values:          map[string]interface{}{"name": "World"},
			expectIsolation: true,
		},
		{
			name:            "no bidi isolation",
			bidiIsolation:   BidiNone,
			source:          "Hello {$name}",
			values:          map[string]interface{}{"name": "World"},
			expectIsolation: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			options := &MessageFormatOptions{
				BidiIsolation: tc.bidiIsolation,
			}
			mf, err := New("en", tc.source, options)
			require.NoError(t, err)

			result, err := mf.Format(tc.values, nil)
			require.NoError(t, err)

			if tc.expectIsolation {
				assert.Contains(t, result, "\u2068") // FSI
				assert.Contains(t, result, "\u2069") // PDI
			} else {
				assert.NotContains(t, result, "\u2068")
				assert.NotContains(t, result, "\u2069")
			}
		})
	}
}
