package messageformat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Core MessageFormat 2.0 Syntax Tests

// TestPatternMatching tests the .match syntax for pattern selection
func TestPatternMatching(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		values   map[string]interface{}
		expected string
	}{
		{
			name: "zero_case",
			message: `.input {$count :number}
.match $count
0   {{You have no new notifications}}
one {{You have {$count} new notification}}
*   {{You have {$count} new notifications}}`,
			values:   map[string]interface{}{"count": 0},
			expected: "You have no new notifications",
		},
		{
			name: "singular_case",
			message: `.input {$count :number}
.match $count
0   {{You have no new notifications}}
one {{You have {$count} new notification}}
*   {{You have {$count} new notifications}}`,
			values:   map[string]interface{}{"count": 1},
			expected: "You have 1 new notification",
		},
		{
			name: "plural_case",
			message: `.input {$count :number}
.match $count
0   {{You have no new notifications}}
one {{You have {$count} new notification}}
*   {{You have {$count} new notifications}}`,
			values:   map[string]interface{}{"count": 5},
			expected: "You have 5 new notifications",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse message successfully")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message successfully")
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestInputDeclarations tests .input variable declarations with formatting
func TestInputDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		values   map[string]interface{}
		expected string
	}{
		{
			name: "currency_input",
			message: `.input {$amount :number style=currency currency=EUR}
{{Your balance is {$amount}}}`,
			values:   map[string]interface{}{"amount": 42.50},
			expected: "Your balance is €42.50",
		},
		{
			name: "decimal_input",
			message: `.input {$price :number style=decimal}
{{Price: {$price}}}`,
			values:   map[string]interface{}{"price": 123.456},
			expected: "Price: 123.456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse message successfully")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message successfully")
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestLocalVariables tests .local variable declarations
func TestLocalVariables(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		values   map[string]interface{}
		expected string
	}{
		{
			name: "tax_calculation",
			message: `.input {$price :number}
.input {$taxRate :number}
.local $tax = {$taxRate :number style=percent}
.local $total = {$price :number style=currency currency=USD}
{{Item: {$total} (includes {$tax} tax)}}`,
			values: map[string]interface{}{
				"price":   100.0,
				"taxRate": 0.15,
			},
			expected: "Item: $100.00 (includes 15% tax)",
		},
		{
			name: "multiple_locals",
			message: `.input {$base :number}
.local $doubled = {$base :number}
.local $tripled = {$base :number}
{{Base: {$base}, Doubled: {$doubled}, Tripled: {$tripled}}}`,
			values:   map[string]interface{}{"base": 10},
			expected: "Base: 10, Doubled: 10, Tripled: 10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse message successfully")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message successfully")
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Number and Currency Formatting Tests

// TestCurrencyFormatting tests currency formatting with different currencies
func TestCurrencyFormatting(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		currency string
		amount   float64
		expected string
	}{
		{"usd", "en", "USD", 42.50, "Price: $42.50"},
		{"eur", "en", "EUR", 42.50, "Price: €42.50"},
		{"gbp", "en", "GBP", 42.50, "Price: £42.50"},
		{"jpy", "en", "JPY", 1000, "Price: ¥1,000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			message := fmt.Sprintf("Price: {$amount :number style=currency currency=%s}", tc.currency)
			mf, err := New(tc.locale, message, nil)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{"amount": tc.amount})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestPercentageFormatting tests percentage formatting
func TestPercentageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"three_quarters", 0.75, "Progress: 75%"},
		{"with_decimals", 0.123, "Completion: 12.3%"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var message string
			if tc.name == "three_quarters" {
				message = "Progress: {$rate :number style=percent}"
			} else {
				message = "Completion: {$rate :number style=percent}"
			}

			mf, err := New("en", message, nil)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{"rate": tc.value})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestNumberPluralSelection tests number function with plural selection
func TestNumberPluralSelection(t *testing.T) {
	message := `.input {$count :number}
.match $count
one {{You have {$count} item.}}
*   {{You have {$count} items.}}`

	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"singular", 1, "You have 1 item."},
		{"plural", 5, "You have 5 items."},
		{"zero", 0, "You have 0 items."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", message, nil)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{"count": tc.count})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestExactNumberMatching tests exact numeric matching priority
func TestExactNumberMatching(t *testing.T) {
	message := `.input {$count :number}
.match $count
0   {{no items}}
1   {{one item}}
*   {{{$count} items}}`

	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"exact_zero", 0, "no items"},
		{"exact_one", 1, "one item"},
		{"fallback", 5, "5 items"},
	}

	// Test exact match priority over plural categories
	priorityMessage := `.input {$count :number}
.match $count
1   {{exactly one}}
one {{plural one}}
*   {{other}}`

	t.Run("exact_priority", func(t *testing.T) {
		mf, err := New("en", priorityMessage, nil)
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"count": 1})
		require.NoError(t, err)
		assert.Equal(t, "exactly one", result, "exact match should have priority over plural category")
	})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", message, nil)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{"count": tc.count})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestStringSelection tests string-based pattern matching
func TestStringSelection(t *testing.T) {
	statusMessage := `.input {$status :string}
.match $status
online  {{🟢 Online}}
offline {{🔴 Offline}}
*       {{❓ Unknown}}`

	roleMessage := `.input {$role :string}
.match $role
admin     {{👑 Administrator}}
moderator {{🛡️ Moderator}}
user      {{👤 User}}
*         {{❓ Unknown Role}}`

	tests := []struct {
		name     string
		message  string
		key      string
		value    string
		expected string
	}{
		{"status_online", statusMessage, "status", "online", "🟢 Online"},
		{"status_offline", statusMessage, "status", "offline", "🔴 Offline"},
		{"status_unknown", statusMessage, "status", "unknown", "❓ Unknown"},
		{"role_admin", roleMessage, "role", "admin", "👑 Administrator"},
		{"role_moderator", roleMessage, "role", "moderator", "🛡️ Moderator"},
		{"role_user", roleMessage, "role", "user", "👤 User"},
		{"role_unknown", roleMessage, "role", "guest", "❓ Unknown Role"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err)

			result, err := mf.Format(map[string]interface{}{tc.key: tc.value})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Custom Functions and Advanced Features

// TestCustomFunctions tests custom function integration
func TestCustomFunctions(t *testing.T) {
	// Helper function to create uppercase function
	createUppercaseFunc := func() functions.MessageFunction {
		return func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			input interface{},
		) messagevalue.MessageValue {
			inputStr := fmt.Sprintf("%v", input)
			upperStr := strings.ToUpper(inputStr)
			locale := "en"
			if locales := ctx.Locales(); len(locales) > 0 {
				locale = locales[0]
			}
			return messagevalue.NewStringValue(upperStr, locale, ctx.Source())
		}
	}

	// Helper function to create reverse function
	createReverseFunc := func() functions.MessageFunction {
		return func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			input interface{},
		) messagevalue.MessageValue {
			inputStr := fmt.Sprintf("%v", input)
			runes := []rune(inputStr)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			reversedStr := string(runes)
			locale := "en"
			if locales := ctx.Locales(); len(locales) > 0 {
				locale = locales[0]
			}
			return messagevalue.NewStringValue(reversedStr, locale, ctx.Source())
		}
	}

	tests := []struct {
		name      string
		message   string
		values    map[string]interface{}
		functions map[string]functions.MessageFunction
		expected  string
	}{
		{
			name:    "uppercase_function",
			message: "Hello {$name :uppercase}!",
			values:  map[string]interface{}{"name": "world"},
			functions: map[string]functions.MessageFunction{
				"uppercase": createUppercaseFunc(),
			},
			expected: "Hello \u2068WORLD\u2069!",
		},
		{
			name:    "reverse_function",
			message: "Reversed: {$text :reverse}",
			values:  map[string]interface{}{"text": "hello"},
			functions: map[string]functions.MessageFunction{
				"reverse": createReverseFunc(),
			},
			expected: "Reversed: \u2068olleh\u2069",
		},
		{
			name:    "multiple_functions",
			message: "Name: {$first :uppercase} {$last :reverse}",
			values:  map[string]interface{}{"first": "john", "last": "doe"},
			functions: map[string]functions.MessageFunction{
				"uppercase": createUppercaseFunc(),
				"reverse":   createReverseFunc(),
			},
			expected: "Name: \u2068JOHN\u2069 \u2068eod\u2069",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := &MessageFormatOptions{Functions: tc.functions}
			mf, err := New("en", tc.message, opts)
			require.NoError(t, err)

			result, err := mf.Format(tc.values)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Markup and Syntax Features

// TestEscapeSequences tests escape sequence handling
func TestEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		values   map[string]interface{}
		expected string
	}{
		{
			name:     "escaped_braces",
			message:  "Just braces: {{ and }}",
			values:   nil,
			expected: "Just braces: { and }",
		},
		{
			name:     "message_block",
			message:  "{{Hello}}",
			values:   nil,
			expected: "Hello",
		},
		{
			name:     "mixed_content",
			message:  "Text {{ escaped }} more text",
			values:   nil,
			expected: "Text { escaped } more text",
		},
		{
			name:     "escaped_with_variable",
			message:  "Object: {{ key: {$key} }}",
			values:   map[string]interface{}{"key": "name"},
			expected: "Object: { key: \u2068name\u2069 }",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse message with escape sequences")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message with escape sequences")
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestMarkupPlaceholders tests markup placeholder support
func TestMarkupPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "bold_markup",
			message:  "Welcome {#b}bold text{/b} and normal text",
			expected: "Welcome bold text and normal text",
		},
		{
			name:     "standalone_markup",
			message:  "Image: {#img src=logo.png /} here",
			expected: "Image:  here",
		},
		{
			name:     "mixed_markup",
			message:  "Text with {#b}bold{/b} and {#i}italic{/i} and {#br /} line break",
			expected: "Text with bold and italic and  line break",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse markup syntax")

			result, err := mf.Format(nil)
			require.NoError(t, err, "should format markup")
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestMarkupFormatToParts tests markup in formatToParts output
func TestMarkupFormatToParts(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedParts []struct {
			partType string
			kind     string
			name     string
		}
	}{
		{
			name:    "bold_markup_parts",
			message: "Welcome {#b}bold text{/b} and normal text",
			expectedParts: []struct {
				partType string
				kind     string
				name     string
			}{
				{"text", "", ""},
				{"markup", "open", "b"},
				{"text", "", ""},
				{"markup", "close", "b"},
				{"text", "", ""},
			},
		},
		{
			name:    "standalone_markup_parts",
			message: "Image: {#img src=logo.png /} here",
			expectedParts: []struct {
				partType string
				kind     string
				name     string
			}{
				{"text", "", ""},
				{"markup", "standalone", "img"},
				{"text", "", ""},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err)

			parts, err := mf.FormatToParts(nil)
			require.NoError(t, err)
			assert.Len(t, parts, len(tc.expectedParts))

			for i, expectedPart := range tc.expectedParts {
				if i >= len(parts) {
					t.Fatalf("Expected part %d not found", i)
				}

				part := parts[i]
				assert.Equal(t, expectedPart.partType, part.Type(), "Part %d type mismatch", i)

				if expectedPart.partType == "markup" {
					if mp, ok := part.(*messagevalue.MarkupPart); ok {
						assert.Equal(t, expectedPart.kind, mp.Kind(), "Part %d kind mismatch", i)
						assert.Equal(t, expectedPart.name, mp.Name(), "Part %d name mismatch", i)
					} else {
						t.Errorf("Part %d should be MarkupPart but got %T", i, part)
					}
				}
			}
		})
	}
}

// API and Configuration Tests

// TestResolvedOptions tests the ResolvedOptions method
func TestResolvedOptions(t *testing.T) {
	tests := []struct {
		name     string
		locales  string
		message  string
		options  *MessageFormatOptions
		expected ResolvedMessageFormatOptions
	}{
		{
			name:    "default_options",
			locales: "en",
			message: "Hello {$name}!",
			options: nil,
			expected: ResolvedMessageFormatOptions{
				BidiIsolation: BidiDefault,
				Dir:           DirLTR,
				LocaleMatcher: LocaleBestFit,
				Functions:     nil, // Will be checked separately
			},
		},
		{
			name:    "custom_options",
			locales: "ar",
			message: "مرحبا {$name}!",
			options: &MessageFormatOptions{
				BidiIsolation: BidiNone,
				Dir:           DirRTL,
				LocaleMatcher: LocaleLookup,
			},
			expected: ResolvedMessageFormatOptions{
				BidiIsolation: BidiNone,
				Dir:           DirRTL,
				LocaleMatcher: LocaleLookup,
				Functions:     nil, // Will be checked separately
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New(tc.locales, tc.message, tc.options)
			require.NoError(t, err)

			resolved := mf.ResolvedOptions()

			assert.Equal(t, tc.expected.BidiIsolation, resolved.BidiIsolation)
			assert.Equal(t, tc.expected.Dir, resolved.Dir)
			assert.Equal(t, tc.expected.LocaleMatcher, resolved.LocaleMatcher)

			// Check that functions map is not nil and contains default functions
			assert.NotNil(t, resolved.Functions)
			assert.Contains(t, resolved.Functions, "number")
		})
	}
}

// TestErrorHandling tests error handling mechanisms
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		values         map[string]interface{}
		expectedResult string
		expectError    bool
		errorCallback  bool
	}{
		{
			name:           "missing_variable_fallback",
			message:        "Hello {$missing}!",
			values:         map[string]interface{}{},
			expectedResult: "Hello \u2068{$missing}\u2069!",
			expectError:    false,
			errorCallback:  false,
		},
		{
			name:           "unknown_function_fallback",
			message:        "Value: {$value :unknown}",
			values:         map[string]interface{}{"value": "test"},
			expectedResult: "Value: \u2068{$value}\u2069",
			expectError:    false,
			errorCallback:  false,
		},
		{
			name:           "normal_case",
			message:        "Hello {$name}!",
			values:         map[string]interface{}{"name": "World"},
			expectedResult: "Hello \u2068World\u2069!",
			expectError:    false,
			errorCallback:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errorCalled bool
			var capturedError error

			onError := func(err error) {
				errorCalled = true
				capturedError = err
			}

			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should parse message")

			result, err := mf.Format(tc.values, onError)

			if tc.expectError {
				assert.Error(t, err, "should return error for invalid cases")
			} else {
				assert.NoError(t, err, "should not return error, should use fallback representation")
			}

			if tc.errorCallback {
				assert.True(t, errorCalled, "error callback should be called")
				assert.NotNil(t, capturedError, "should capture error in callback")
			}

			assert.Equal(t, tc.expectedResult, result)
		})
	}

	// Test error callback functionality
	t.Run("error_callback_functionality", func(t *testing.T) {
		var errors []error
		onError := func(err error) {
			errors = append(errors, err)
		}

		mf, err := New("en", "Hello {$name}!", nil)
		require.NoError(t, err)

		result, err := mf.Format(map[string]interface{}{"name": "World"}, onError)
		require.NoError(t, err)
		assert.Equal(t, "Hello \u2068World\u2069!", result)
		assert.Empty(t, errors, "no errors should be captured for valid formatting")
	})
}

// TestMessageDataInterface tests MessageData interface support
func TestMessageDataInterface(t *testing.T) {
	tests := []struct {
		name     string
		message  datamodel.Message
		values   map[string]interface{}
		expected string
	}{
		{
			name: "pattern_message_simple_text",
			message: datamodel.NewPatternMessage(
				nil, // no declarations
				datamodel.NewPattern([]datamodel.PatternElement{
					datamodel.NewTextElement("Hello World!"),
				}),
				"", // no comment
			),
			values:   nil,
			expected: "Hello World!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err, "should create MessageFormat with datamodel.Message")

			assert.Equal(t, tc.message.Type(), "message")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message from datamodel.Message")

			if tc.name == "pattern_message_simple_text" {
				assert.Equal(t, tc.expected, result)
			}
		})
	}

	// Test SelectMessage support
	t.Run("select_message_support", func(t *testing.T) {
		selectors := []datamodel.VariableRef{
			*datamodel.NewVariableRef("count"),
		}

		variants := []datamodel.Variant{
			*datamodel.NewVariant(
				[]datamodel.VariantKey{datamodel.NewCatchallKey("*")},
				datamodel.NewPattern([]datamodel.PatternElement{
					datamodel.NewTextElement("Default message"),
				}),
			),
		}

		selectMessage := datamodel.NewSelectMessage(
			nil, // no declarations
			selectors,
			variants,
			"", // no comment
		)

		assert.Equal(t, selectMessage.Type(), "select")

		_, err := New("en", selectMessage, nil)
		if err != nil {
			// If validation fails, that's expected for this minimal example
			assert.Contains(t, err.Error(), "missing-selector-annotation",
				"error should be about missing selector annotation, not interface rejection")
		}
	})
}

// Internationalization and Localization Tests

// TestBidirectionalTextSupport tests bidirectional text support
func TestBidirectionalTextSupport(t *testing.T) {
	tests := []struct {
		name            string
		locale          string
		message         string
		options         *MessageFormatOptions
		values          map[string]interface{}
		expectedDir     Direction
		expectedBidi    BidiIsolation
		containsIsolate bool
	}{
		{
			name:            "arabic_rtl_auto_detection",
			locale:          "ar",
			message:         "مرحبا {$name}!",
			options:         nil,
			values:          map[string]interface{}{"name": "أحمد"},
			expectedDir:     DirRTL,
			expectedBidi:    BidiDefault,
			containsIsolate: true,
		},
		{
			name:    "hebrew_rtl",
			locale:  "he",
			message: "שלום {$name}!",
			options: &MessageFormatOptions{
				BidiIsolation: BidiDefault,
			},
			values:          map[string]interface{}{"name": "דוד"},
			expectedDir:     DirRTL,
			expectedBidi:    BidiDefault,
			containsIsolate: true,
		},
		{
			name:    "english_explicit_rtl",
			locale:  "en",
			message: "Hello {$name}!",
			options: &MessageFormatOptions{
				Dir:           DirRTL,
				BidiIsolation: BidiDefault,
			},
			values:          map[string]interface{}{"name": "World"},
			expectedDir:     DirRTL,
			expectedBidi:    BidiDefault,
			containsIsolate: true,
		},
		{
			name:    "arabic_bidi_disabled",
			locale:  "ar",
			message: "مرحبا {$name}!",
			options: &MessageFormatOptions{
				BidiIsolation: BidiNone,
			},
			values:          map[string]interface{}{"name": "أحمد"},
			expectedDir:     DirRTL,
			expectedBidi:    BidiNone,
			containsIsolate: false,
		},
		{
			name:    "mixed_ltr_rtl_content",
			locale:  "ar",
			message: "Email: {$email} - مرحبا {$name}!",
			options: &MessageFormatOptions{
				BidiIsolation: BidiDefault,
			},
			values: map[string]interface{}{
				"email": "user@example.com",
				"name":  "أحمد",
			},
			expectedDir:     DirRTL,
			expectedBidi:    BidiDefault,
			containsIsolate: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New(tc.locale, tc.message, tc.options)
			require.NoError(t, err, "should create MessageFormat with RTL support")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format RTL message")
			assert.NotEmpty(t, result, "formatted result should not be empty")

			// Test ResolvedOptions
			resolved := mf.ResolvedOptions()
			assert.Equal(t, tc.expectedDir, resolved.Dir, "direction should match expected")
			assert.Equal(t, tc.expectedBidi, resolved.BidiIsolation, "BidiIsolation should match expected")

			// Test bidi isolation characters presence
			if tc.containsIsolate {
				hasBidiChars := strings.Contains(result, "\u2068") || strings.Contains(result, "\u2069") ||
					strings.Contains(result, "\u202D") || strings.Contains(result, "\u202C")
				assert.True(t, hasBidiChars, "result should contain bidi isolation characters when enabled")
			} else {
				hasBidiChars := strings.Contains(result, "\u2068") || strings.Contains(result, "\u2069") ||
					strings.Contains(result, "\u202D") || strings.Contains(result, "\u202C")
				assert.False(t, hasBidiChars, "result should NOT contain bidi isolation characters when disabled")
			}
		})
	}
}

// TestLocaleNegotiation tests locale negotiation
func TestLocaleNegotiation(t *testing.T) {
	tests := []struct {
		name          string
		locales       interface{}
		message       string
		options       *MessageFormatOptions
		values        map[string]interface{}
		expectedDir   Direction
		expectedMatch LocaleMatcher
	}{
		{
			name:    "single_locale_string",
			locales: "en-US",
			message: "Hello {$name}!",
			options: &MessageFormatOptions{
				LocaleMatcher: LocaleBestFit,
			},
			values:        map[string]interface{}{"name": "World"},
			expectedDir:   DirLTR,
			expectedMatch: LocaleBestFit,
		},
		{
			name:    "multiple_locales_array",
			locales: []string{"zh-CN", "en", "fr"},
			message: "Hello {$name}!",
			options: &MessageFormatOptions{
				LocaleMatcher: LocaleLookup,
			},
			values:        map[string]interface{}{"name": "世界"},
			expectedDir:   DirLTR,
			expectedMatch: LocaleLookup,
		},
		{
			name:    "rtl_locale",
			locales: "ar",
			message: "مرحبا {$name}!",
			options: &MessageFormatOptions{
				LocaleMatcher: LocaleBestFit,
			},
			values:        map[string]interface{}{"name": "أحمد"},
			expectedDir:   DirRTL,
			expectedMatch: LocaleBestFit,
		},
		{
			name:    "explicit_dir_override",
			locales: "en",
			message: "Hello {$name}!",
			options: &MessageFormatOptions{
				Dir:           DirRTL,
				LocaleMatcher: LocaleLookup,
			},
			values:        map[string]interface{}{"name": "World"},
			expectedDir:   DirRTL,
			expectedMatch: LocaleLookup,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New(tc.locales, tc.message, tc.options)
			require.NoError(t, err, "should create MessageFormat with locale negotiation")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format message")
			assert.NotEmpty(t, result, "formatted result should not be empty")

			// Test ResolvedOptions to verify locale negotiation
			resolved := mf.ResolvedOptions()
			assert.Equal(t, tc.expectedDir, resolved.Dir, "direction should match expected")
			assert.Equal(t, tc.expectedMatch, resolved.LocaleMatcher, "LocaleMatcher should match expected")

			// Verify that functions are available
			assert.NotNil(t, resolved.Functions, "functions should be available")
			assert.Contains(t, resolved.Functions, "number", "default number function should be available")
		})
	}
}

// TestFormatToParts tests the formatToParts method
func TestFormatToParts(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		values   map[string]interface{}
		expected int // Expected number of parts
	}{
		{
			name:     "simple_message_with_variable",
			message:  "Hello {$name}!",
			values:   map[string]interface{}{"name": "World"},
			expected: 5, // "Hello ", bidi_start, "World", bidi_end, "!"
		},
		{
			name:     "currency_formatting",
			message:  "Price: {$amount :number style=currency currency=USD}",
			values:   map[string]interface{}{"amount": 42.50},
			expected: 2, // "Price: ", "$42.50"
		},
		{
			name:     "plain_text",
			message:  "Hello World!",
			values:   nil,
			expected: 1, // "Hello World!"
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			require.NoError(t, err)

			parts, err := mf.FormatToParts(tc.values)
			require.NoError(t, err)

			// Check that we get the expected number of parts
			assert.Len(t, parts, tc.expected)

			// Check that all parts have required fields
			for i, part := range parts {
				assert.NotNil(t, part, "part %d should not be nil", i)
				assert.NotEmpty(t, part.Type(), "part %d should have a type", i)
				assert.NotNil(t, part.Value(), "part %d should have a value", i)
			}

			// Verify that concatenating parts gives the same result as Format()
			formatResult, err := mf.Format(tc.values)
			require.NoError(t, err)

			var partsResult strings.Builder
			for _, part := range parts {
				if str, ok := part.Value().(string); ok {
					partsResult.WriteString(str)
				} else {
					partsResult.WriteString(fmt.Sprintf("%v", part.Value()))
				}
			}

			assert.Equal(t, formatResult, partsResult.String())
		})
	}
}

// Edge Cases and Error Handling Tests

// TestEdgeCases tests edge case handling
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		values      map[string]interface{}
		expected    string
		shouldError bool
	}{
		{
			name:        "empty_message",
			message:     "",
			values:      nil,
			expected:    "",
			shouldError: false,
		},
		{
			name:        "missing_variable_fallback",
			message:     "Hello {$missing}!",
			values:      map[string]interface{}{},
			expected:    "Hello \u2068{$missing}\u2069!",
			shouldError: false,
		},
		{
			name:        "whitespace_only",
			message:     "   ",
			values:      nil,
			expected:    "   ",
			shouldError: false,
		},
		{
			name:        "special_characters",
			message:     "Special: !@#$%^&*()_+-=[]|;':\",./<>?",
			values:      nil,
			expected:    "Special: !@#$%^&*()_+-=[]|;':\",./<>?",
			shouldError: false,
		},
		{
			name:        "unicode_characters",
			message:     "Unicode: 🌍🚀💻🎉",
			values:      nil,
			expected:    "Unicode: 🌍🚀💻🎉",
			shouldError: false,
		},
		{
			name:        "mixed_rtl_ltr",
			message:     "English مرحبا Hebrew שלום",
			values:      nil,
			expected:    "English مرحبا Hebrew שלום",
			shouldError: false,
		},
		{
			name:        "null_value",
			message:     "Value: {$value}",
			values:      map[string]interface{}{"value": nil},
			expected:    "Value: \u2068{$value}\u2069",
			shouldError: false,
		},
		{
			name:        "zero_value",
			message:     "Count: {$count}",
			values:      map[string]interface{}{"count": 0},
			expected:    "Count: 0",
			shouldError: false,
		},
		{
			name:        "empty_string",
			message:     "Text: {$text}",
			values:      map[string]interface{}{"text": ""},
			expected:    "Text: \u2068\u2069",
			shouldError: false,
		},
		{
			name:        "boolean_false",
			message:     "Flag: {$flag}",
			values:      map[string]interface{}{"flag": false},
			expected:    "Flag: \u2068false\u2069",
			shouldError: false,
		},
		{
			name:        "boolean_true",
			message:     "Flag: {$flag}",
			values:      map[string]interface{}{"flag": true},
			expected:    "Flag: \u2068true\u2069",
			shouldError: false,
		},
		{
			name:        "escaped_braces",
			message:     "Object: {{ key: value }}",
			values:      nil,
			expected:    "Object: { key: value }",
			shouldError: false,
		},
		{
			name:        "multiple_variables",
			message:     "{$a} {$b} {$c}",
			values:      map[string]interface{}{"a": "A", "b": "B", "c": "C"},
			expected:    "\u2068A\u2069 \u2068B\u2069 \u2068C\u2069",
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf, err := New("en", tc.message, nil)
			if tc.shouldError {
				if err != nil {
					return // Expected error during parsing
				}
			} else {
				require.NoError(t, err, "should parse message")
			}

			result, err := mf.Format(tc.values)
			if tc.shouldError {
				assert.Error(t, err, "should return error for invalid cases")
			} else {
				require.NoError(t, err, "should not return error, should use fallback representation")
				assert.Equal(t, tc.expected, result, "result should match expected fallback")
			}
		})
	}
}

// TestConcurrentAccess tests concurrent access safety
func TestConcurrentAccess(t *testing.T) {
	mf, err := New("en", "Hello {$name}!", nil)
	require.NoError(t, err)

	const numGoroutines = 10
	const numIterations = 10

	results := make(chan string, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				result, err := mf.Format(map[string]interface{}{
					"name": fmt.Sprintf("User%d-%d", id, j),
				})
				if err == nil {
					results <- result
				}
			}
		}(i)
	}

	// Collect results
	var collectedResults []string
	for i := 0; i < numGoroutines*numIterations; i++ {
		result := <-results
		collectedResults = append(collectedResults, result)
	}

	// Should have all results
	assert.Len(t, collectedResults, numGoroutines*numIterations)

	// All results should be valid
	for _, result := range collectedResults {
		assert.Contains(t, result, "Hello")
		assert.Contains(t, result, "User")
	}
}

// TestComplexScenarios tests complex feature combinations
func TestComplexScenarios(t *testing.T) {
	// Helper functions for custom functions
	createUppercaseFunc := func() functions.MessageFunction {
		return func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			input interface{},
		) messagevalue.MessageValue {
			inputStr := fmt.Sprintf("%v", input)
			upperStr := strings.ToUpper(inputStr)
			locale := "en"
			if locales := ctx.Locales(); len(locales) > 0 {
				locale = locales[0]
			}
			return messagevalue.NewStringValue(upperStr, locale, ctx.Source())
		}
	}

	createReverseFunc := func() functions.MessageFunction {
		return func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			input interface{},
		) messagevalue.MessageValue {
			inputStr := fmt.Sprintf("%v", input)
			runes := []rune(inputStr)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			reversedStr := string(runes)
			locale := "en"
			if locales := ctx.Locales(); len(locales) > 0 {
				locale = locales[0]
			}
			return messagevalue.NewStringValue(reversedStr, locale, ctx.Source())
		}
	}

	tests := []struct {
		name      string
		message   string
		values    map[string]interface{}
		expected  string
		functions map[string]functions.MessageFunction
	}{
		{
			name: "exact_number_matching",
			message: `.input {$count :number}
.match $count
0        {{No items found}}
1        {{Found one item}}
*        {{Found {$count} items}}`,
			values:   map[string]interface{}{"count": 5},
			expected: "Found 5 items",
		},
		{
			name:     "currency_with_markup",
			message:  "Price: {#b}{$amount :number style=currency currency=USD}{/b}",
			values:   map[string]interface{}{"amount": 99.99},
			expected: "Price: $99.99",
		},
		{
			name:     "multiple_currencies",
			message:  "USD: {$usd :number style=currency currency=USD}, EUR: {$eur :number style=currency currency=EUR}",
			values:   map[string]interface{}{"usd": 100.50, "eur": 85.75},
			expected: "USD: $100.50, EUR: €85.75",
		},
		{
			name:     "percentage_with_custom_function",
			message:  "Progress: {$rate :number style=percent} - Status: {$status :uppercase}",
			values:   map[string]interface{}{"rate": 0.75, "status": "completed"},
			expected: "Progress: 75% - Status: \u2068COMPLETED\u2069",
			functions: map[string]functions.MessageFunction{
				"uppercase": createUppercaseFunc(),
			},
		},
		{
			name:     "nested_markup_with_variables",
			message:  "Welcome {#strong}{$name}{/strong}! You have {#em}{$count :number}{/em} new messages.",
			values:   map[string]interface{}{"name": "Alice", "count": 5},
			expected: "Welcome \u2068Alice\u2069! You have 5 new messages.",
		},
		{
			name:     "mixed_content_with_formatToParts",
			message:  "Order #{$id}: {$amount :number style=currency currency=USD} for {$items :number} items",
			values:   map[string]interface{}{"id": "12345", "amount": 299.99, "items": 3},
			expected: "Order #\u206812345\u2069: $299.99 for 3 items",
		},
		{
			name:     "escape_sequences_with_variables",
			message:  "Config: {{ \"key\": \"{$value}\" }}",
			values:   map[string]interface{}{"value": "test"},
			expected: "Config: { \"key\": \"\u2068test\u2069\" }",
		},
		{
			name:     "complex_number_formatting",
			message:  "Stats: {$big :number}, {$decimal :number}, {$percent :number style=percent}",
			values:   map[string]interface{}{"big": 1234567, "decimal": 123.456, "percent": 0.85},
			expected: "Stats: 1,234,567, 123.456, 85%",
		},
		{
			name:     "multiple_custom_functions",
			message:  "Name: {$first :uppercase} {$last :reverse}, Email: {$email}",
			values:   map[string]interface{}{"first": "john", "last": "doe", "email": "john@example.com"},
			expected: "Name: \u2068JOHN\u2069 \u2068eod\u2069, Email: \u2068john@example.com\u2069",
			functions: map[string]functions.MessageFunction{
				"uppercase": createUppercaseFunc(),
				"reverse":   createReverseFunc(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var opts *MessageFormatOptions
			if tc.functions != nil {
				opts = &MessageFormatOptions{Functions: tc.functions}
			}

			locale := "en"
			if strings.Contains(tc.message, "مرحبا") {
				locale = "ar"
			}

			mf, err := New(locale, tc.message, opts)
			require.NoError(t, err, "should parse complex message")

			result, err := mf.Format(tc.values)
			require.NoError(t, err, "should format complex message")
			assert.Equal(t, tc.expected, result, "complex formatting should match expected result")
		})
	}
}
