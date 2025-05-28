package messageformat

import (
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

// TestMustNew tests the MustNew constructor
func TestMustNew(t *testing.T) {
	t.Run("valid input - should not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNew panicked unexpectedly: %v", r)
			}
		}()

		mf := MustNew("en", "Hello World", nil)
		require.NotNil(t, mf)
		assert.Equal(t, []string{"en"}, mf.locales)
	})

	t.Run("invalid input - should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNew should have panicked but didn't")
			}
		}()

		// This should panic due to invalid locales type
		MustNew(123, "Hello", nil)
	})

	t.Run("syntax error - should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNew should have panicked but didn't")
			}
		}()

		// This should panic due to syntax error
		MustNew("en", "Hello {$name", nil)
	})

	t.Run("with options - should not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNew panicked unexpectedly: %v", r)
			}
		}()

		options := &MessageFormatOptions{
			BidiIsolation: BidiNone,
			Dir:           DirRTL,
		}
		mf := MustNew("en", "Hello", options)
		require.NotNil(t, mf)
		assert.False(t, mf.bidiIsolation)
		assert.Equal(t, "rtl", mf.dir)
	})
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

// Index Exports API Tests

// TestIndexExports tests the exported functions from index.go
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
