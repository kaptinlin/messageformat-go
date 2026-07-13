package messageformat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func TestFunctionalOptions(t *testing.T) {
	t.Parallel()

	custom := func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom", ctx.Locales()[0], ctx.Source())
	}
	alsoCustom := func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
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

func TestOptionConstructors(t *testing.T) {
	t.Parallel()

	messageOptions := NewMessageFormatOptions(nil, WithDir(DirRTL))
	assert.Equal(t, DirRTL, messageOptions.Dir)

	freshMessageOptions := NewMessageFormatOptions()
	assert.Equal(t, Direction(""), freshMessageOptions.Dir)
	assert.NotSame(t, messageOptions, freshMessageOptions)

	formatOptions := NewFormatOptions(nil, WithErrorHandler(func(error) {}))
	assert.NotNil(t, formatOptions.OnError)

	freshFormatOptions := NewFormatOptions()
	assert.Nil(t, freshFormatOptions.OnError)
	assert.NotSame(t, formatOptions, freshFormatOptions)
}

func TestConstructorsRejectInvalidBidiIsolation(t *testing.T) {
	t.Parallel()

	message, err := datamodel.ParseMessage("Hello")
	require.NoError(t, err)

	tests := []struct {
		name      string
		construct func() error
	}{
		{
			name: "parse",
			construct: func() error {
				_, err := Parse([]string{"en"}, "Hello", WithBidiIsolation(BidiIsolation("invalid")))
				return err
			},
		},
		{
			name: "compile",
			construct: func() error {
				_, err := Compile([]string{"en"}, message, WithBidiIsolation(BidiIsolation("invalid")))
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.construct()
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidOption))
		})
	}
}

func TestConstructorsRejectInvalidDirectionAndLocaleMatcher(t *testing.T) {
	t.Parallel()

	message, err := datamodel.ParseMessage("Hello")
	require.NoError(t, err)

	tests := []struct {
		name   string
		option Option
	}{
		{name: "direction", option: WithDir(Direction("invalid"))},
		{name: "locale matcher", option: WithLocaleMatcher(LocaleMatcher("invalid"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			t.Run("parse", func(t *testing.T) {
				t.Parallel()

				_, err := Parse([]string{"en"}, "Hello", tt.option)
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidOption))
			})

			t.Run("compile", func(t *testing.T) {
				t.Parallel()

				_, err := Compile([]string{"en"}, message, tt.option)
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidOption))
			})
		})
	}
}

func TestConstructorOptionValidationCoversStructAndValidVocabulary(t *testing.T) {
	t.Parallel()

	t.Run("options struct rejects invalid values", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			options MessageFormatOptions
		}{
			{name: "bidi isolation", options: MessageFormatOptions{BidiIsolation: BidiIsolation("invalid")}},
			{name: "direction", options: MessageFormatOptions{Dir: Direction("invalid")}},
			{name: "locale matcher", options: MessageFormatOptions{LocaleMatcher: LocaleMatcher("invalid")}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := Parse([]string{"en"}, "Hello", Options(tt.options))
				assert.ErrorIs(t, err, ErrInvalidOption)
			})
		}
	})

	t.Run("public constants and zero values remain valid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			option Option
		}{
			{name: "zero value", option: Options(MessageFormatOptions{})},
			{name: "bidi default", option: WithBidiIsolation(BidiDefault)},
			{name: "bidi none", option: WithBidiIsolation(BidiNone)},
			{name: "dir ltr", option: WithDir(DirLTR)},
			{name: "dir rtl", option: WithDir(DirRTL)},
			{name: "dir auto", option: WithDir(DirAuto)},
			{name: "locale best fit", option: WithLocaleMatcher(LocaleBestFit)},
			{name: "locale lookup", option: WithLocaleMatcher(LocaleLookup)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := Parse([]string{"en"}, "Hello", tt.option)
				require.NoError(t, err)
			})
		}
	})
}
