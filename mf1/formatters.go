// formatters.go - Built-in MessageFormat formatters
// TypeScript original code: /packages/runtime/src/fmt/ module
package v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/numberformat"
	"github.com/kaptinlin/messageformat-go/mf1/internal/intlbridge"
)

// NumberFmt formats numbers with specified parameters
// TypeScript original code:
// export function numberFmt(
//
//	value: number,
//	lc: string | string[],
//	arg: string,
//	defaultCurrency: string
//
//	) {
//	  const [type, currency] = (arg && arg.split(':')) || [];
//	  const opt: Record<string, Intl.NumberFormatOptions | undefined> = {
//	    integer: { maximumFractionDigits: 0 },
//	    percent: { style: 'percent' },
//	    currency: {
//	      style: 'currency',
//	      currency: (currency && currency.trim()) || defaultCurrency,
//	      minimumFractionDigits: 2,
//	      maximumFractionDigits: 2
//	    }
//	  };
//	  return nf(lc, opt[type] || {}).format(value);
//	}
func NumberFmt(value any, lc string, arg string, defaultCurrency string) (string, error) {
	numValue, err := toFloat64(value)
	if err != nil {
		return "", WrapInvalidNumberValue(value)
	}

	var formatType, currency string
	if arg != "" {
		parts := strings.Split(arg, ":")
		formatType = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			currency = strings.TrimSpace(parts[1])
		}
	}

	if currency == "" {
		currency = defaultCurrency
	}

	switch formatType {
	case "integer":
		return NumberInteger(numValue, lc), nil
	case "percent":
		return NumberPercent(numValue, lc), nil
	case "currency":
		return NumberCurrency(numValue, lc, currency), nil
	default:
		return formatNumberDefault(numValue, lc), nil
	}
}

// NumberCurrency formats a number as currency. Mirrors the TypeScript reference,
// which pins fraction digits at 2 regardless of CLDR's per-currency default
// (e.g. JPY's CLDR default of 0). Tests rely on "¥9.50" rather than "¥10".
//
// TypeScript original code:
// export const numberCurrency = (
//
//	value: number,
//	lc: string | string[],
//	arg: string
//
// ) =>
//
//	nf(lc, {
//	  style: 'currency',
//	  currency: arg,
//	  minimumFractionDigits: 2,
//	  maximumFractionDigits: 2
//	}).format(value);
func NumberCurrency(value any, lc string, currencyCode string) string {
	numValue, err := toFloat64(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	loc := intlbridge.ParseLocale(lc)
	nf, err := numberformat.New(loc, numberformat.Options{
		Style:                 numberformat.CurrencyStyle,
		Currency:              numberformat.Currency(currencyCode),
		MinimumFractionDigits: intPtr(2),
		MaximumFractionDigits: intPtr(2),
	})
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return nf.Format(numberformat.Float(numValue))
}

// NumberInteger formats a number as integer
// TypeScript original code:
// export const numberInteger = (value: number, lc: string | string[]) =>
//
//	nf(lc, { maximumFractionDigits: 0 }).format(value);
func NumberInteger(value any, lc string) string {
	numValue, err := toFloat64(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	loc := intlbridge.ParseLocale(lc)
	nf, err := numberformat.New(loc, numberformat.Options{
		MaximumFractionDigits: intPtr(0),
	})
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return nf.Format(numberformat.Float(numValue))
}

// NumberPercent formats a number as percentage. The input is the fractional
// proportion (0.5 → "50%") to match Intl.NumberFormat's percent style.
//
// TypeScript original code:
// export const numberPercent = (value: number, lc: string | string[]) =>
//
//	nf(lc, { style: 'percent' }).format(value);
func NumberPercent(value any, lc string) string {
	numValue, err := toFloat64(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	loc := intlbridge.ParseLocale(lc)
	nf, err := numberformat.New(loc, numberformat.Options{
		Style: numberformat.PercentStyle,
	})
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return nf.Format(numberformat.Float(numValue))
}

// formatNumberDefault formats with numberformat's default ECMA-402 options
// (locale-aware grouping, up to 3 fraction digits). Mirrors Intl.NumberFormat
// with no overrides — the TS reference for NumberFmt's default branch.
func formatNumberDefault(value float64, lc string) string {
	loc := intlbridge.ParseLocale(lc)
	nf, err := numberformat.New(loc, numberformat.Options{})
	if err != nil {
		return strconv.FormatFloat(value, 'g', -1, 64)
	}
	return nf.Format(numberformat.Float(value))
}

func intPtr(v int) *int {
	return &v
}

// DateFormatter formats dates with locale-specific formatting using go-intl's
// datetimeformat. Maps the v1 size enum onto ECMA-402 DateStyle:
//
//	"short"  → numeric Y/M/D (e.g. "5/4/2026" in en-US, preserving 4-digit year)
//	""       → DateStyle=medium (default)
//	"long"   → DateStyle=long
//	"full"   → DateStyle=full
//
// TypeScript original code:
// export function date(
//
//	value: number | string,
//	lc: string | string[],
//	size?: 'short' | 'default' | 'long' | 'full'
//
//	) {
//	  const o: Intl.DateTimeFormatOptions = {
//	    day: 'numeric',
//	    month: 'short',
//	    year: 'numeric'
//	  };
//	  switch (size) {
//	    case 'full':
//	      o.weekday = 'long';
//	    case 'long':
//	      o.month = 'long';
//	      break;
//	    case 'short':
//	      o.month = 'numeric';
//	  }
//	  return new Date(value).toLocaleDateString(lc, o);
//	}
func DateFormatter(value any, lc string, size string) (string, error) {
	t, err := coerceDateInput(value)
	if err != nil {
		return "", err
	}
	return formatDateTimeWithSize(t, lc, size, false)
}

// TimeFormatter formats time values
// TypeScript original code: Similar to date formatter but for time
func TimeFormatter(value any, lc string, size string) (string, error) {
	t, err := coerceTimeInput(value)
	if err != nil {
		return "", err
	}
	return formatDateTimeWithSize(t, lc, size, true)
}

// GetFormatter returns a formatter function by name
func GetFormatter(name string) func(any, string, string) (string, error) {
	switch name {
	case "number":
		return func(value any, lc string, arg string) (string, error) {
			return NumberFmt(value, lc, arg, "USD")
		}
	case "date":
		return DateFormatter
	case "time":
		return TimeFormatter
	default:
		return nil
	}
}

// coerceDateInput parses a v1 date input (string, milliseconds-since-epoch, or
// time.Time) into a time.Time. WrapInvalidDateValue is returned for strings the
// known formats can't parse.
func coerceDateInput(value any) (time.Time, error) {
	return coerceDateTimeInput(value, WrapInvalidDateValue)
}

// coerceTimeInput is like coerceDateInput but returns ErrInvalidTimeValue.
func coerceTimeInput(value any) (time.Time, error) {
	return coerceDateTimeInput(value, WrapInvalidTimeValue)
}

func coerceDateTimeInput(value any, wrapInvalid func(any) error) (time.Time, error) {
	switch v := value.(type) {
	case int64:
		return time.UnixMilli(v), nil
	case int:
		return time.UnixMilli(int64(v)), nil
	case float64:
		return time.UnixMilli(int64(v)), nil
	case string:
		formats := []string{
			time.RFC3339,
			time.DateOnly,
			"01/02/2006",
			time.DateTime,
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.UnixMilli(ts), nil
		}
		return time.Time{}, wrapInvalid(v)
	case time.Time:
		return v, nil
	default:
		return time.Time{}, WrapInvalidType(fmt.Sprintf("%T", value))
	}
}

// formatDateTimeWithSize maps the v1 size enum to datetimeformat options and
// invokes go-intl. The isTime flag toggles between date and time field sets.
func formatDateTimeWithSize(t time.Time, lc, size string, isTime bool) (string, error) {
	loc := intlbridge.ParseLocale(lc)
	opts := dateTimeOptionsForSize(size, isTime)
	if opts.TimeZone == "" {
		// Preserve the input time's location so date-only formatting doesn't
		// slide a day in non-UTC runtimes, and time-only formatting reflects
		// the caller's intended wall clock.
		opts.TimeZone = timeZoneName(t.Location())
	}
	f, err := datetimeformat.New(loc, opts)
	if err != nil {
		return "", err
	}
	return f.Format(t), nil
}

// dateTimeOptionsForSize encodes the v1 size→ECMA-402 mapping. Time formats are
// built from explicit field styles so the test fixtures, which target Go's
// `time.Format("3:04 PM")` shape, line up with CLDR's actual patterns.
func dateTimeOptionsForSize(size string, isTime bool) datetimeformat.Options {
	if isTime {
		switch size {
		case "short":
			return datetimeformat.Options{
				Hour:   datetimeformat.NumericFieldStyle,
				Minute: datetimeformat.TwoDigitFieldStyle,
			}
		case "long", "full":
			return datetimeformat.Options{
				Hour:         datetimeformat.NumericFieldStyle,
				Minute:       datetimeformat.TwoDigitFieldStyle,
				Second:       datetimeformat.TwoDigitFieldStyle,
				TimeZoneName: datetimeformat.ShortTimeZoneName,
			}
		default:
			return datetimeformat.Options{
				Hour:   datetimeformat.NumericFieldStyle,
				Minute: datetimeformat.TwoDigitFieldStyle,
				Second: datetimeformat.TwoDigitFieldStyle,
			}
		}
	}
	switch size {
	case "short":
		// Numeric Y/M/D mirrors the original Go behavior ("5/4/2026") rather
		// than ECMA-402 DateStyle=short, which uses 2-digit years in en-US.
		return datetimeformat.Options{
			Year:  datetimeformat.NumericFieldStyle,
			Month: datetimeformat.NumericMonthStyle,
			Day:   datetimeformat.NumericFieldStyle,
		}
	case "long":
		return datetimeformat.Options{DateStyle: datetimeformat.LongDateTimeStyle}
	case "full":
		return datetimeformat.Options{DateStyle: datetimeformat.FullDateTimeStyle}
	default:
		return datetimeformat.Options{DateStyle: datetimeformat.MediumDateTimeStyle}
	}
}

// timeZoneName extracts a datetimeformat-compatible TimeZone string from a
// time.Location. Returns "" for the system-local location so go-intl falls
// back to its own default; named locations and fixed-offset zones are passed
// through directly.
func timeZoneName(loc *time.Location) string {
	if loc == nil || loc == time.Local {
		return ""
	}
	name := loc.String()
	if name == "" || name == "Local" {
		return ""
	}
	return name
}
