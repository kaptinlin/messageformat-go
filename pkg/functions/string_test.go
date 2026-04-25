package functions

import (
	"errors"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
)

func TestStringFunction(t *testing.T) {
	t.Parallel()

	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"auto",
		"",
	)

	tests := []struct {
		name     string
		operand  any
		options  map[string]any
		expected string
	}{
		{"string input", "hello", nil, "hello"},
		{"nil input", nil, nil, ""},
		{"number input", 42, nil, "42"},
		{"boolean input", true, nil, "true"},
		{"with locale option", "test", map[string]any{"locale": "fr"}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := StringFunction(ctx, tt.options, tt.operand)

			assert.Equal(t, "string", result.Type())
			assert.Equal(t, "test source", result.Source())

			// Test string conversion
			str, err := result.ToString()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, str)
		})
	}
}

type errStringMessageValue struct{}

func (errStringMessageValue) String() string                               { return "string fallback" }
func (errStringMessageValue) Type() string                                 { return "test" }
func (errStringMessageValue) Source() string                               { return "" }
func (errStringMessageValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (errStringMessageValue) Locale() string                               { return "" }
func (errStringMessageValue) Options() map[string]any                      { return nil }
func (errStringMessageValue) ToString() (string, error)                    { return "", errors.New("boom") }
func (errStringMessageValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (errStringMessageValue) ValueOf() (any, error)                        { return nil, nil }
func (errStringMessageValue) SelectKeys([]string) ([]string, error)        { return nil, nil }

func TestStringFunctionWithDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contextDir  string
		expectedDir string
	}{
		{"ltr direction", "ltr", "ltr"},
		{"rtl direction", "rtl", "rtl"},
		{"auto direction", "auto", "auto"},
		{"empty direction", "", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewMessageFunctionContext(
				[]string{"en"},
				"test source",
				"best fit",
				nil,
				nil,
				tt.contextDir,
				"",
			)

			result := StringFunction(ctx, nil, "test")
			assert.Equal(t, "string", result.Type())
		})
	}
}

func TestStringFunctionFallsBackWhenMessageValueToStringFails(t *testing.T) {
	t.Parallel()

	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"auto",
		"",
	)

	result := StringFunction(ctx, nil, errStringMessageValue{})
	str, err := result.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "string fallback", str)
}
