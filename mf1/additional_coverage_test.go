package v1

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeHelpers(t *testing.T) {
	t.Parallel()

	t.Run("Number formats locale fallback and offsets", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "999", Number("en", 1000, 1))
		assert.Equal(t, "1.25", Number("bad-locale", 1.25, 0))
	})

	t.Run("StrictNumber accepts numeric inputs and reports non numeric values", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			value  any
			offset float64
			want   string
		}{
			{name: "string", value: "12", offset: 2, want: "10"},
			{name: "float32", value: float32(12.5), offset: 0.5, want: "12"},
			{name: "int8", value: int8(12), offset: 2, want: "10"},
			{name: "int16", value: int16(12), offset: 2, want: "10"},
			{name: "int32", value: int32(12), offset: 2, want: "10"},
			{name: "int64", value: int64(12), offset: 2, want: "10"},
			{name: "uint", value: uint(12), offset: 2, want: "10"},
			{name: "uint8", value: uint8(12), offset: 2, want: "10"},
			{name: "uint16", value: uint16(12), offset: 2, want: "10"},
			{name: "uint32", value: uint32(12), offset: 2, want: "10"},
			{name: "uint64", value: uint64(12), offset: 2, want: "10"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				formatted, err := StrictNumber("en", tc.value, tc.offset, "count")
				require.NoError(t, err)
				assert.Equal(t, tc.want, formatted)
			})
		}

		_, err := StrictNumber("en", "nope", 0, "count")
		require.ErrorIs(t, err, ErrMissingParameter)

		_, err = StrictNumber("en", math.NaN(), 0, "count")
		require.ErrorIs(t, err, ErrMissingParameter)
	})
	t.Run("Plural prefers exact keys and falls back", func(t *testing.T) {
		t.Parallel()

		pluralFn := func(value any, ord ...bool) (PluralCategory, error) {
			if len(ord) > 0 && ord[0] {
				return PluralFew, nil
			}
			if value == float64(1) {
				return PluralOne, nil
			}
			return PluralOther, nil
		}

		data := map[string]any{
			"=2":    "exact",
			"one":   "single",
			"few":   "ordinal few",
			"other": "fallback",
		}
		assert.Equal(t, "exact", Plural(2, 0, pluralFn, data))
		assert.Equal(t, "single", Plural(3, 2, pluralFn, data))
		assert.Equal(t, "ordinal few", Plural(3, 0, pluralFn, data, true))
		assert.Equal(t, "fallback", Plural("1.0", 0, pluralFn, data))
		assert.Equal(t, "fallback", Plural(struct{}{}, 0, pluralFn, data))
	})

	t.Run("Plural falls back when selector errors", func(t *testing.T) {
		t.Parallel()

		pluralFn := func(any, ...bool) (PluralCategory, error) {
			return PluralOther, errors.New("selector failed")
		}

		assert.Equal(t, "fallback", Plural(2, 0, pluralFn, map[string]any{"other": "fallback"}))
		assert.Equal(t, "", Plural(2, 0, pluralFn, map[string]any{}))
	})

	t.Run("SelectValue falls back to other", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "draft", SelectValue("status", map[string]any{"status": "draft", "other": "fallback"}))
		assert.Equal(t, "fallback", SelectValue("missing", map[string]any{"other": "fallback"}))
		assert.Equal(t, "", SelectValue("missing", map[string]any{}))
	})

	t.Run("ReqArgs validates nil and missing maps", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, ReqArgs([]string{"name"}, map[string]any{"name": nil}))
		require.ErrorIs(t, ReqArgs([]string{"name"}, nil), ErrMissingArgument)
		require.ErrorIs(t, ReqArgs([]string{"name"}, map[string]any{}), ErrMissingArgument)
	})

	t.Run("ReplaceOctothorpe only replaces placeholders", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "1 item # literal", ReplaceOctothorpe("__OCTOTHORPE__ item # literal", 2, "en", 1))
		assert.Equal(t, "# unchanged", ReplaceOctothorpe("# unchanged", "bad", "en", 0))
		assert.Equal(t, "3 files", ProcessPluralContent("__OCTOTHORPE__ files", 3, "en", 0))
	})
}

func TestDateSkeletonHelpers(t *testing.T) {
	t.Parallel()

	t.Run("token strings and errors expose readable values", func(t *testing.T) {
		t.Parallel()

		stringToken := &DateTokenString{Value: "literal"}
		fieldToken := &DateTokenField{Char: "y", Width: 2}
		errorToken := &DateTokenError{Error: "bad token"}
		formatErr := NewDateFormatError("invalid", "bad date skeleton", fieldToken)

		assert.Equal(t, "literal", stringToken.String())
		assert.Equal(t, "{char: y, width: 2}", fieldToken.String())
		assert.Equal(t, "{error: bad token}", errorToken.String())
		assert.Equal(t, "DateFormat invalid: bad date skeleton", formatErr.Error())
		assert.Same(t, fieldToken, formatErr.Token)
	})

	t.Run("ParseDateTokens handles fields literals and quote errors", func(t *testing.T) {
		t.Parallel()

		tokens := ParseDateTokens("yyyy-'week''s'-MM")
		require.Len(t, tokens, 5)

		field, ok := tokens[0].(*DateTokenField)
		require.True(t, ok)
		assert.Equal(t, "y", field.Char)
		assert.Equal(t, 4, field.Width)

		literal, ok := tokens[2].(*DateTokenString)
		require.True(t, ok)
		assert.Equal(t, "week's", literal.Value)

		unterminated := ParseDateTokens("yyyy-'bad")
		require.Len(t, unterminated, 3)
		_, ok = unterminated[2].(*DateTokenError)
		assert.True(t, ok)
	})

	t.Run("GetDateTimeFormatOptions maps fields and reports unsupported tokens", func(t *testing.T) {
		t.Parallel()

		var errs []string
		tokens := ParseDateTokens("GGGG yy MMMMM dd EEEE HH mm ss SSS zzzz u XXXXXX")
		options := GetDateTimeFormatOptions(tokens, func(errorType, message string, token DateToken) {
			errs = append(errs, errorType+":"+message+":"+fmt.Sprint(token))
		})

		assert.Equal(t, "long", options.Era)
		assert.Equal(t, "2-digit", options.Year)
		assert.Equal(t, "narrow", options.Month)
		assert.Equal(t, "2-digit", options.Day)
		assert.Equal(t, "long", options.Weekday)
		assert.Equal(t, "2-digit", options.Hour)
		assert.Equal(t, "h23", options.HourCycle)
		assert.Equal(t, "2-digit", options.Minute)
		assert.Equal(t, "2-digit", options.Second)
		assert.Equal(t, "3", options.FractionalSecond)
		assert.Equal(t, "long", options.TimeZoneName)
		assert.Equal(t, "gregory", options.Calendar)
		assert.NotEmpty(t, errs)
	})

	t.Run("GetDateTimeFormatOptions maps alternate widths and invalid callbacks", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			token DateToken
			want  DateTimeFormatOptions
		}{
			{name: "era short", token: &DateTokenField{Char: "G", Width: 3}, want: DateTimeFormatOptions{Era: "short"}},
			{name: "era narrow", token: &DateTokenField{Char: "G", Width: 5}, want: DateTimeFormatOptions{Era: "narrow"}},
			{name: "month numeric", token: &DateTokenField{Char: "M", Width: 1}, want: DateTimeFormatOptions{Month: "numeric"}},
			{name: "month two digit", token: &DateTokenField{Char: "M", Width: 2}, want: DateTimeFormatOptions{Month: "2-digit"}},
			{name: "month short", token: &DateTokenField{Char: "L", Width: 3}, want: DateTimeFormatOptions{Month: "short"}},
			{name: "month long", token: &DateTokenField{Char: "L", Width: 4}, want: DateTimeFormatOptions{Month: "long"}},
			{name: "weekday short", token: &DateTokenField{Char: "E", Width: 3}, want: DateTimeFormatOptions{Weekday: "short"}},
			{name: "weekday narrow", token: &DateTokenField{Char: "c", Width: 6}, want: DateTimeFormatOptions{Weekday: "narrow"}},
			{name: "hour h12", token: &DateTokenField{Char: "h", Width: 1}, want: DateTimeFormatOptions{Hour: "numeric", HourCycle: "h12"}},
			{name: "hour h24", token: &DateTokenField{Char: "k", Width: 2}, want: DateTimeFormatOptions{Hour: "2-digit", HourCycle: "h24"}},
			{name: "hour h11", token: &DateTokenField{Char: "K", Width: 1}, want: DateTimeFormatOptions{Hour: "numeric", HourCycle: "h11"}},
			{name: "minute numeric", token: &DateTokenField{Char: "m", Width: 1}, want: DateTimeFormatOptions{Minute: "numeric"}},
			{name: "second numeric", token: &DateTokenField{Char: "s", Width: 1}, want: DateTimeFormatOptions{Second: "numeric"}},
			{name: "timezone short", token: &DateTokenField{Char: "z", Width: 3}, want: DateTimeFormatOptions{TimeZoneName: "short"}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got := GetDateTimeFormatOptions([]DateToken{tc.token}, nil)
				if diff := cmp.Diff(tc.want, *got); diff != "" {
					t.Fatalf("GetDateTimeFormatOptions() mismatch (-want +got):\n%s", diff)
				}
			})
		}

		var gotTypes []string
		GetDateTimeFormatOptions([]DateToken{
			&DateTokenField{Char: "G", Width: 6},
			&DateTokenField{Char: "M", Width: 6},
			&DateTokenField{Char: "E", Width: 7},
			&DateTokenField{Char: "S", Width: 4},
			&DateTokenField{Char: "z", Width: 5},
			&DateTokenField{Char: "Q", Width: 1},
			&DateTokenError{Error: "bad token"},
		}, func(errorType, _ string, _ DateToken) {
			gotTypes = append(gotTypes, errorType)
		})
		wantTypes := []string{"invalid", "invalid", "invalid", "invalid", "invalid", "unsupported", "invalid"}
		if diff := cmp.Diff(wantTypes, gotTypes); diff != "" {
			t.Fatalf("date option error types mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("date formatter helpers produce formatter and source", func(t *testing.T) {
		t.Parallel()

		var errs []*DateFormatError
		formatter, err := GetDateFormatter("en", "yyyy-MM-dd", "UTC", func(err *DateFormatError) {
			errs = append(errs, err)
		})
		require.NoError(t, err)
		formatted, err := formatter(time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.Contains(t, formatted, "DateFormatter[yyyy-MM-dd]")
		assert.Empty(t, errs)

		source, err := GetDateFormatterSource([]string{"en", "fr"}, "yyyy-MM-dd z", "UTC", nil)
		require.NoError(t, err)
		assert.Contains(t, source, `"year":"numeric"`)
		assert.Contains(t, source, `"timeZone":"UTC"`)
		assert.Contains(t, source, `["en","fr"]`)

		_, err = GetDateFormatterSource(42, "yyyy", "", nil)
		require.ErrorIs(t, err, ErrInvalidType)
	})
}

func TestNumberSkeletonHelpers(t *testing.T) {
	t.Parallel()

	t.Run("NumberSkeletonError formats details", func(t *testing.T) {
		t.Parallel()

		err := &NumberSkeletonError{Type: "BadStem", Message: "unknown", Stem: "bad"}
		assert.Equal(t, "NumberSkeleton BadStem: unknown", err.Error())
	})

	t.Run("ParseNumberSkeleton maps common tokens", func(t *testing.T) {
		t.Parallel()

		skeleton, err := ParseNumberSkeleton(strings.Join([]string{
			"engineering/+ee/sign-always",
			"currency/USD",
			"precision-increment/5/stripIfInteger",
			"sign-accounting-except-zero",
			"group-thousands",
			"decimal-always",
			"unit-width-full-name",
			"rounding-mode-half-up",
			"scale/2",
			"integer-width/2/4",
			"numbering-system/latn",
		}, " "))
		require.NoError(t, err)

		require.NotNil(t, skeleton.Notation)
		assert.Equal(t, NotationEngineering, skeleton.Notation.Style)
		require.NotNil(t, skeleton.Notation.ExpDigits)
		assert.Equal(t, 2, *skeleton.Notation.ExpDigits)
		assert.Equal(t, SignAlways, skeleton.Notation.ExpSign)
		require.NotNil(t, skeleton.Unit)
		assert.Equal(t, UnitCurrency, skeleton.Unit.Style)
		require.NotNil(t, skeleton.Unit.Currency)
		assert.Equal(t, "USD", *skeleton.Unit.Currency)
		require.NotNil(t, skeleton.Precision)
		assert.Equal(t, PrecisionIncrement, skeleton.Precision.Style)
		require.NotNil(t, skeleton.Precision.Increment)
		assert.Equal(t, 5, *skeleton.Precision.Increment)
		assert.Equal(t, TrailingZeroStripIfInteger, skeleton.Precision.TrailingZero)
		assert.Equal(t, SignAccountingExceptZero, skeleton.Sign)
		assert.Equal(t, GroupThousands, skeleton.Group)
		assert.Equal(t, DecimalAlways, skeleton.Decimal)
		assert.Equal(t, UnitWidthFullName, skeleton.UnitWidth)
		assert.Equal(t, RoundingHalfUp, skeleton.RoundingMode)
		require.NotNil(t, skeleton.Scale)
		assert.Equal(t, 2, *skeleton.Scale)
		require.NotNil(t, skeleton.IntegerWidth)
		assert.Equal(t, 2, skeleton.IntegerWidth.Min)
		require.NotNil(t, skeleton.IntegerWidth.Max)
		assert.Equal(t, 4, *skeleton.IntegerWidth.Max)
		require.NotNil(t, skeleton.NumberingSystem)
		assert.Equal(t, "latn", *skeleton.NumberingSystem)
	})

	t.Run("TokenParser covers alternate token families", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			stem   string
			option []string
			check  func(*testing.T, Skeleton)
		}{
			{name: "compact short alias", stem: "K", check: func(t *testing.T, s Skeleton) { assert.Equal(t, NotationCompactShort, s.Notation.Style) }},
			{name: "compact long alias", stem: "KK", check: func(t *testing.T, s Skeleton) { assert.Equal(t, NotationCompactLong, s.Notation.Style) }},
			{name: "notation simple", stem: "notation-simple", check: func(t *testing.T, s Skeleton) { assert.Equal(t, NotationSimple, s.Notation.Style) }},
			{name: "percent", stem: "percent", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitPercent, s.Unit.Style) }},
			{name: "permille", stem: "permille", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitPermille, s.Unit.Style) }},
			{name: "base unit", stem: "base-unit", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitBaseUnit, s.Unit.Style) }},
			{name: "measure unit", stem: "measure-unit", option: []string{"meter"}, check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitMeasureUnit, s.Unit.Style) }},
			{name: "concise unit", stem: "concise-unit", option: []string{"meter"}, check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitConciseUnit, s.Unit.Style) }},
			{name: "precision integer", stem: "precision-integer", option: []string{"auto"}, check: func(t *testing.T, s Skeleton) { assert.Equal(t, PrecisionInteger, s.Precision.Style) }},
			{name: "precision unlimited", stem: "precision-unlimited", option: []string{"stripIfInteger"}, check: func(t *testing.T, s Skeleton) { assert.Equal(t, PrecisionUnlimited, s.Precision.Style) }},
			{name: "precision currency standard", stem: "precision-currency-standard", check: func(t *testing.T, s Skeleton) { assert.Equal(t, PrecisionCurrencyStandard, s.Precision.Style) }},
			{name: "precision currency cash", stem: "precision-currency-cash", check: func(t *testing.T, s Skeleton) { assert.Equal(t, PrecisionCurrencyCash, s.Precision.Style) }},
			{name: "sign negative", stem: "sign-negative", check: func(t *testing.T, s Skeleton) { assert.Equal(t, SignNegative, s.Sign) }},
			{name: "group off", stem: "group-off", check: func(t *testing.T, s Skeleton) { assert.Equal(t, GroupOff, s.Group) }},
			{name: "decimal auto", stem: "decimal-auto", check: func(t *testing.T, s Skeleton) { assert.Equal(t, DecimalAuto, s.Decimal) }},
			{name: "unit width hidden", stem: "unit-width-hidden", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitWidthHidden, s.UnitWidth) }},
			{name: "rounding ceiling", stem: "rounding-mode-ceiling", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingCeiling, s.RoundingMode) }},
			{name: "scientific sign never", stem: "scientific", option: []string{"+e", "sign-never"}, check: func(t *testing.T, s Skeleton) {
				assert.Equal(t, NotationScientific, s.Notation.Style)
				require.NotNil(t, s.Notation.ExpDigits)
				assert.Equal(t, 1, *s.Notation.ExpDigits)
				assert.Equal(t, SignNever, s.Notation.ExpSign)
			}},
			{name: "sign accounting", stem: "sign-accounting", check: func(t *testing.T, s Skeleton) { assert.Equal(t, SignAccounting, s.Sign) }},
			{name: "sign accounting always", stem: "sign-accounting-always", check: func(t *testing.T, s Skeleton) { assert.Equal(t, SignAccountingAlways, s.Sign) }},
			{name: "sign except zero", stem: "sign-except-zero", check: func(t *testing.T, s Skeleton) { assert.Equal(t, SignExceptZero, s.Sign) }},
			{name: "sign accounting negative", stem: "sign-accounting-negative", check: func(t *testing.T, s Skeleton) { assert.Equal(t, SignAccountingNegative, s.Sign) }},
			{name: "group min2", stem: "group-min2", check: func(t *testing.T, s Skeleton) { assert.Equal(t, GroupMin2, s.Group) }},
			{name: "group auto", stem: "group-auto", check: func(t *testing.T, s Skeleton) { assert.Equal(t, GroupAuto, s.Group) }},
			{name: "group aligned", stem: "group-on-aligned", check: func(t *testing.T, s Skeleton) { assert.Equal(t, GroupOnAligned, s.Group) }},
			{name: "unit width narrow", stem: "unit-width-narrow", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitWidthNarrow, s.UnitWidth) }},
			{name: "unit width short", stem: "unit-width-short", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitWidthShort, s.UnitWidth) }},
			{name: "unit width iso", stem: "unit-width-iso-code", check: func(t *testing.T, s Skeleton) { assert.Equal(t, UnitWidthIsoCode, s.UnitWidth) }},
			{name: "rounding floor", stem: "rounding-mode-floor", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingFloor, s.RoundingMode) }},
			{name: "rounding down", stem: "rounding-mode-down", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingDown, s.RoundingMode) }},
			{name: "rounding up", stem: "rounding-mode-up", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingUp, s.RoundingMode) }},
			{name: "rounding half even", stem: "rounding-mode-half-even", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingHalfEven, s.RoundingMode) }},
			{name: "rounding half odd", stem: "rounding-mode-half-odd", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingHalfOdd, s.RoundingMode) }},
			{name: "rounding half ceiling", stem: "rounding-mode-half-ceiling", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingHalfCeiling, s.RoundingMode) }},
			{name: "rounding half floor", stem: "rounding-mode-half-floor", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingHalfFloor, s.RoundingMode) }},
			{name: "rounding half down", stem: "rounding-mode-half-down", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingHalfDown, s.RoundingMode) }},
			{name: "rounding unnecessary", stem: "rounding-mode-unnecessary", check: func(t *testing.T, s Skeleton) { assert.Equal(t, RoundingUnnecessary, s.RoundingMode) }},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				parser := NewTokenParser(func(error) { t.Fatal("unexpected skeleton error") })
				parser.ParseToken(tc.stem, tc.option)
				tc.check(t, parser.Skeleton())
			})
		}
	})

	t.Run("ParseNumberSkeleton reports validation errors", func(t *testing.T) {
		t.Parallel()

		tests := []string{
			"unknown-token",
			"compact-short/unexpected",
			"currency",
			"currency/USD/EUR",
			"precision-increment/not-number",
			"scale/not-number",
			"integer-width/not-number",
			"integer-width/1/not-number",
			"scientific/bad-option",
		}

		for _, src := range tests {
			t.Run(src, func(t *testing.T) {
				t.Parallel()

				_, err := ParseNumberSkeleton(src)
				require.Error(t, err)
			})
		}
	})
}

func TestLexerAndParserHelpers(t *testing.T) {
	t.Parallel()

	t.Run("Lexer reset tokenize and iterator expose tokens", func(t *testing.T) {
		t.Parallel()

		lexer := NewLexer("Hello {name} '')")
		lexer.Reset("Hi {name, number} {count, plural, offset:1 one {# item} other {# items}}")
		tokens, err := lexer.Tokenize()
		require.NoError(t, err)
		require.NotEmpty(t, tokens)

		var tokenTypes []string
		iter, err := lexer.Iterator()
		require.NoError(t, err)
		for token := iter(); token != nil; token = iter() {
			tokenTypes = append(tokenTypes, token.Type)
		}
		assert.Contains(t, tokenTypes, TokenArgument)
		assert.Contains(t, tokenTypes, TokenFuncSimple)
		assert.Contains(t, tokenTypes, TokenSelect)
		assert.Contains(t, tokenTypes, TokenOffset)
		assert.Contains(t, tokenTypes, TokenCase)
		assert.Contains(t, tokenTypes, TokenOctothorpe)

		assert.Contains(t, lexer.FormatError(nil, "bad"), "ParseError: bad")
		assert.Contains(t, lexer.FormatError(&tokens[0], "bad"), "line 1 col 1")
	})

	t.Run("token accessors return type and context", func(t *testing.T) {
		t.Parallel()

		ctx := Context{Offset: 1, Line: 2, Col: 3, Text: "x", LineBreaks: 0}
		tokens := []Token{
			&Content{Type: "content", Value: "hello", Ctx: ctx},
			&PlainArg{Type: "argument", Arg: "name", Ctx: ctx},
			&FunctionArg{Type: "function", Arg: "value", Key: "number", Ctx: ctx},
			&Select{Type: "plural", Arg: "count", Ctx: ctx},
			&Octothorpe{Type: "octothorpe", Ctx: ctx},
		}

		for _, token := range tokens {
			require.NotEmpty(t, token.GetType())
			if diff := cmp.Diff(ctx, token.GetContext()); diff != "" {
				t.Fatalf("GetContext() mismatch (-want +got):\n%s", diff)
			}
		}
	})

	t.Run("Parse returns typed AST and strict errors", func(t *testing.T) {
		t.Parallel()

		tokens, err := Parse("Hello {name}!", nil)
		require.NoError(t, err)
		require.Len(t, tokens, 3)
		_, ok := tokens[1].(*PlainArg)
		assert.True(t, ok)

		strict := true
		_, err = Parse("{value, unknown}", &ParseOptions{Strict: true, StrictPluralKeys: &strict})
		require.Error(t, err)
		parseErr, ok := errors.AsType[*ParseError](err)
		require.True(t, ok)
		assert.Equal(t, "Invalid strict mode function arg type: unknown", parseErr.Message)
		require.NotNil(t, parseErr.Token)
		assert.Equal(t, TokenArgument, parseErr.Token.Type)

		_, err = Parse("{count, plural, invalid {bad} other {ok}}", nil)
		require.Error(t, err)
		parseErr, ok = errors.AsType[*ParseError](err)
		require.True(t, ok)
		assert.Equal(t, "The plural case invalid is not valid in this locale", parseErr.Message)
		require.NotNil(t, parseErr.Token)
		assert.Equal(t, TokenCase, parseErr.Token.Type)
	})
}

func TestMessageFormatPublicBehavior(t *testing.T) {
	t.Parallel()

	t.Run("SupportedLocalesOf preserves supported input order", func(t *testing.T) {
		t.Parallel()

		got, err := SupportedLocalesOf([]string{"en", "eo", "pt-BR"})
		require.NoError(t, err)
		want := []string{"en", "pt-BR"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("SupportedLocalesOf() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("custom plural function drives compiled plural selection", func(t *testing.T) {
		t.Parallel()

		profile := PluralProfile{
			Locale: "en",
			Select: func(any, ...bool) (PluralCategory, error) {
				return PluralMany, nil
			},
			Cardinals: []PluralCategory{PluralMany, PluralOther},
			Ordinals:  []PluralCategory{PluralOther},
		}
		mf, err := NewWithPlural(profile, nil)
		require.NoError(t, err)
		assert.Equal(t, "en", mf.ResolvedOptions().Locale)

		msg, err := mf.Compile("{count, plural, many {many} other {other}}")
		require.NoError(t, err)
		got, err := msg.Format(map[string]any{"count": 2})
		require.NoError(t, err)
		assert.Equal(t, "many", got)
	})

	t.Run("FormatValues preserves message parts", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)

		simple, err := mf.Compile("Hello {name}!")
		require.NoError(t, err)
		gotParts, err := simple.FormatValues(map[string]any{"name": "Ada"})
		require.NoError(t, err)
		want := []any{"Hello ", "Ada", "!"}
		if diff := cmp.Diff(want, gotParts); diff != "" {
			t.Fatalf("simple values mismatch (-want +got):\n%s", diff)
		}

		plural, err := mf.Compile("{count, plural, one {# item} other {# items}}")
		require.NoError(t, err)
		gotParts, err = plural.FormatValues(map[string]any{"count": 3})
		require.NoError(t, err)
		want = []any{"3", " items"}
		if diff := cmp.Diff(want, gotParts); diff != "" {
			t.Fatalf("plural values mismatch (-want +got):\n%s", diff)
		}

		standard, err := mf.Compile("Report: {count, number} for {name}")
		require.NoError(t, err)
		gotParts, err = standard.FormatValues(map[string]any{"count": 7, "name": "Ada"})
		require.NoError(t, err)
		want = []any{"Report: ", "7", " for ", "Ada"}
		if diff := cmp.Diff(want, gotParts); diff != "" {
			t.Fatalf("standard values mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("compiled messages apply one missing argument policy", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)

		simple, err := mf.Compile("Hello {name}")
		require.NoError(t, err)
		text, err := simple.Format(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, "Hello ", text)

		plural, err := mf.Compile("{count, plural, one {one} other {other}}")
		require.NoError(t, err)
		text, err = plural.Format(nil)
		require.NoError(t, err)
		assert.Equal(t, "other", text)
		text, err = plural.Format(map[string]any{})
		require.NoError(t, err)
		assert.Equal(t, "other", text)

		required, err := New("en", &MessageFormatOptions{RequireAllArguments: true})
		require.NoError(t, err)
		standard, err := required.Compile("{name} scored {points, number}")
		require.NoError(t, err)
		_, err = standard.Format(map[string]any{"points": 9})
		require.ErrorIs(t, err, ErrMissingArgument)
	})

	t.Run("custom formatter receives locale style and value", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en-US", &MessageFormatOptions{CustomFormatters: map[string]Formatter{
			"tag": func(value any, locale, style string) (string, error) {
				return fmt.Sprintf("%s:%s:%v", locale, style, value), nil
			},
		}})
		require.NoError(t, err)

		msg, err := mf.Compile("{text, tag, label}")
		require.NoError(t, err)
		got, err := msg.Format(map[string]any{"text": "go"})
		require.NoError(t, err)
		assert.Equal(t, "en-US:label:go", got)
	})

	t.Run("relaxed plural keys allow non CLDR case labels", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", &MessageFormatOptions{StrictPluralKeys: PluralKeyModeRelaxed})
		require.NoError(t, err)
		msg, err := mf.Compile("{count, plural, invalid {bad} other {ok}}")
		require.NoError(t, err)
		got, err := msg.Format(map[string]any{"count": 1})
		require.NoError(t, err)
		assert.Equal(t, "ok", got)
	})
}

func TestMessagesPublicBehavior(t *testing.T) {
	t.Parallel()

	compiler, err := New("en", nil)
	require.NoError(t, err)
	dynamic, err := compiler.Compile("Hi {name}")
	require.NoError(t, err)

	messages := NewMessages(map[string]MessageData{
		"en": {
			"7":       "lucky",
			"dynamic": dynamic,
			"object":  map[string]any{"title": "Nested"},
		},
	}, "en")

	assert.True(t, messages.HasMessage("dynamic", nil))
	got, err := messages.Get("dynamic", map[string]any{"name": "Lin"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hi Lin", got)

	got, err = messages.Get(7, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "lucky", got)

	assert.True(t, messages.HasObject("object", nil))
	got, err = messages.Get([]string{"object", "title"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Nested", got)
}

func TestMessageFormatAdditionalBehavior(t *testing.T) {
	t.Parallel()

	t.Run("standard execution handles typed parameters and optional missing args", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)
		msg, err := mf.Compile("{name, number} / {missing}")
		require.NoError(t, err)

		got, err := msg.Format(map[string]any{"name": 7})
		require.NoError(t, err)
		assert.Equal(t, "7 / ", got)
	})

	t.Run("invalid octothorpe values fall back to literal hash", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)
		msg, err := mf.Compile("prefix {kind, select, use {{count, plural, other {# items}}} other {none}}")
		require.NoError(t, err)

		got, err := msg.Format(map[string]any{"kind": "use", "count": "not-a-number"})
		require.NoError(t, err)
		assert.Equal(t, "prefix # items", got)
	})

	t.Run("selector reports missing other case", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)
		msg, err := mf.Compile("prefix {kind, select, known {Known}}")
		require.NoError(t, err)

		_, err = msg.Format(map[string]any{"kind": "unknown"})
		require.ErrorIs(t, err, ErrNoMatchingCase)

		_, err = msg.Format(map[string]any{})
		require.ErrorIs(t, err, ErrNoOtherCase)
	})

	t.Run("strict function arguments collapse to formatted content", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", &MessageFormatOptions{Strict: true, CustomFormatters: map[string]Formatter{
			"spellout": func(value any, locale, style string) (string, error) {
				require.NotEmpty(t, style)
				return fmt.Sprintf("%s:%s:%v", locale, style, value), nil
			},
		}})
		require.NoError(t, err)
		msg, err := mf.Compile("{value, spellout, :: currency/USD}")
		require.NoError(t, err)

		got, err := msg.Format(map[string]any{"value": 12})
		require.NoError(t, err)
		assert.Equal(t, "en::: currency/USD:12", got)
	})

	t.Run("number formatter handles supported numeric types", func(t *testing.T) {
		t.Parallel()

		mf, err := New("en", nil)
		require.NoError(t, err)
		msg, err := mf.Compile("{count, plural, one {# item} other {# items}}")
		require.NoError(t, err)

		for _, value := range []any{int64(2), float32(2), "2"} {
			t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
				t.Parallel()

				got, err := msg.Format(map[string]any{"count": value})
				require.NoError(t, err)
				assert.Equal(t, "2 items", got)
			})
		}
	})
}

func TestPluralHelpers(t *testing.T) {
	t.Parallel()

	t.Run("HasPlural accepts supported locales", func(t *testing.T) {
		t.Parallel()

		assert.True(t, HasPlural("en-US"))
		assert.True(t, HasPlural("pt-PT"))
		assert.False(t, HasPlural("x"))
	})

	t.Run("GetPlural resolves string custom and invalid inputs", func(t *testing.T) {
		t.Parallel()

		english, err := GetPlural("en-US")
		require.NoError(t, err)
		assert.Equal(t, "en-US", english.Locale)

		tests := []struct {
			name  string
			value any
			want  PluralCategory
		}{
			{name: "int", value: 1, want: PluralOne},
			{name: "int32", value: int32(1), want: PluralOne},
			{name: "int64", value: int64(2), want: PluralOther},
			{name: "float32", value: float32(1.9), want: PluralOne},
			{name: "float64", value: 2.1, want: PluralOther},
			{name: "string", value: "1", want: PluralOne},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				category, err := english.Func(tc.value)
				require.NoError(t, err)
				assert.Equal(t, tc.want, category)
			})
		}
		category, err := english.Func(2, true)
		require.NoError(t, err)
		assert.NotEmpty(t, category)

		_, err = english.Func("bad")
		require.ErrorIs(t, err, ErrInvalidNumberStr)
		_, err = english.Func(struct{}{})
		require.ErrorIs(t, err, ErrInvalidType)
	})
}

func TestErrorWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "invalid locale", err: WrapInvalidLocale("x"), want: ErrInvalidLocale},
		{name: "invalid number value", err: WrapInvalidNumberValue("x"), want: ErrInvalidNumberValue},
		{name: "invalid date value", err: WrapInvalidDateValue("x"), want: ErrInvalidDateValue},
		{name: "invalid time value", err: WrapInvalidTimeValue("x"), want: ErrInvalidTimeValue},
		{name: "invalid type", err: WrapInvalidType("struct"), want: ErrInvalidType},
		{name: "invalid param type", err: WrapInvalidParamType("struct"), want: ErrInvalidParamType},
		{name: "missing parameter", err: WrapMissingParameter("name"), want: ErrMissingParameter},
		{name: "missing argument", err: WrapMissingArgument("name"), want: ErrMissingArgument},
		{name: "no matching case", err: WrapNoMatchingCase("count", "plural"), want: ErrNoMatchingCase},
		{name: "invalid number string", err: WrapInvalidNumberStr("x"), want: ErrInvalidNumberStr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tc.err, tc.want)
		})
	}
}
