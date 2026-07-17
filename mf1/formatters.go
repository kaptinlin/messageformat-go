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

// formatNumber formats one numeric value with the closed MF1 style vocabulary.
// TypeScript original code:
// return nf(lc, options).format(value);
func formatNumber(value any, locale, style, defaultCurrency string) (string, error) {
	number, err := toFloat64(value)
	if err != nil {
		return "", WrapInvalidNumberValue(value)
	}

	styleName, currency, hasCurrency := strings.Cut(strings.TrimSpace(style), ":")
	styleName = strings.TrimSpace(styleName)
	currency = strings.TrimSpace(currency)
	if hasCurrency && (styleName != "currency" || strings.Contains(currency, ":")) {
		return "", WrapInvalidFormatterStyle("number", style)
	}

	var options numberformat.Options
	switch styleName {
	case "":
	case "integer":
		options.MaximumFractionDigits = intPtr(0)
	case "percent":
		options.Style = stringPtr(string(numberformat.PercentStyle))
	case "currency":
		if currency == "" {
			currency = defaultCurrency
		}
		options.Style = stringPtr(string(numberformat.CurrencyStyle))
		options.Currency = stringPtr(currency)
		options.MinimumFractionDigits = intPtr(2)
		options.MaximumFractionDigits = intPtr(2)
	default:
		return "", WrapInvalidFormatterStyle("number", style)
	}

	formatter, err := numberformat.New(intlbridge.ParseLocale(locale), options)
	if err != nil {
		return "", fmt.Errorf("create number formatter: %w", err)
	}
	return formatter.Format(numberformat.Float(number)), nil
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

// formatDate formats a date with an optional compiler-configured time zone.
// TypeScript original code:
// return new Date(value).toLocaleDateString(lc, options);
func formatDate(value any, lc, size, timeZone string) (string, error) {
	if err := validateDateTimeStyle("date", size); err != nil {
		return "", err
	}
	t, err := coerceDateInput(value)
	if err != nil {
		return "", err
	}
	return formatDateTimeWithSize(t, lc, size, timeZone, false)
}

// TypeScript original code:
// return new Date(value).toLocaleTimeString(lc, options);
func formatTime(value any, lc, size, timeZone string) (string, error) {
	if err := validateDateTimeStyle("time", size); err != nil {
		return "", err
	}
	t, err := coerceTimeInput(value)
	if err != nil {
		return "", err
	}
	return formatDateTimeWithSize(t, lc, size, timeZone, true)
}

// validateDateTimeStyle validates the closed MF1 date/time size vocabulary.
// TypeScript original code:
// size?: 'short' | 'default' | 'long' | 'full';
func validateDateTimeStyle(formatter, size string) error {
	switch size {
	case "", "default", "short", "long", "full":
		return nil
	default:
		return WrapInvalidFormatterStyle(formatter, size)
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
func formatDateTimeWithSize(t time.Time, lc, size, timeZone string, isTime bool) (string, error) {
	loc := intlbridge.ParseLocale(lc)
	opts := dateTimeOptionsForSize(size, isTime)
	if timeZone != "" {
		opts.TimeZone = stringPtr(timeZone)
	} else if opts.TimeZone == nil {
		// Preserve the input time's location so date-only formatting doesn't
		// slide a day in non-UTC runtimes, and time-only formatting reflects
		// the caller's intended wall clock.
		if timeZone := timeZoneName(t.Location()); timeZone != "" {
			opts.TimeZone = stringPtr(timeZone)
		}
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
				Hour:   stringPtr(string(datetimeformat.NumericFieldStyle)),
				Minute: stringPtr(string(datetimeformat.TwoDigitFieldStyle)),
			}
		case "long", "full":
			return datetimeformat.Options{
				Hour:         stringPtr(string(datetimeformat.NumericFieldStyle)),
				Minute:       stringPtr(string(datetimeformat.TwoDigitFieldStyle)),
				Second:       stringPtr(string(datetimeformat.TwoDigitFieldStyle)),
				TimeZoneName: stringPtr(string(datetimeformat.ShortTimeZoneName)),
			}
		default:
			return datetimeformat.Options{
				Hour:   stringPtr(string(datetimeformat.NumericFieldStyle)),
				Minute: stringPtr(string(datetimeformat.TwoDigitFieldStyle)),
				Second: stringPtr(string(datetimeformat.TwoDigitFieldStyle)),
			}
		}
	}
	switch size {
	case "short":
		// Numeric Y/M/D mirrors the original Go behavior ("5/4/2026") rather
		// than ECMA-402 DateStyle=short, which uses 2-digit years in en-US.
		return datetimeformat.Options{
			Year:  stringPtr(string(datetimeformat.NumericFieldStyle)),
			Month: stringPtr(string(datetimeformat.NumericMonthStyle)),
			Day:   stringPtr(string(datetimeformat.NumericFieldStyle)),
		}
	case "long":
		return datetimeformat.Options{DateStyle: stringPtr(string(datetimeformat.LongDateTimeStyle))}
	case "full":
		return datetimeformat.Options{DateStyle: stringPtr(string(datetimeformat.FullDateTimeStyle))}
	default:
		return datetimeformat.Options{DateStyle: stringPtr(string(datetimeformat.MediumDateTimeStyle))}
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
