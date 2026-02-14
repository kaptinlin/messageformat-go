package v1

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticMessageFormat(t *testing.T) {
	t.Run("should exist", func(t *testing.T) {
		mf, err := New("en", nil)
		require.NoError(t, err)
		require.NotNil(t, mf)
	})

	t.Run("should have a supportedLocalesOf() function", func(t *testing.T) {
		result, err := SupportedLocalesOf("en")
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("supportedLocalesOf([]string)", func(t *testing.T) {
		lc, err := SupportedLocalesOf([]string{"fi", "xx", "en-CA"})
		require.NoError(t, err)
		require.NotEmpty(t, lc)

		hasUnsupported := slices.Contains(lc, "xx")
		assert.False(t, hasUnsupported, "Should not include unsupported locale 'xx'")
	})

	t.Run("supportedLocalesOf(string)", func(t *testing.T) {
		lc, err := SupportedLocalesOf("en")
		require.NoError(t, err)
		expected := []string{"en"}
		assert.Equal(t, expected, lc)
	})
}

func TestMessageFormatConstructor(t *testing.T) {
	t.Run("should be a constructor", func(t *testing.T) {
		mf, err := New("en", nil)
		require.NoError(t, err)
		require.NotNil(t, mf)
	})

	t.Run("should have a Compile() function", func(t *testing.T) {
		mf, err := New("en", nil)
		require.NoError(t, err)
		_, err = mf.Compile("test")
		require.NoError(t, err)
	})

	t.Run("should have a ResolvedOptions() function", func(t *testing.T) {
		mf, err := New("en", nil)
		require.NoError(t, err)
		opts := mf.ResolvedOptions()
		assert.NotEmpty(t, opts.Locale)
	})

	t.Run("should fallback on non-existing locales", func(t *testing.T) {
		mf, err := New("lawlz", nil)
		require.NoError(t, err)
		opt := mf.ResolvedOptions()
		assert.Equal(t, "en", opt.Locale)
	})

	t.Run("should default to en when no locale is passed", func(t *testing.T) {
		mf, err := New(nil, nil)
		require.NoError(t, err)
		opt := mf.ResolvedOptions()
		assert.Equal(t, "en", opt.Locale)
	})
}

func TestMessageFormatOptions(t *testing.T) {
	t.Run("should apply default currency", func(t *testing.T) {
		mf, err := New("en", &MessageFormatOptions{
			Currency: "EUR",
		})
		require.NoError(t, err)
		opts := mf.ResolvedOptions()
		assert.Equal(t, "EUR", opts.Currency)
	})

	t.Run("should apply strict mode", func(t *testing.T) {
		mf, err := New("en", &MessageFormatOptions{
			Strict: true,
		})
		require.NoError(t, err)
		opts := mf.ResolvedOptions()
		assert.True(t, opts.Strict)

		_, err = mf.Compile("{foo, bar}")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid strict mode function arg type")

		_, err = mf.Compile("{foo, date}")
		require.NoError(t, err)
	})

	t.Run("should apply biDiSupport", func(t *testing.T) {
		mf, err := New("en", &MessageFormatOptions{
			BiDiSupport: true,
		})
		require.NoError(t, err)
		opts := mf.ResolvedOptions()
		assert.True(t, opts.BiDiSupport)
	})

	t.Run("should apply custom formatters", func(t *testing.T) {
		formatters := map[string]any{
			"upper": func(value any, locale string, arg *string) any {
				return strings.ToUpper(value.(string))
			},
		}

		mf, err := New("en", &MessageFormatOptions{
			CustomFormatters: formatters,
		})
		require.NoError(t, err)

		msgFunc, err := mf.Compile("{text, upper}")
		require.NoError(t, err)

		result, err := msgFunc(map[string]any{"text": "hello"})
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)
	})
}

func TestTypeSafeBasics(t *testing.T) {
	t.Run("Basic Creation with Type-Safe Constants", func(t *testing.T) {
		mf, err := New("en", &MessageFormatOptions{
			ReturnType: ReturnTypeString,
			Currency:   "USD",
		})
		require.NoError(t, err)
		require.NotNil(t, mf)

		options := mf.ResolvedOptions()
		assert.Equal(t, ReturnTypeString, options.ReturnType)
		assert.Equal(t, "USD", options.Currency)
		assert.Equal(t, "en", options.Locale)
	})

	t.Run("Values Return Type", func(t *testing.T) {
		mf, err := New("en", &MessageFormatOptions{
			ReturnType:  ReturnTypeValues,
			BiDiSupport: true,
		})
		require.NoError(t, err)

		options := mf.ResolvedOptions()
		assert.Equal(t, ReturnTypeValues, options.ReturnType)
		assert.True(t, options.BiDiSupport)
	})

	t.Run("Static Methods", func(t *testing.T) {
		escaped := Escape("Hello {name}!", true)
		assert.Equal(t, "Hello '{'name'}'!", escaped)

		supported, err := SupportedLocalesOf([]string{"en", "fr", "de"})
		require.NoError(t, err)
		assert.Contains(t, supported, "en")
	})
}

func TestNumberSkeletonTypeSafety(t *testing.T) {
	t.Run("Type-Safe Skeleton Creation", func(t *testing.T) {
		skeleton := &Skeleton{
			Group:        GroupThousands,
			Sign:         SignAlways,
			Decimal:      DecimalAuto,
			RoundingMode: RoundingHalfUp,
			Unit: &UnitConfig{
				Style:    UnitCurrency,
				Currency: new("EUR"),
			},
			Notation: &NotationConfig{
				Style: NotationCompactShort,
			},
			UnitWidth: UnitWidthShort,
		}

		assert.Equal(t, GroupThousands, skeleton.Group)
		assert.Equal(t, SignAlways, skeleton.Sign)
		assert.Equal(t, UnitCurrency, skeleton.Unit.Style)
		assert.Equal(t, "EUR", *skeleton.Unit.Currency)
		assert.Equal(t, NotationCompactShort, skeleton.Notation.Style)
	})

	t.Run("Helper Functions", func(t *testing.T) {
		assert.Equal(t, "EUR", *new("EUR"))
		assert.True(t, *new(true))
		assert.Equal(t, 2, *new(2))
		assert.Equal(t, ReturnTypeValues, *new(ReturnTypeValues))
		assert.Equal(t, SignAlways, *new(SignAlways))
	})
}

func TestMessageExecution(t *testing.T) {
	mf, err := New("en", nil)
	require.NoError(t, err)

	t.Run("simple text", func(t *testing.T) {
		msgFunc, err := mf.Compile("Hello world")
		require.NoError(t, err)

		result, err := msgFunc(nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello world", result)
	})

	t.Run("with variable", func(t *testing.T) {
		msgFunc, err := mf.Compile("Hello {name}")
		require.NoError(t, err)

		result, err := msgFunc(map[string]any{"name": "World"})
		require.NoError(t, err)
		assert.Equal(t, "Hello World", result)
	})

	t.Run("with plural", func(t *testing.T) {
		msgFunc, err := mf.Compile("{count, plural, one{1 item} other{# items}}")
		require.NoError(t, err)

		result, err := msgFunc(map[string]any{"count": 1})
		require.NoError(t, err)
		assert.Equal(t, "1 item", result)

		result, err = msgFunc(map[string]any{"count": 3})
		require.NoError(t, err)
		assert.Equal(t, "3 items", result)
	})

	t.Run("with select", func(t *testing.T) {
		msgFunc, err := mf.Compile("{gender, select, male{He} female{She} other{They}} liked this.")
		require.NoError(t, err)

		result, err := msgFunc(map[string]any{"gender": "male"})
		require.NoError(t, err)
		assert.Equal(t, "He liked this.", result)

		result, err = msgFunc(map[string]any{"gender": "unknown"})
		require.NoError(t, err)
		assert.Equal(t, "They liked this.", result)
	})
}

type octothorpeTestCase struct {
	name     string
	message  string
	params   map[string]any
	expected string
}

func TestOctothorpeReplacement(t *testing.T) {
	mf, err := New("en", nil)
	require.NoError(t, err, "Failed to create MessageFormat")

	tests := []octothorpeTestCase{
		{
			name:     "Basic plural with octothorpe - singular",
			message:  "{count, plural, one {# item} other {# items}}",
			params:   map[string]any{"count": 1},
			expected: "1 item",
		},
		{
			name:     "Basic plural with octothorpe - plural",
			message:  "{count, plural, one {# item} other {# items}}",
			params:   map[string]any{"count": 3},
			expected: "3 items",
		},
		{
			name:     "Multiple octothorpes",
			message:  "{count, plural, one {# item (total: #)} other {# items (total: #)}}",
			params:   map[string]any{"count": 2},
			expected: "2 items (total: 2)",
		},
		{
			name:     "Octothorpe outside plural context",
			message:  "Hash symbol: # and {count, plural, one {# item} other {# items}}",
			params:   map[string]any{"count": 1},
			expected: "Hash symbol: # and 1 item",
		},
		{
			name:     "Selectordinal with octothorpe",
			message:  "{pos, selectordinal, one {#st place} two {#nd place} few {#rd place} other {#th place}}",
			params:   map[string]any{"pos": 3},
			expected: "3rd place",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := mf.Compile(tt.message)
			require.NoError(t, err, "Failed to compile message: %s", tt.message)

			result, err := compiled(tt.params)
			require.NoError(t, err, "Failed to execute compiled message")

			assert.Equal(t, tt.expected, result, "Message: %s, Params: %v", tt.message, tt.params)
		})
	}
}

func TestOctothorpeEdgeCases(t *testing.T) {
	mf, err := New("en", nil)
	require.NoError(t, err)

	tests := []octothorpeTestCase{
		{
			name:     "Zero value",
			message:  "{count, plural, =0 {no items} one {# item} other {# items}}",
			params:   map[string]any{"count": 0},
			expected: "no items",
		},
		{
			name:     "Negative value",
			message:  "{count, plural, one {# item} other {# items}}",
			params:   map[string]any{"count": -2},
			expected: "-2 items",
		},
		{
			name:     "Float value",
			message:  "{count, plural, one {# item} other {# items}}",
			params:   map[string]any{"count": 2.5},
			expected: "2.5 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := mf.Compile(tt.message)
			require.NoError(t, err, "Failed to compile message: %s", tt.message)

			result, err := compiled(tt.params)
			require.NoError(t, err, "Failed to execute compiled message")

			assert.Equal(t, tt.expected, result, "Message: %s, Params: %v", tt.message, tt.params)
		})
	}
}

func TestEscapeFunction(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		octothorpe bool
		expected   string
	}{
		{"Escape hash with octothorpe", "#", true, "'#'"},
		{"Escape hash without octothorpe", "#", false, "#"},
		{"Escape left brace", "{", false, "'{'"},
		{"Escape right brace", "}", false, "'}'"},
		{"Escape complex string", "{test}", false, "'{'test'}'"},
		{"Escape with hash", "{#}", true, "'{''#''}'"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Escape(test.input, test.octothorpe)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRealWorldCompatibility(t *testing.T) {
	t.Run("TypeScript Compatibility Scenarios", func(t *testing.T) {
		mf, err := New("en", nil)
		require.NoError(t, err)

		tests := []struct {
			name     string
			message  string
			params   map[string]any
			expected string
		}{
			{
				name:     "Shopping cart",
				message:  "You have {itemCount, plural, =0 {no items} one {# item} other {# items}} in your cart.",
				params:   map[string]any{"itemCount": 2},
				expected: "You have 2 items in your cart.",
			},
			{
				name:     "Notification count",
				message:  "{count, plural, =0 {No new messages} one {# new message} other {# new messages}}",
				params:   map[string]any{"count": 5},
				expected: "5 new messages",
			},
			{
				name:     "File upload progress",
				message:  "{completed, plural, one {# file uploaded} other {# files uploaded}} of {total}",
				params:   map[string]any{"completed": 3, "total": 10},
				expected: "3 files uploaded of 10",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				compiled, err := mf.Compile(tt.message)
				require.NoError(t, err, "Failed to compile: %s", tt.message)

				result, err := compiled(tt.params)
				require.NoError(t, err, "Failed to execute")

				assert.Equal(t, tt.expected, result)
			})
		}
	})
}
