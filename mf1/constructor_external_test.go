package mf1_test

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/mf1"
)

var errPluralSelectorTest = errors.New("plural selector failure")

// TestCompiledMessageTypedProjections proves one immutable compiled message owns both terminal projections.
// TypeScript original code:
// const message = mf.compile(source); message(values);
func TestCompiledMessageTypedProjections(t *testing.T) {
	t.Parallel()

	options := &mf1.MessageFormatOptions{}
	compiler, err := mf1.New("en", options)
	require.NoError(t, err)
	options.RequireAllArguments = true

	message, err := compiler.Compile("Hello {name}, {count}")
	require.NoError(t, err)

	text, err := message.Format(map[string]any{"name": "Ada", "count": 3})
	require.NoError(t, err)
	assert.Equal(t, "Hello Ada, 3", text)

	values, err := message.FormatValues(map[string]any{"name": "Ada", "count": 3})
	require.NoError(t, err)
	assert.Equal(t, []any{"Hello ", "Ada", ", ", 3}, values)

	text, err = message.Format(nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello , ", text)
	values, err = message.FormatValues(nil)
	require.NoError(t, err)
	assert.Equal(t, []any{"Hello ", "", ", ", ""}, values)

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			got, formatErr := message.Format(map[string]any{"name": "Ada", "count": 3})
			assert.NoError(t, formatErr)
			assert.Equal(t, "Hello Ada, 3", got)
		})
	}
	wg.Wait()
}

// TestBiDiSupportMarksPlainArguments covers
// .reference/messageformat/mf1/packages/core/src/messages.test.ts, "Bi-directional text".
func TestBiDiSupportMarksPlainArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		mark   string
	}{
		{name: "English uses LRM", locale: "en", mark: "\u200e"},
		{name: "Arabic uses RLM", locale: "ar-EG", mark: "\u200f"},
		{name: "explicit Arabic script uses RLM", locale: "en-Arab", mark: "\u200f"},
		{name: "explicit Latin script uses LRM", locale: "ar-Latn", mark: "\u200e"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			compiler, err := mf1.New(tc.locale, &mf1.MessageFormatOptions{BiDiSupport: true})
			require.NoError(t, err)
			message, err := compiler.Compile("{first} >> {second}")
			require.NoError(t, err)

			values := map[string]any{
				"first":  "Hello! English",
				"second": "Hello هلا؍",
			}
			text, err := message.Format(values)
			require.NoError(t, err)
			assert.Equal(t,
				tc.mark+"Hello! English"+tc.mark+" >> "+tc.mark+"Hello هلا؍"+tc.mark,
				text,
			)

			terminals, err := message.FormatValues(values)
			require.NoError(t, err)
			assert.Equal(t, []any{
				tc.mark + "Hello! English" + tc.mark,
				" >> ",
				tc.mark + "Hello هلا؍" + tc.mark,
			}, terminals)
		})
	}
}

// TestBiDiSupportOnlyMarksPlainArguments covers the token boundary defined by
// .reference/messageformat/mf1/packages/core/src/compiler.ts:139-155.
func TestBiDiSupportOnlyMarksPlainArguments(t *testing.T) {
	t.Parallel()

	const source = "A {plain} B {choice, select, yes {selected} other {other}} C {number, number}"
	values := map[string]any{"plain": "value", "choice": "yes", "number": 12}

	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%t", enabled), func(t *testing.T) {
			t.Parallel()

			compiler, err := mf1.New("en", &mf1.MessageFormatOptions{BiDiSupport: enabled})
			require.NoError(t, err)
			message, err := compiler.Compile(source)
			require.NoError(t, err)

			plain := "value"
			if enabled {
				plain = "\u200evalue\u200e"
			}
			text, err := message.Format(values)
			require.NoError(t, err)
			assert.Equal(t, "A "+plain+" B selected C 12", text)

			terminals, err := message.FormatValues(values)
			require.NoError(t, err)
			assert.Equal(t, []any{"A ", plain, " B ", "selected", " C ", "12"}, terminals)
		})
	}
}

func TestBiDiSupportMissingAndEmptyPlainArguments(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("ar", &mf1.MessageFormatOptions{BiDiSupport: true})
	require.NoError(t, err)
	message, err := compiler.Compile("{value}")
	require.NoError(t, err)

	for _, values := range []map[string]any{nil, {"value": ""}} {
		text, err := message.Format(values)
		require.NoError(t, err)
		assert.Equal(t, "\u200f\u200f", text)

		terminals, err := message.FormatValues(values)
		require.NoError(t, err)
		assert.Equal(t, []any{"\u200f\u200f"}, terminals)
	}

	required, err := mf1.New("ar", &mf1.MessageFormatOptions{
		BiDiSupport:         true,
		RequireAllArguments: true,
	})
	require.NoError(t, err)
	requiredMessage, err := required.Compile("{value}")
	require.NoError(t, err)
	_, err = requiredMessage.Format(nil)
	require.ErrorIs(t, err, mf1.ErrMissingArgument)
	_, err = requiredMessage.FormatValues(nil)
	require.ErrorIs(t, err, mf1.ErrMissingArgument)
}

// TestCompiledMessageEvaluatorMatrix proves message shape does not choose hidden evaluation semantics.
// TypeScript original code:
// const message = mf.compile(source); message(values);
func TestCompiledMessageEvaluatorMatrix(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	tests := []struct {
		name   string
		source string
		input  map[string]any
		text   string
		values []any
	}{
		{
			name:   "interpolation preserves value type",
			source: "Value {value}",
			input:  map[string]any{"value": 7},
			text:   "Value 7",
			values: []any{"Value ", 7},
		},
		{
			name:   "nested select and cardinal",
			source: "{kind, select, use {{count, plural, one {# item} other {# items}}} other {none}}",
			input:  map[string]any{"kind": "use", "count": 2},
			text:   "2 items",
			values: []any{"2", " items"},
		},
		{
			name:   "ordinal",
			source: "{count, selectordinal, one {st} two {nd} few {rd} other {th}}",
			input:  map[string]any{"count": 2},
			text:   "nd",
			values: []any{"nd"},
		},
		{
			name:   "exact match",
			source: "{count, plural, =2 {exact} other {other}}",
			input:  map[string]any{"count": 2},
			text:   "exact",
			values: []any{"exact"},
		},
		{
			name:   "exact match precedes offset",
			source: "{count, plural, offset:1 =2 {exact} one {one} other {other}}",
			input:  map[string]any{"count": 2},
			text:   "exact",
			values: []any{"exact"},
		},
		{
			name:   "offset octothorpe",
			source: "{count, plural, offset:1 one {# item} other {# items}}",
			input:  map[string]any{"count": 3},
			text:   "2 items",
			values: []any{"2", " items"},
		},
		{
			name:   "optional missing argument",
			source: "Hello {name}",
			text:   "Hello ",
			values: []any{"Hello ", ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)

			text, err := message.Format(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.text, text)

			values, err := message.FormatValues(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.values, values)
		})
	}
}

// TestCompiledMessageEvaluatorErrorIdentity proves message shape does not choose an error policy.
// TypeScript original code:
// new MessageFormat('en', { requireAllArguments: true }).compile(source)(values);
func TestCompiledMessageEvaluatorErrorIdentity(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", &mf1.MessageFormatOptions{RequireAllArguments: true})
	require.NoError(t, err)
	tests := []struct {
		name   string
		source string
		input  map[string]any
	}{
		{name: "simple interpolation", source: "Hello {name}"},
		{
			name:   "generic interpolation",
			source: "Hello {name}, count {count, number}",
			input:  map[string]any{"count": 1},
		},
		{name: "basic plural", source: "{count, plural, one {one} other {other}}"},
		{name: "generic plural", source: "Count: {count, plural, one {one} other {other}}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)

			_, err = message.Format(tc.input)
			require.ErrorIs(t, err, mf1.ErrMissingArgument)
			_, err = message.FormatValues(tc.input)
			require.ErrorIs(t, err, mf1.ErrMissingArgument)
		})
	}
}

func TestCompileSyntaxErrorOwnsSourceContext(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", &mf1.MessageFormatOptions{Strict: true})
	require.NoError(t, err)

	_, err = compiler.Compile("before\n{value, unknown}")
	require.Error(t, err)

	var parseErr *mf1.ParseError
	require.ErrorAs(t, err, &parseErr)
	require.NotNil(t, parseErr.Token)
	assert.Equal(t, 2, parseErr.Token.Line)
	assert.Equal(t, 1, parseErr.Token.Col)
	assert.Contains(t, parseErr.Error(), "{value, unknown}")
	assert.Contains(t, parseErr.Error(), "^")
}

func TestCompileRejectsUnmatchedLexerInput(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)

	_, err = compiler.Compile("prefix {name, ,} suffix")
	require.Error(t, err)

	var parseErr *mf1.ParseError
	require.ErrorAs(t, err, &parseErr)
	require.NotNil(t, parseErr.Token)
	assert.Equal(t, 12, parseErr.Token.Offset)
	assert.Equal(t, 1, parseErr.Token.Line)
	assert.Equal(t, 13, parseErr.Token.Col)
	assert.Contains(t, parseErr.Error(), "prefix {name, ,} suffix")
}

func TestCompileUnexpectedEndReportsEndPosition(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)

	const source = "first\n{name"
	_, err = compiler.Compile(source)
	require.Error(t, err)

	var parseErr *mf1.ParseError
	require.ErrorAs(t, err, &parseErr)
	require.NotNil(t, parseErr.Token)
	assert.Equal(t, len(source), parseErr.Token.Offset)
	assert.Equal(t, 2, parseErr.Token.Line)
	assert.Equal(t, 6, parseErr.Token.Col)
	assert.Contains(t, parseErr.Error(), "{name")
	assert.Contains(t, parseErr.Error(), "     ^")
}

func TestTypedConstructors(t *testing.T) {
	t.Parallel()

	t.Run("locale constructor", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			locale     string
			wantLocale string
			wantErr    error
		}{
			{name: "supported", locale: "en", wantLocale: "en"},
			{name: "valid unsupported", locale: "eo", wantLocale: "en"},
			{name: "empty", locale: "", wantErr: mf1.ErrInvalidLocale},
			{name: "malformed", locale: "x", wantErr: mf1.ErrInvalidLocale},
			{name: "unknown language", locale: "xx", wantErr: mf1.ErrInvalidLocale},
			{name: "unknown malformed", locale: "lawlz", wantErr: mf1.ErrInvalidLocale},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				mf, err := mf1.New(tt.locale, nil)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.wantLocale, mf.ResolvedOptions().Locale)
			})
		}
	})

	t.Run("custom plural constructor", func(t *testing.T) {
		t.Parallel()

		var formatterLocale string
		profile := mf1.PluralProfile{
			Locale: "fr",
			Select: func(any, ...bool) (mf1.PluralCategory, error) {
				return mf1.PluralMany, nil
			},
			Cardinals: []mf1.PluralCategory{mf1.PluralMany, mf1.PluralOther},
			Ordinals:  []mf1.PluralCategory{mf1.PluralOther},
		}
		mf, err := mf1.NewWithPlural(profile, &mf1.MessageFormatOptions{
			CustomFormatters: map[string]mf1.Formatter{
				"locale": func(_ any, locale, _ string) (string, error) {
					formatterLocale = locale
					return locale, nil
				},
			},
		})
		require.NoError(t, err)
		resolved := mf.ResolvedOptions()
		assert.Equal(t, "fr", resolved.Locale)
		assert.Equal(t, profile.Cardinals, resolved.Plurals[0].Cardinals)
		assert.Equal(t, profile.Ordinals, resolved.Plurals[0].Ordinals)

		compiled, err := mf.Compile("{count, plural, many {{value, locale}} other {other}}")
		require.NoError(t, err)
		got, err := compiled.Format(map[string]any{"count": 2, "value": "ignored"})
		require.NoError(t, err)
		assert.Equal(t, "fr", got)
		assert.Equal(t, "fr", formatterLocale)

		_, err = mf.Compile("{count, plural, one {one} other {other}}")
		require.Error(t, err)
	})
}

// TestPluralProfileValidation proves invalid plural facts fail at construction.
// TypeScript original code:
// getPlural(pluralFunction)
func TestPluralProfileValidation(t *testing.T) {
	t.Parallel()

	selector := func(any, ...bool) (mf1.PluralCategory, error) {
		return mf1.PluralOther, nil
	}
	tests := []struct {
		name    string
		profile mf1.PluralProfile
		wantErr error
	}{
		{
			name: "empty locale",
			profile: mf1.PluralProfile{
				Select: selector, Cardinals: []mf1.PluralCategory{mf1.PluralOther},
				Ordinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidLocale,
		},
		{
			name: "malformed locale",
			profile: mf1.PluralProfile{
				Locale: "x", Select: selector, Cardinals: []mf1.PluralCategory{mf1.PluralOther},
				Ordinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidLocale,
		},
		{
			name: "nil selector",
			profile: mf1.PluralProfile{
				Locale: "en", Cardinals: []mf1.PluralCategory{mf1.PluralOther},
				Ordinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralFunction,
		},
		{
			name: "empty cardinals",
			profile: mf1.PluralProfile{
				Locale: "en", Select: selector, Ordinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralCategories,
		},
		{
			name: "empty ordinals",
			profile: mf1.PluralProfile{
				Locale: "en", Select: selector, Cardinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralCategories,
		},
		{
			name: "duplicate category",
			profile: mf1.PluralProfile{
				Locale: "en", Select: selector,
				Cardinals: []mf1.PluralCategory{mf1.PluralOther, mf1.PluralOther},
				Ordinals:  []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralCategories,
		},
		{
			name: "missing other",
			profile: mf1.PluralProfile{
				Locale: "en", Select: selector, Cardinals: []mf1.PluralCategory{mf1.PluralOne},
				Ordinals: []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralCategories,
		},
		{
			name: "unknown category",
			profile: mf1.PluralProfile{
				Locale: "en", Select: selector,
				Cardinals: []mf1.PluralCategory{"invalid", mf1.PluralOther},
				Ordinals:  []mf1.PluralCategory{mf1.PluralOther},
			},
			wantErr: mf1.ErrInvalidPluralCategories,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mf1.NewWithPlural(tc.profile, nil)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestPluralProfileSnapshotsCategories proves construction detaches caller-owned slices.
// TypeScript original code:
// const plural = getPlural(pluralFunction);
func TestPluralProfileSnapshotsCategories(t *testing.T) {
	t.Parallel()

	profile := mf1.PluralProfile{
		Locale: "fr",
		Select: func(any, ...bool) (mf1.PluralCategory, error) {
			return mf1.PluralMany, nil
		},
		Cardinals: []mf1.PluralCategory{mf1.PluralMany, mf1.PluralOther},
		Ordinals:  []mf1.PluralCategory{mf1.PluralOther},
	}
	compiler, err := mf1.NewWithPlural(profile, nil)
	require.NoError(t, err)
	profile.Cardinals[0] = mf1.PluralOne
	profile.Ordinals[0] = mf1.PluralOne

	resolved := compiler.ResolvedOptions()
	assert.Equal(t, []mf1.PluralCategory{mf1.PluralMany, mf1.PluralOther}, resolved.Plurals[0].Cardinals)
	assert.Equal(t, []mf1.PluralCategory{mf1.PluralOther}, resolved.Plurals[0].Ordinals)
	message, err := compiler.Compile("{count, plural, many {many} other {other}}")
	require.NoError(t, err)
	text, err := message.Format(map[string]any{"count": 2})
	require.NoError(t, err)
	assert.Equal(t, "many", text)
}

// TestCompiledMessageNumberFormatting proves built-ins use Intl through the public evaluator.
// TypeScript original code:
// mf.compile('{value, number, style}')({ value });
func TestCompiledMessageNumberFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		locale   string
		currency string
		source   string
		value    any
		want     string
	}{
		{name: "en default", locale: "en", source: "{value, number}", value: 12345.67, want: "12,345.67"},
		{name: "fr default", locale: "fr", source: "{value, number}", value: 12345.67, want: "12\u202f345,67"},
		{name: "ar default", locale: "ar-EG", source: "{value, number}", value: 12345.67, want: "١٢٬٣٤٥٫٦٧"},
		{
			name: "plural octothorpe uses locale", locale: "ar-EG",
			source: "{value, plural, other {#}}", value: 12345, want: "١٢٬٣٤٥",
		},
		{name: "integer", locale: "en", source: "{value, number, integer}", value: 12.6, want: "13"},
		{name: "percent", locale: "en", source: "{value, number, percent}", value: 0.25, want: "25%"},
		{
			name: "configured currency", locale: "en", currency: "EUR",
			source: "{value, number, currency}", value: 3, want: "€3.00",
		},
		{
			name: "inline currency", locale: "en", currency: "USD",
			source: "{value, number, currency:GBP}", value: 3, want: "£3.00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiler, err := mf1.New(tc.locale, &mf1.MessageFormatOptions{Currency: tc.currency})
			require.NoError(t, err)
			assert.Equal(t, tc.locale, compiler.ResolvedOptions().Locale)
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)
			got, err := message.Format(map[string]any{"value": tc.value})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestCompiledMessageNumberFormattingErrors proves public formatting preserves typed failures.
// TypeScript original code:
// mf.compile('{value, number, style}')({ value });
func TestCompiledMessageNumberFormattingErrors(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	tests := []struct {
		name    string
		source  string
		value   any
		wantErr error
	}{
		{name: "invalid operand", source: "{value, number}", value: "bad", wantErr: mf1.ErrInvalidNumberValue},
		{name: "invalid style", source: "{value, number, decimalish}", value: 1, wantErr: mf1.ErrInvalidFormatterStyle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)
			_, err = message.Format(map[string]any{"value": tc.value})
			require.ErrorIs(t, err, tc.wantErr)
			_, err = message.FormatValues(map[string]any{"value": tc.value})
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestCompiledPluralRejectsInvalidNumericStrings(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, =0 {zero} other {other}}")
	require.NoError(t, err)

	for _, value := range []string{"not-a-number", "1e10000"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": value})
			assert.Empty(t, formatted)
			require.ErrorIs(t, err, mf1.ErrInvalidNumberStr)

			values, err := message.FormatValues(map[string]any{"value": value})
			assert.Nil(t, values)
			require.ErrorIs(t, err, mf1.ErrInvalidNumberStr)
		})
	}
}

func TestCompiledPluralAcceptsAllIntegerWidths(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, =2 {exact} other {other}}")
	require.NoError(t, err)

	values := []any{
		int(2), int8(2), int16(2), int32(2), int64(2),
		uint(2), uint8(2), uint16(2), uint32(2), uint64(2),
	}
	for _, value := range values {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": value})
			require.NoError(t, err)
			assert.Equal(t, "exact", formatted)
		})
	}
}

func TestCompiledPluralPreservesFractionalOperands(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, one {one #} other {other #}}")
	require.NoError(t, err)

	values := []any{float32(1.5), float64(1.5), "1.5"}
	for _, value := range values {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": value})
			require.NoError(t, err)
			assert.Equal(t, "other 1.5", formatted)
		})
	}
}

func TestCompiledPluralRejectsNonFiniteOperands(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, other {#}}")
	require.NoError(t, err)

	values := []any{math.NaN(), math.Inf(1), math.Inf(-1), "NaN", "+Inf", "-Inf"}
	for _, value := range values {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": value})
			assert.Empty(t, formatted)
			require.ErrorIs(t, err, mf1.ErrInvalidNumberValue)

			values, err := message.FormatValues(map[string]any{"value": value})
			assert.Nil(t, values)
			require.ErrorIs(t, err, mf1.ErrInvalidNumberValue)
		})
	}
}

func TestCompiledPluralRejectsUnsupportedOperandTypes(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, =0 {zero} other {other}}")
	require.NoError(t, err)

	values := []any{nil, true, struct{}{}, []int{1}}
	for _, value := range values {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": value})
			assert.Empty(t, formatted)
			require.ErrorIs(t, err, mf1.ErrInvalidType)

			values, err := message.FormatValues(map[string]any{"value": value})
			assert.Nil(t, values)
			require.ErrorIs(t, err, mf1.ErrInvalidType)
		})
	}
}

func TestCompiledPluralPreservesIntegerBoundaries(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, other {#}}")
	require.NoError(t, err)

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "minimum int64", value: int64(math.MinInt64), want: "-9,223,372,036,854,775,808"},
		{name: "maximum uint64", value: uint64(math.MaxUint64), want: "18,446,744,073,709,551,615"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatted, err := message.Format(map[string]any{"value": tt.value})
			require.NoError(t, err)
			assert.Equal(t, tt.want, formatted)
		})
	}
}

func TestCompiledPluralPreservesLargeIntegerStrings(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "exact selection",
			source: "{value, plural, =9007199254740993 {exact} other {other}}",
		},
		{
			name:   "category selection",
			source: "{value, plural, one {one} other {other}}",
		},
		{
			name:   "offset and octothorpe",
			source: "{value, plural, offset:1 one {# item} other {# items}}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)

			integerText, err := message.Format(map[string]any{"value": int64(9007199254740993)})
			require.NoError(t, err)
			stringText, err := message.Format(map[string]any{"value": "9007199254740993"})
			require.NoError(t, err)
			assert.Equal(t, integerText, stringText)

			integerValues, err := message.FormatValues(map[string]any{"value": int64(9007199254740993)})
			require.NoError(t, err)
			stringValues, err := message.FormatValues(map[string]any{"value": "9007199254740993"})
			require.NoError(t, err)
			assert.Equal(t, integerValues, stringValues)
		})
	}
}

func TestCompiledPluralPropagatesSelectorError(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.NewWithPlural(mf1.PluralProfile{
		Locale: "en",
		Select: func(any, ...bool) (mf1.PluralCategory, error) {
			return mf1.PluralOther, errPluralSelectorTest
		},
		Cardinals: []mf1.PluralCategory{mf1.PluralOther},
		Ordinals:  []mf1.PluralCategory{mf1.PluralOther},
	}, nil)
	require.NoError(t, err)
	message, err := compiler.Compile("{value, plural, other {other}}")
	require.NoError(t, err)

	formatted, err := message.Format(map[string]any{"value": 1})

	assert.Empty(t, formatted)
	require.ErrorIs(t, err, errPluralSelectorTest)

	values, err := message.FormatValues(map[string]any{"value": 1})
	assert.Nil(t, values)
	require.ErrorIs(t, err, errPluralSelectorTest)
}

// TestCompiledMessageDateTimeFormatting proves date/time styles and timezone use Intl publicly.
// TypeScript original code:
// mf.compile('{value, date, style} {value, time, style}')({ value });
func TestCompiledMessageDateTimeFormatting(t *testing.T) {
	t.Parallel()

	instant := time.Date(2026, time.May, 4, 15, 30, 45, 0, time.UTC)
	nearMidnight := time.Date(2026, time.May, 4, 3, 30, 45, 0, time.UTC)
	tests := []struct {
		name     string
		timeZone string
		source   string
		value    time.Time
		want     string
		contains []string
	}{
		{name: "date short", source: "{value, date, short}", value: instant, want: "5/4/2026"},
		{name: "date default", source: "{value, date}", value: instant, want: "May 4, 2026"},
		{name: "date long", source: "{value, date, long}", value: instant, want: "May 4, 2026"},
		{name: "date full", source: "{value, date, full}", value: instant, want: "Monday, May 4, 2026"},
		{name: "time short", source: "{value, time, short}", value: instant, want: "3:30 PM"},
		{name: "time default", source: "{value, time}", value: instant, want: "3:30:45 PM"},
		{name: "time long", source: "{value, time, long}", value: instant, contains: []string{"3:30:45", "GMT"}},
		{name: "time full", source: "{value, time, full}", value: instant, contains: []string{"3:30:45", "GMT"}},
		{
			name: "configured timezone changes date", timeZone: "America/New_York",
			source: "{value, date, short}", value: nearMidnight, want: "5/3/2026",
		},
		{
			name: "configured timezone changes time", timeZone: "America/New_York",
			source: "{value, time, short}", value: nearMidnight, want: "11:30 PM",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiler, err := mf1.New("en", &mf1.MessageFormatOptions{TimeZone: tc.timeZone})
			require.NoError(t, err)
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)
			got, err := message.Format(map[string]any{"value": tc.value})
			require.NoError(t, err)
			if tc.want != "" {
				assert.Equal(t, tc.want, got)
			}
			for _, fragment := range tc.contains {
				assert.Contains(t, got, fragment)
			}
		})
	}
}

// TestCompiledMessageDateTimeFormattingErrors proves public date/time failures stay typed.
// TypeScript original code:
// mf.compile('{value, date, style} {value, time, style}')({ value });
func TestCompiledMessageDateTimeFormattingErrors(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	tests := []struct {
		name    string
		source  string
		value   any
		wantErr error
	}{
		{name: "invalid date operand", source: "{value, date}", value: "bad", wantErr: mf1.ErrInvalidDateValue},
		{name: "invalid time operand", source: "{value, time}", value: "bad", wantErr: mf1.ErrInvalidTimeValue},
		{name: "invalid date style", source: "{value, date, wide}", value: time.Time{}, wantErr: mf1.ErrInvalidFormatterStyle},
		{name: "invalid time style", source: "{value, time, wide}", value: time.Time{}, wantErr: mf1.ErrInvalidFormatterStyle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			message, err := compiler.Compile(tc.source)
			require.NoError(t, err)
			_, err = message.Format(map[string]any{"value": tc.value})
			require.ErrorIs(t, err, tc.wantErr)
			_, err = message.FormatValues(map[string]any{"value": tc.value})
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestCustomFormatterContract proves one typed handler receives public evaluator facts and errors.
// TypeScript original code:
// customFormatters[name](value, locale, arg);
func TestCustomFormatterContract(t *testing.T) {
	t.Parallel()

	formatterErr := errors.New("formatter failed")
	var gotValue any
	var gotLocale, gotStyle string
	compiler, err := mf1.New("fr", &mf1.MessageFormatOptions{
		CustomFormatters: map[string]mf1.Formatter{
			"inspect": func(value any, locale, style string) (string, error) {
				gotValue = value
				gotLocale = locale
				gotStyle = style
				if value == "fail" {
					return "", formatterErr
				}
				return locale + ":" + style + ":" + value.(string), nil
			},
		},
	})
	require.NoError(t, err)
	message, err := compiler.Compile("{value, inspect, upper}")
	require.NoError(t, err)

	text, err := message.Format(map[string]any{"value": "Ada"})
	require.NoError(t, err)
	assert.Equal(t, "fr:upper:Ada", text)
	assert.Equal(t, "Ada", gotValue)
	assert.Equal(t, "fr", gotLocale)
	assert.Equal(t, "upper", gotStyle)

	_, err = message.Format(map[string]any{"value": "fail"})
	require.ErrorIs(t, err, formatterErr)
}

// TestCustomFormatterValidation proves invalid registrations fail at construction.
// TypeScript original code:
// customFormatters?: { [key: string]: CustomFormatter };
func TestCustomFormatterValidation(t *testing.T) {
	t.Parallel()

	formatter := mf1.Formatter(func(any, string, string) (string, error) { return "", nil })
	tests := []struct {
		name       string
		formatters map[string]mf1.Formatter
	}{
		{name: "empty name", formatters: map[string]mf1.Formatter{"": formatter}},
		{name: "whitespace name", formatters: map[string]mf1.Formatter{"bad name": formatter}},
		{name: "reserved built-in", formatters: map[string]mf1.Formatter{"number": formatter}},
		{name: "nil handler", formatters: map[string]mf1.Formatter{"missing": nil}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mf1.New("en", &mf1.MessageFormatOptions{CustomFormatters: tc.formatters})
			require.ErrorIs(t, err, mf1.ErrInvalidFormatter)
		})
	}
}
