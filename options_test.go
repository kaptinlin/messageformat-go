package messageformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func TestFunctionalOptions(t *testing.T) {
	t.Parallel()

	custom := func(ctx functions.MessageFunctionContext, options map[string]any, operand any) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom", ctx.Locales()[0], ctx.Source())
	}
	alsoCustom := func(ctx functions.MessageFunctionContext, options map[string]any, operand any) messagevalue.MessageValue {
		return messagevalue.NewStringValue("also-custom", ctx.Locales()[0], ctx.Source())
	}

	mf, err := Parse(
		[]string{"en"},
		"{$name :custom}",
		WithBidiIsolation(BidiDefault),
		WithDir(DirRTL),
		WithLocaleMatcher(LocaleLookup),
		WithFunction("custom", custom),
		WithFunctions(map[string]functions.MessageFunction{"alsoCustom": alsoCustom}),
	)
	require.NoError(t, err)

	formatted, err := mf.Format(map[string]any{"name": "ignored"})
	require.NoError(t, err)
	assert.Contains(t, formatted, "custom")
	assert.Equal(t, "rtl", mf.Dir())
	assert.True(t, mf.BidiIsolation())

	resolved := mf.ResolvedOptions()
	assert.Equal(t, BidiDefault, resolved.BidiIsolation)
	assert.Equal(t, DirRTL, resolved.Dir)
	assert.Equal(t, LocaleLookup, resolved.LocaleMatcher)
	assert.Contains(t, resolved.Functions, "custom")
	assert.Contains(t, resolved.Functions, "alsoCustom")

	delete(resolved.Functions, "custom")
	assert.Contains(t, mf.ResolvedOptions().Functions, "custom")
}

func TestStringFunctionalOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		options       []Option
		wantBidi      BidiIsolation
		wantDir       Direction
		wantMatcher   LocaleMatcher
		wantAccessDir string
	}{
		{
			name:          "recognized strings",
			options:       []Option{WithBidiIsolationString("none"), WithDirString("ltr"), WithLocaleMatcherString("lookup")},
			wantBidi:      BidiNone,
			wantDir:       DirLTR,
			wantMatcher:   LocaleLookup,
			wantAccessDir: "ltr",
		},
		{
			name:          "unknown strings use defaults",
			options:       []Option{WithBidiIsolationString("unexpected"), WithDirString("unexpected"), WithLocaleMatcherString("unexpected")},
			wantBidi:      BidiDefault,
			wantDir:       DirLTR,
			wantMatcher:   LocaleBestFit,
			wantAccessDir: "ltr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mf, err := Parse([]string{"en"}, "Hello", tt.options...)
			require.NoError(t, err)

			resolved := mf.ResolvedOptions()
			assert.Equal(t, tt.wantBidi, resolved.BidiIsolation)
			assert.Equal(t, tt.wantDir, resolved.Dir)
			assert.Equal(t, tt.wantMatcher, resolved.LocaleMatcher)
			assert.Equal(t, tt.wantAccessDir, mf.Dir())
		})
	}
}
