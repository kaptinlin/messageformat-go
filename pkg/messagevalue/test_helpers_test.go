package messagevalue

import (
	"testing"
	"time"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/stretchr/testify/require"
)

// mustNumberValue constructs a valid number value for behavior tests.
// TypeScript original code:
// const value = getMessageNumber(context, number, options, true);
func mustNumberValue(t *testing.T, value any, locale, source string, options map[string]any) *NumberValue {
	t.Helper()

	number, err := NewNumberValue(value, locale, source, options)
	require.NoError(t, err)
	return number
}

// mustDateTimeValue constructs a valid datetime value for behavior tests.
// TypeScript original code:
// const value = getMessageDateTime(context, date, options);
func mustDateTimeValue(t *testing.T, value time.Time, locale, source string, options map[string]any) *DateTimeValue {
	t.Helper()

	dateTime, err := NewDateTimeValue(value, locale, source, options)
	require.NoError(t, err)
	return dateTime
}

// mustDateTimeValueWithDir constructs a valid directed datetime value for behavior tests.
// TypeScript original code:
// const value = getMessageDateTime({ ...context, dir }, date, options);
func mustDateTimeValueWithDir(t *testing.T, value time.Time, locale, source string, dir bidi.Direction, options map[string]any) *DateTimeValue {
	t.Helper()

	dateTime, err := NewDateTimeValueWithDir(value, locale, source, dir, options)
	require.NoError(t, err)
	return dateTime
}

// mustNumberValueWithDir constructs a valid directed number value for behavior tests.
// TypeScript original code:
// const value = getMessageNumber({ ...context, dir }, number, options, true);
func mustNumberValueWithDir(t *testing.T, value any, locale, source string, dir bidi.Direction, options map[string]any) *NumberValue {
	t.Helper()

	number, err := NewNumberValueWithDir(value, locale, source, dir, options)
	require.NoError(t, err)
	return number
}

// mustNumberValueWithSelection constructs a valid selectable number value for behavior tests.
// TypeScript original code:
// const value = getMessageNumber(context, number, options, selectable);
func mustNumberValueWithSelection(t *testing.T, value any, locale, source string, dir bidi.Direction, options map[string]any, selectable bool) *NumberValue {
	t.Helper()

	number, err := NewNumberValueWithSelection(value, locale, source, dir, options, selectable)
	require.NoError(t, err)
	return number
}
