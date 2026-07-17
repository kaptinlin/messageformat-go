package v1_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/kaptinlin/messageformat-go/mf1"
)

// TestCompiledMessageTypedProjections proves one immutable compiled message owns both terminal projections.
// TypeScript original code:
// const message = mf.compile(source); message(values);
func TestCompiledMessageTypedProjections(t *testing.T) {
	t.Parallel()

	options := &v1.MessageFormatOptions{}
	compiler, err := v1.New("en", options)
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

// TestCompiledMessageEvaluatorMatrix proves message shape does not choose hidden evaluation semantics.
// TypeScript original code:
// const message = mf.compile(source); message(values);
func TestCompiledMessageEvaluatorMatrix(t *testing.T) {
	t.Parallel()

	compiler, err := v1.New("en", nil)
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

	compiler, err := v1.New("en", &v1.MessageFormatOptions{RequireAllArguments: true})
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
			require.ErrorIs(t, err, v1.ErrMissingArgument)
			_, err = message.FormatValues(tc.input)
			require.ErrorIs(t, err, v1.ErrMissingArgument)
		})
	}
}

func TestCompileSyntaxErrorOwnsSourceContext(t *testing.T) {
	t.Parallel()

	compiler, err := v1.New("en", &v1.MessageFormatOptions{Strict: true})
	require.NoError(t, err)

	_, err = compiler.Compile("before\n{value, unknown}")
	require.Error(t, err)

	var parseErr *v1.ParseError
	require.ErrorAs(t, err, &parseErr)
	require.NotNil(t, parseErr.Token)
	assert.Equal(t, 2, parseErr.Token.Line)
	assert.Equal(t, 1, parseErr.Token.Col)
	assert.Contains(t, parseErr.Error(), "{value, unknown}")
	assert.Contains(t, parseErr.Error(), "^")
}

func TestCompileRejectsUnmatchedLexerInput(t *testing.T) {
	t.Parallel()

	compiler, err := v1.New("en", nil)
	require.NoError(t, err)

	_, err = compiler.Compile("prefix {name, ,} suffix")
	require.Error(t, err)

	var parseErr *v1.ParseError
	require.ErrorAs(t, err, &parseErr)
	require.NotNil(t, parseErr.Token)
	assert.Equal(t, 12, parseErr.Token.Offset)
	assert.Equal(t, 1, parseErr.Token.Line)
	assert.Equal(t, 13, parseErr.Token.Col)
	assert.Contains(t, parseErr.Error(), "prefix {name, ,} suffix")
}

func TestCompileUnexpectedEndReportsEndPosition(t *testing.T) {
	t.Parallel()

	compiler, err := v1.New("en", nil)
	require.NoError(t, err)

	const source = "first\n{name"
	_, err = compiler.Compile(source)
	require.Error(t, err)

	var parseErr *v1.ParseError
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
			{name: "empty", locale: "", wantErr: v1.ErrInvalidLocale},
			{name: "malformed", locale: "x", wantErr: v1.ErrInvalidLocale},
			{name: "unknown language", locale: "xx", wantErr: v1.ErrInvalidLocale},
			{name: "unknown malformed", locale: "lawlz", wantErr: v1.ErrInvalidLocale},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				mf, err := v1.New(tt.locale, nil)
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
		profile := v1.PluralProfile{
			Locale: "fr",
			Select: func(any, ...bool) (v1.PluralCategory, error) {
				return v1.PluralMany, nil
			},
			Cardinals: []v1.PluralCategory{v1.PluralMany, v1.PluralOther},
			Ordinals:  []v1.PluralCategory{v1.PluralOther},
		}
		mf, err := v1.NewWithPlural(profile, &v1.MessageFormatOptions{
			CustomFormatters: map[string]v1.Formatter{
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

	selector := func(any, ...bool) (v1.PluralCategory, error) {
		return v1.PluralOther, nil
	}
	tests := []struct {
		name    string
		profile v1.PluralProfile
		wantErr error
	}{
		{
			name: "empty locale",
			profile: v1.PluralProfile{
				Select: selector, Cardinals: []v1.PluralCategory{v1.PluralOther},
				Ordinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidLocale,
		},
		{
			name: "malformed locale",
			profile: v1.PluralProfile{
				Locale: "x", Select: selector, Cardinals: []v1.PluralCategory{v1.PluralOther},
				Ordinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidLocale,
		},
		{
			name: "nil selector",
			profile: v1.PluralProfile{
				Locale: "en", Cardinals: []v1.PluralCategory{v1.PluralOther},
				Ordinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralFunction,
		},
		{
			name: "empty cardinals",
			profile: v1.PluralProfile{
				Locale: "en", Select: selector, Ordinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralCategories,
		},
		{
			name: "empty ordinals",
			profile: v1.PluralProfile{
				Locale: "en", Select: selector, Cardinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralCategories,
		},
		{
			name: "duplicate category",
			profile: v1.PluralProfile{
				Locale: "en", Select: selector,
				Cardinals: []v1.PluralCategory{v1.PluralOther, v1.PluralOther},
				Ordinals:  []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralCategories,
		},
		{
			name: "missing other",
			profile: v1.PluralProfile{
				Locale: "en", Select: selector, Cardinals: []v1.PluralCategory{v1.PluralOne},
				Ordinals: []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralCategories,
		},
		{
			name: "unknown category",
			profile: v1.PluralProfile{
				Locale: "en", Select: selector,
				Cardinals: []v1.PluralCategory{"invalid", v1.PluralOther},
				Ordinals:  []v1.PluralCategory{v1.PluralOther},
			},
			wantErr: v1.ErrInvalidPluralCategories,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := v1.NewWithPlural(tc.profile, nil)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestPluralProfileSnapshotsCategories proves construction detaches caller-owned slices.
// TypeScript original code:
// const plural = getPlural(pluralFunction);
func TestPluralProfileSnapshotsCategories(t *testing.T) {
	t.Parallel()

	profile := v1.PluralProfile{
		Locale: "fr",
		Select: func(any, ...bool) (v1.PluralCategory, error) {
			return v1.PluralMany, nil
		},
		Cardinals: []v1.PluralCategory{v1.PluralMany, v1.PluralOther},
		Ordinals:  []v1.PluralCategory{v1.PluralOther},
	}
	compiler, err := v1.NewWithPlural(profile, nil)
	require.NoError(t, err)
	profile.Cardinals[0] = v1.PluralOne
	profile.Ordinals[0] = v1.PluralOne

	resolved := compiler.ResolvedOptions()
	assert.Equal(t, []v1.PluralCategory{v1.PluralMany, v1.PluralOther}, resolved.Plurals[0].Cardinals)
	assert.Equal(t, []v1.PluralCategory{v1.PluralOther}, resolved.Plurals[0].Ordinals)
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
			compiler, err := v1.New(tc.locale, &v1.MessageFormatOptions{Currency: tc.currency})
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

	compiler, err := v1.New("en", nil)
	require.NoError(t, err)
	tests := []struct {
		name    string
		source  string
		value   any
		wantErr error
	}{
		{name: "invalid operand", source: "{value, number}", value: "bad", wantErr: v1.ErrInvalidNumberValue},
		{name: "invalid style", source: "{value, number, decimalish}", value: 1, wantErr: v1.ErrInvalidFormatterStyle},
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
			compiler, err := v1.New("en", &v1.MessageFormatOptions{TimeZone: tc.timeZone})
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

	compiler, err := v1.New("en", nil)
	require.NoError(t, err)
	tests := []struct {
		name    string
		source  string
		value   any
		wantErr error
	}{
		{name: "invalid date operand", source: "{value, date}", value: "bad", wantErr: v1.ErrInvalidDateValue},
		{name: "invalid time operand", source: "{value, time}", value: "bad", wantErr: v1.ErrInvalidTimeValue},
		{name: "invalid date style", source: "{value, date, wide}", value: time.Time{}, wantErr: v1.ErrInvalidFormatterStyle},
		{name: "invalid time style", source: "{value, time, wide}", value: time.Time{}, wantErr: v1.ErrInvalidFormatterStyle},
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
	compiler, err := v1.New("fr", &v1.MessageFormatOptions{
		CustomFormatters: map[string]v1.Formatter{
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

	formatter := v1.Formatter(func(any, string, string) (string, error) { return "", nil })
	tests := []struct {
		name       string
		formatters map[string]v1.Formatter
	}{
		{name: "empty name", formatters: map[string]v1.Formatter{"": formatter}},
		{name: "whitespace name", formatters: map[string]v1.Formatter{"bad name": formatter}},
		{name: "reserved built-in", formatters: map[string]v1.Formatter{"number": formatter}},
		{name: "nil handler", formatters: map[string]v1.Formatter{"missing": nil}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := v1.New("en", &v1.MessageFormatOptions{CustomFormatters: tc.formatters})
			require.ErrorIs(t, err, v1.ErrInvalidFormatter)
		})
	}
}
