package main

import (
	"os"
	"strings"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomFunctionHelpers(t *testing.T) {
	t.Parallel()

	ctx := functions.NewMessageFunctionContext([]string{"en"}, "source", "", func(error) {}, nil, "", "")
	tests := []struct {
		name         string
		fn           functions.MessageFunction
		operand      any
		options      functions.Options
		want         string
		wantContains string
	}{
		{name: "uppercase string", fn: uppercaseFunction, operand: "hello", want: "HELLO"},
		{name: "uppercase non-string", fn: uppercaseFunction, operand: 123, want: "123"},
		{name: "uppercase nil", fn: uppercaseFunction, operand: nil, want: ""},
		{name: "reverse runes", fn: reverseFunction, operand: "hé", want: "éh"},
		{name: "emoji happy", fn: emojiFunction, operand: "done", options: functions.Options{"type": "happy"}, wantContains: "done"},
		{name: "emoji default", fn: emojiFunction, operand: "plain", wantContains: "plain"},
		{name: "time now", fn: timeAgoFunction, operand: 0, want: "just now"},
		{name: "time one hour", fn: timeAgoFunction, operand: 1, want: "1 hour ago"},
		{name: "time hours", fn: timeAgoFunction, operand: 3, want: "3 hours ago"},
		{name: "time one day", fn: timeAgoFunction, operand: 25, want: "1 day ago"},
		{name: "time days", fn: timeAgoFunction, operand: 72, want: "3 days ago"},
		{name: "time weeks", fn: timeAgoFunction, operand: 24 * 14, want: "2 weeks ago"},
		{name: "time months", fn: timeAgoFunction, operand: float64(24 * 60), want: "2 months ago"},
		{name: "format left", fn: formatFunction, operand: "Go", options: functions.Options{"width": 4, "align": "left", "pad": "."}, want: "Go.."},
		{name: "format right", fn: formatFunction, operand: "Go", options: functions.Options{"width": float64(4), "align": "right", "pad": "."}, want: "..Go"},
		{name: "format center", fn: formatFunction, operand: "Go", options: functions.Options{"width": 5, "align": "center", "pad": "."}, want: ".Go.."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value := tc.fn(ctx, tc.options, tc.operand)
			got, err := value.ToString()
			require.NoError(t, err)
			if tc.wantContains != "" {
				assert.True(t, strings.Contains(got, tc.wantContains), "formatted value %q should contain %q", got, tc.wantContains)
			} else {
				assert.Equal(t, tc.want, got)
			}
			assert.Equal(t, "en", value.Locale())
			assert.Equal(t, "source", value.Source())
		})
	}
}

func TestCustomFunctionsExampleRuns(t *testing.T) {
	silenceStdout(t)

	main()
}

func silenceStdout(t *testing.T) {
	t.Helper()

	originalStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)

	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = originalStdout
		require.NoError(t, devNull.Close())
	})
}
