package functions

import (
	"testing"
	"time"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatetimeFunctionReturnsDateTimeValue(t *testing.T) {
	// Create a function context
	var errors []error
	onError := func(err error) {
		errors = append(errors, err)
	}

	ctx := NewMessageFunctionContext(
		[]string{"en-US"},
		"test",
		"best fit",
		onError,
		nil,
		"",
		"",
	)

	// Test with ISO date string
	options := map[string]interface{}{
		"dateStyle": "medium",
		"timeStyle": "short",
	}

	result := DatetimeFunction(ctx, options, "2006-01-02T15:04:05")

	// Check for any errors
	if len(errors) > 0 {
		t.Logf("Errors: %v", errors)
	}

	// Verify it returns a DateTimeValue, not a StringValue
	assert.Equal(t, "datetime", result.Type())

	// Verify it's actually a DateTimeValue
	dtv, ok := result.(*messagevalue.DateTimeValue)
	require.True(t, ok, "Expected DateTimeValue, got %T", result)

	// Debug the options
	t.Logf("DateTimeValue options: %+v", dtv.Options())

	// Test the underlying value
	val, err := dtv.ValueOf()
	require.NoError(t, err)
	t.Logf("DateTimeValue time: %v", val)
	expectedTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	assert.Equal(t, expectedTime, val)

	// Verify that options contains new LDML 48 options
	assert.Equal(t, "year-month-day", dtv.Options()["dateFields"], "Expected dateFields option")
	// Note: dateStyle/timeStyle are now converted to dateFields/timePrecision

	// Test string representation
	str, err := dtv.ToString()
	require.NoError(t, err)
	t.Logf("Formatted string: %q", str)

	// Let's also check the locale and source
	t.Logf("DateTimeValue locale: %q", dtv.Locale())
	t.Logf("DateTimeValue source: %q", dtv.Source())

	assert.NotEmpty(t, str)
}

func TestDatetimeFunctionDefaultOptions(t *testing.T) {
	var errors []error
	onError := func(err error) {
		errors = append(errors, err)
	}

	ctx := NewMessageFunctionContext(
		[]string{"en-US"},
		"test",
		"best fit",
		onError,
		nil,
		"",
		"",
	)

	// Test with no options - should get default dateStyle=medium, timeStyle=short
	result := DatetimeFunction(ctx, map[string]interface{}{}, "2006-01-02T15:04:05")

	dtv, ok := result.(*messagevalue.DateTimeValue)
	require.True(t, ok)

	options := dtv.Options()
	assert.Equal(t, "year-month-day", options["dateFields"], "Expected default dateFields")
	// Note: Default behavior now uses dateFields instead of dateStyle/timeStyle
}
