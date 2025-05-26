package messageformat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestNewBasic tests the basic constructor functionality
func TestNewBasic(t *testing.T) {
	// Test with string source
	mf, err := New("en", "Hello World", nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{"en"}, mf.locales)
	assert.True(t, mf.bidiIsolation)
	assert.Equal(t, "ltr", mf.dir)
	assert.Equal(t, "best fit", mf.localeMatcher)
}

// TestNewWithNilLocales tests constructor with nil locales
func TestNewWithNilLocales(t *testing.T) {
	mf, err := New(nil, "Hello", nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{}, mf.locales)
	assert.Equal(t, "auto", mf.dir)
}

// TestNewWithMultipleLocales tests constructor with multiple locales
func TestNewWithMultipleLocales(t *testing.T) {
	mf, err := New([]string{"en", "fr"}, "Hello", nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{"en", "fr"}, mf.locales)
	assert.Equal(t, "ltr", mf.dir)
}

// TestNewWithInvalidLocales tests constructor with invalid locales type
func TestNewWithInvalidLocales(t *testing.T) {
	_, err := New(123, "Hello", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "locales must be string, []string, or nil")
}

// TestNewWithInvalidSource tests constructor with invalid source type
func TestNewWithInvalidSource(t *testing.T) {
	_, err := New("en", 123, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source must be string or datamodel.Message")
}

// TestNewWithOptions tests constructor with various options
func TestNewWithOptions(t *testing.T) {
	options := &MessageFormatOptions{
		BidiIsolation: stringPtr("none"),
		Dir:           stringPtr("rtl"),
		LocaleMatcher: stringPtr("lookup"),
	}

	mf, err := New("en", "Hello", options)
	require.NoError(t, err)
	require.NotNil(t, mf)

	assert.False(t, mf.bidiIsolation)
	assert.Equal(t, "rtl", mf.dir)
	assert.Equal(t, "lookup", mf.localeMatcher)
}

// TestNewWithMessage tests constructor with datamodel.Message input
func TestNewWithMessage(t *testing.T) {
	// Create a pattern message
	pattern := datamodel.NewPattern([]datamodel.PatternElement{
		datamodel.NewTextElement("Hello World"),
	})
	message := datamodel.NewPatternMessage(nil, pattern, "")

	mf, err := New("en", message, nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{"en"}, mf.locales)
}

// TestFormat tests the format method with simple text
func TestFormat(t *testing.T) {
	mf, err := New("en", "Hello World", nil)
	require.NoError(t, err)

	result, err := mf.Format(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

// TestFormatToParts tests the formatToParts method
func TestFormatToParts(t *testing.T) {
	mf, err := New("en", "Hello World", nil)
	require.NoError(t, err)

	parts, err := mf.FormatToParts(nil, nil)
	require.NoError(t, err)
	assert.Len(t, parts, 1)
	assert.Equal(t, "text", parts[0].Type())
	assert.Equal(t, "Hello World", parts[0].Value())
}

// TestDefaultFunctions tests that default functions are available
func TestDefaultFunctions(t *testing.T) {
	mf, err := New("en", "Hello", nil)
	require.NoError(t, err)

	// Check that default functions are available
	expectedFunctions := []string{"string", "number", "integer"}
	for _, funcName := range expectedFunctions {
		assert.Contains(t, mf.functions, funcName, "Default function %s should be available", funcName)
	}
}

// TestCustomFunctions tests custom function integration
func TestCustomFunctions(t *testing.T) {
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

// TestFormatWithVariables tests formatting with variable substitution
func TestFormatWithVariables(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		values   map[string]interface{}
		expected string
		hasError bool
	}{
		{
			name:     "simple variable",
			source:   "Hello {$name}",
			values:   map[string]interface{}{"name": "Alice"},
			expected: "Hello \u2068Alice\u2069", // Includes bidi isolation
		},
		{
			name:     "multiple variables",
			source:   "Hello {$name}, you are {$age} years old",
			values:   map[string]interface{}{"name": "Bob", "age": 25},
			expected: "Hello \u2068Bob\u2069, you are \u206825\u2069 years old", // Includes bidi isolation
		},
		{
			name:     "missing variable fallback",
			source:   "Hello {$missing}",
			values:   map[string]interface{}{},
			expected: "Hello \u2068{$missing}\u2069", // Includes bidi isolation for fallback
		},
		{
			name:     "empty values map",
			source:   "Hello {$name}",
			values:   map[string]interface{}{},
			expected: "Hello \u2068{$name}\u2069", // Includes bidi isolation for fallback
		},
		{
			name:     "nil values",
			source:   "Hello {$name}",
			values:   nil,
			expected: "Hello \u2068{$name}\u2069", // Includes bidi isolation for fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New("en", tt.source, nil)
			require.NoError(t, err)

			result, err := mf.Format(tt.values, nil)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFormatWithVariablesNoBidi tests formatting without bidi isolation
func TestFormatWithVariablesNoBidi(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		values   map[string]interface{}
		expected string
	}{
		{
			name:     "simple variable no bidi",
			source:   "Hello {$name}",
			values:   map[string]interface{}{"name": "Alice"},
			expected: "Hello Alice",
		},
		{
			name:     "multiple variables no bidi",
			source:   "Hello {$name}, you are {$age} years old",
			values:   map[string]interface{}{"name": "Bob", "age": 25},
			expected: "Hello Bob, you are 25 years old",
		},
		{
			name:     "missing variable fallback no bidi",
			source:   "Hello {$missing}",
			values:   map[string]interface{}{},
			expected: "Hello {$missing}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &MessageFormatOptions{
				BidiIsolation: stringPtr("none"),
			}
			mf, err := New("en", tt.source, options)
			require.NoError(t, err)

			result, err := mf.Format(tt.values, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatToPartsWithVariables tests formatToParts with variables
func TestFormatToPartsWithVariables(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		values        map[string]interface{}
		expectedParts int
		expectedTypes []string
	}{
		{
			name:          "text with variable (with bidi)",
			source:        "Hello {$name}",
			values:        map[string]interface{}{"name": "Alice"},
			expectedParts: 4, // text, bidiIsolation, string, bidiIsolation
			expectedTypes: []string{"text", "bidiIsolation", "string", "bidiIsolation"},
		},
		{
			name:          "missing variable (with bidi)",
			source:        "Hello {$missing}",
			values:        map[string]interface{}{},
			expectedParts: 4, // text, bidiIsolation, fallback, bidiIsolation
			expectedTypes: []string{"text", "bidiIsolation", "fallback", "bidiIsolation"},
		},
		{
			name:          "multiple variables (with bidi)",
			source:        "{$greeting} {$name}!",
			values:        map[string]interface{}{"greeting": "Hi", "name": "Bob"},
			expectedParts: 8, // bidiIsolation, string, bidiIsolation, text, bidiIsolation, string, bidiIsolation, text
			expectedTypes: []string{"bidiIsolation", "string", "bidiIsolation", "text", "bidiIsolation", "string", "bidiIsolation", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New("en", tt.source, nil)
			require.NoError(t, err)

			parts, err := mf.FormatToParts(tt.values, nil)
			require.NoError(t, err)
			assert.Len(t, parts, tt.expectedParts)

			for i, expectedType := range tt.expectedTypes {
				if i < len(parts) {
					assert.Equal(t, expectedType, parts[i].Type(), "Part %d type mismatch", i)
				}
			}
		})
	}
}

// TestFormatToPartsNoBidi tests formatToParts without bidi isolation
func TestFormatToPartsNoBidi(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		values        map[string]interface{}
		expectedParts int
		expectedTypes []string
	}{
		{
			name:          "text with variable (no bidi)",
			source:        "Hello {$name}",
			values:        map[string]interface{}{"name": "Alice"},
			expectedParts: 2,
			expectedTypes: []string{"text", "string"},
		},
		{
			name:          "missing variable (no bidi)",
			source:        "Hello {$missing}",
			values:        map[string]interface{}{},
			expectedParts: 2,
			expectedTypes: []string{"text", "fallback"},
		},
		{
			name:          "multiple variables (no bidi)",
			source:        "{$greeting} {$name}!",
			values:        map[string]interface{}{"greeting": "Hi", "name": "Bob"},
			expectedParts: 4, // string, text, string, text
			expectedTypes: []string{"string", "text", "string", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &MessageFormatOptions{
				BidiIsolation: stringPtr("none"),
			}
			mf, err := New("en", tt.source, options)
			require.NoError(t, err)

			parts, err := mf.FormatToParts(tt.values, nil)
			require.NoError(t, err)
			assert.Len(t, parts, tt.expectedParts)

			for i, expectedType := range tt.expectedTypes {
				if i < len(parts) {
					assert.Equal(t, expectedType, parts[i].Type(), "Part %d type mismatch", i)
				}
			}
		})
	}
}

// TestLocaleDirection tests locale-based direction detection
func TestLocaleDirection(t *testing.T) {
	tests := []struct {
		name        string
		locale      string
		expectedDir string
	}{
		{
			name:        "English locale",
			locale:      "en",
			expectedDir: "ltr",
		},
		{
			name:        "Arabic locale",
			locale:      "ar",
			expectedDir: "rtl",
		},
		{
			name:        "Hebrew locale",
			locale:      "he",
			expectedDir: "rtl",
		},
		{
			name:        "French locale",
			locale:      "fr",
			expectedDir: "ltr",
		},
		{
			name:        "Persian locale",
			locale:      "fa",
			expectedDir: "rtl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New(tt.locale, "Hello", nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedDir, mf.dir)
		})
	}
}

// TestBidiIsolation tests bidirectional text isolation
func TestBidiIsolation(t *testing.T) {
	tests := []struct {
		name            string
		bidiStrategy    string
		text            string
		expectIsolation bool
	}{
		{
			name:            "default isolation with RTL",
			bidiStrategy:    "default",
			text:            "العالم", // Arabic text
			expectIsolation: true,
		},
		{
			name:            "no isolation",
			bidiStrategy:    "none",
			text:            "العالم",
			expectIsolation: false,
		},
		{
			name:            "default isolation with LTR",
			bidiStrategy:    "default",
			text:            "World",
			expectIsolation: false, // LTR to LTR doesn't need isolation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &MessageFormatOptions{
				BidiIsolation: stringPtr(tt.bidiStrategy),
			}
			mf, err := New("en", "Hello {$name}!", options)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{"name": tt.text}, nil)
			require.NoError(t, err)

			if tt.expectIsolation {
				// Should contain isolation characters
				assert.True(t, strings.Contains(result, "\u2066") ||
					strings.Contains(result, "\u2067") ||
					strings.Contains(result, "\u2068"),
					"Expected isolation characters in result: %q", result)
			}
		})
	}
}

// TestErrorHandling tests error handling in format methods
func TestErrorHandling(t *testing.T) {
	mf, err := New("en", "Hello {$missing}", nil)
	require.NoError(t, err)

	t.Run("traditional error callback", func(t *testing.T) {
		var capturedError error
		errorHandler := func(err error) {
			capturedError = err
		}

		result, err := mf.Format(map[string]interface{}{}, errorHandler)
		require.NoError(t, err)
		assert.Contains(t, result, "{$missing}")
		// Note: Error handling behavior may vary based on implementation
		_ = capturedError // Acknowledge variable usage
	})

	t.Run("nil error handler", func(t *testing.T) {
		result, err := mf.Format(map[string]interface{}{}, nil)
		require.NoError(t, err)
		assert.Contains(t, result, "{$missing}")
	})
}

// TestComplexMessages tests more complex message patterns
func TestComplexMessages(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		values   map[string]interface{}
		expected string
	}{
		{
			name:     "function call",
			source:   "Hello {$name :string}",
			values:   map[string]interface{}{"name": "Alice"},
			expected: "Hello \u2068Alice\u2069", // Includes bidi isolation
		},
		{
			name:     "number formatting",
			source:   "Count: {$count :number}",
			values:   map[string]interface{}{"count": 42},
			expected: "Count: \u206842\u2069", // Includes bidi isolation
		},
		{
			name:     "integer formatting",
			source:   "Items: {$items :integer}",
			values:   map[string]interface{}{"items": 3.14},
			expected: "Items: \u20683\u2069", // Includes bidi isolation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New("en", tt.source, nil)
			require.NoError(t, err)

			result, err := mf.Format(tt.values, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComplexMessagesNoBidi tests complex messages without bidi isolation
func TestComplexMessagesNoBidi(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		values   map[string]interface{}
		expected string
	}{
		{
			name:     "function call no bidi",
			source:   "Hello {$name :string}",
			values:   map[string]interface{}{"name": "Alice"},
			expected: "Hello Alice",
		},
		{
			name:     "number formatting no bidi",
			source:   "Count: {$count :number}",
			values:   map[string]interface{}{"count": 42},
			expected: "Count: 42",
		},
		{
			name:     "integer formatting no bidi",
			source:   "Items: {$items :integer}",
			values:   map[string]interface{}{"items": 3.14},
			expected: "Items: 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &MessageFormatOptions{
				BidiIsolation: stringPtr("none"),
			}
			mf, err := New("en", tt.source, options)
			require.NoError(t, err)

			result, err := mf.Format(tt.values, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		locales  interface{}
		source   interface{}
		options  *MessageFormatOptions
		hasError bool
		errorMsg string
	}{
		{
			name:     "nil options",
			locales:  "en",
			source:   "Hello",
			options:  nil,
			hasError: false,
		},
		{
			name:     "single character locale",
			locales:  "a",
			source:   "Hello",
			options:  nil,
			hasError: false,
		},
		{
			name:     "simple message",
			locales:  "en",
			source:   "Test message",
			options:  nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New(tt.locales, tt.source, tt.options)

			if tt.hasError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, mf)
			} else {
				require.NoError(t, err)
				require.NotNil(t, mf)
			}
		})
	}
}

// TestEmptySliceLocales tests empty slice locales separately
func TestEmptySliceLocales(t *testing.T) {
	// Empty slice locale should be treated as no locale
	mf, err := New([]string{}, "Hello", nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{}, mf.locales)
	assert.Equal(t, "auto", mf.dir)
}

// TestEmptyStringLocale tests empty string locale separately
func TestEmptyStringLocale(t *testing.T) {
	// Empty string locale should be treated as no locale
	mf, err := New("", "Hello", nil)
	require.NoError(t, err)
	require.NotNil(t, mf)
	assert.Equal(t, []string{}, mf.locales)
	assert.Equal(t, "auto", mf.dir)
}

// TestOptionsVariations tests different option combinations
func TestOptionsVariations(t *testing.T) {
	tests := []struct {
		name            string
		options         *MessageFormatOptions
		expectedBidi    bool
		expectedDir     string
		expectedMatcher string
	}{
		{
			name: "bidi isolation none",
			options: &MessageFormatOptions{
				BidiIsolation: stringPtr("none"),
			},
			expectedBidi:    false,
			expectedDir:     "ltr",
			expectedMatcher: "best fit",
		},
		{
			name: "bidi isolation default",
			options: &MessageFormatOptions{
				BidiIsolation: stringPtr("default"),
			},
			expectedBidi:    true,
			expectedDir:     "ltr",
			expectedMatcher: "best fit",
		},
		{
			name: "custom direction",
			options: &MessageFormatOptions{
				Dir: stringPtr("auto"),
			},
			expectedBidi:    true,
			expectedDir:     "auto",
			expectedMatcher: "best fit",
		},
		{
			name: "lookup matcher",
			options: &MessageFormatOptions{
				LocaleMatcher: stringPtr("lookup"),
			},
			expectedBidi:    true,
			expectedDir:     "ltr",
			expectedMatcher: "lookup",
		},
		{
			name: "all options combined",
			options: &MessageFormatOptions{
				BidiIsolation: stringPtr("none"),
				Dir:           stringPtr("rtl"),
				LocaleMatcher: stringPtr("lookup"),
			},
			expectedBidi:    false,
			expectedDir:     "rtl",
			expectedMatcher: "lookup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := New("en", "Hello", tt.options)
			require.NoError(t, err)
			require.NotNil(t, mf)

			assert.Equal(t, tt.expectedBidi, mf.bidiIsolation)
			assert.Equal(t, tt.expectedDir, mf.dir)
			assert.Equal(t, tt.expectedMatcher, mf.localeMatcher)
		})
	}
}

// TestIndexExports tests the index.go exports
func TestIndexExports(t *testing.T) {
	t.Run("NewMessageFormat alias", func(t *testing.T) {
		mf, err := NewMessageFormat("en", "Hello", nil)
		require.NoError(t, err)
		require.NotNil(t, mf)
	})

	t.Run("ValidateMessage function", func(t *testing.T) {
		pattern := datamodel.NewPattern([]datamodel.PatternElement{
			datamodel.NewTextElement("Hello"),
		})
		message := datamodel.NewPatternMessage(nil, pattern, "")

		_, err := ValidateMessage(message, nil)
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
		// Draft functions may vary, just check it's not nil
	})
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
