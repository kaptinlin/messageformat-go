package main

import (
	"os"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHighlightFunction(t *testing.T) {
	t.Parallel()

	ctx := functions.NewMessageFunctionContext([]string{"en"}, "source", "", func(error) {}, nil, "", "")
	tests := []struct {
		name    string
		operand any
		options functions.Options
		want    string
	}{
		{name: "default bold", operand: "Active", options: nil, want: "**Active**"},
		{name: "italic", operand: "High", options: functions.Options{"style": "italic"}, want: "*High*"},
		{name: "underline", operand: "Low", options: functions.Options{"style": "underline"}, want: "_Low_"},
		{name: "code", operand: 12, options: functions.Options{"style": "code"}, want: "`12`"},
		{name: "unknown style", operand: "plain", options: functions.Options{"style": "none"}, want: "plain"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value := highlightFunction(ctx, tc.options, tc.operand)
			got, err := value.ToString()
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, "en", value.Locale())
			assert.Equal(t, "source", value.Source())
		})
	}
}

func TestAdvancedExampleRuns(t *testing.T) {
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
