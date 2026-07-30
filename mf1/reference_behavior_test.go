package mf1_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kaptinlin/messageformat-go/mf1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These behavior vectors are adapted from
// .reference/messageformat/mf1/packages/core/src/messages.test.ts.
func TestReferenceMessageFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locale  string
		source  string
		values  map[string]any
		options *mf1.MessageFormatOptions
		want    string
	}{
		{
			name:   "plain message",
			locale: "en",
			source: "This is a string.",
			want:   "This is a string.",
		},
		{
			name:   "Unicode interpolation",
			locale: "en",
			source: "中{test}中国话不用彁字。",
			values: map[string]any{"test": "☺"},
			want:   "中☺中国话不用彁字。",
		},
		{
			name:   "quoted braces",
			locale: "en",
			source: "'{{{'{test}'}}}'",
			values: map[string]any{"test": 4},
			want:   "{{{4}}}",
		},
		{
			name:   "Welsh exact plural",
			locale: "cy",
			source: "{NUM, plural, zero{a} one{b} two{c} few{d} many{e} other{f} =42{omg42}}",
			values: map[string]any{"NUM": 42},
			want:   "omg42",
		},
		{
			name:   "Welsh ordinal category",
			locale: "cy",
			source: "{num, selectordinal, zero{0,7,8,9} one{1} two{2} few{3,4} many{5,6} other{+}}",
			values: map[string]any{"num": 5},
			want:   "5,6",
		},
		{
			name:   "offset applies after exact matching",
			locale: "en",
			source: "{NUM, plural, offset:1 =0{a} one{b} other{c}}",
			values: map[string]any{"NUM": 2},
			want:   "b",
		},
		{
			name:   "quoted octothorpe",
			locale: "en",
			source: "{X, plural, other{# is a '#'}}",
			values: map[string]any{"X": 3},
			want:   "3 is a #",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			compiler, err := mf1.New(tt.locale, tt.options)
			require.NoError(t, err)
			message, err := compiler.Compile(tt.source)
			require.NoError(t, err)

			got, err := message.Format(tt.values)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// The terminal-value vector is adapted from
// .reference/messageformat/mf1/packages/core/src/messages.test.ts.
func TestReferenceMessageValues(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", nil)
	require.NoError(t, err)
	message, err := compiler.Compile("msg {foo}")
	require.NoError(t, err)

	got, err := message.FormatValues(map[string]any{"foo": "FOO"})
	require.NoError(t, err)
	want := []any{"msg ", "FOO"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FormatValues() mismatch (-want +got):\n%s", diff)
	}
}

// The missing-argument behavior is adapted from
// .reference/messageformat/mf1/packages/core/src/messages.test.ts.
func TestReferenceRequiredArgument(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", &mf1.MessageFormatOptions{RequireAllArguments: true})
	require.NoError(t, err)
	message, err := compiler.Compile("{foo}")
	require.NoError(t, err)

	_, err = message.Format(nil)
	require.ErrorIs(t, err, mf1.ErrMissingArgument)
}

// The strict unknown-function behavior is adapted from
// .reference/messageformat/mf1/packages/parser/src/parser.test.ts.
func TestReferenceStrictFunctionValidation(t *testing.T) {
	t.Parallel()

	compiler, err := mf1.New("en", &mf1.MessageFormatOptions{Strict: true})
	require.NoError(t, err)

	_, err = compiler.Compile("{foo, bar}")
	require.Error(t, err)
}

// These escape vectors are adapted from
// .reference/messageformat/mf1/packages/core/src/messages.test.ts.
func TestReferenceEscape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		octothorpe bool
		want       string
	}{
		{name: "left brace", input: "{", want: "'{'"},
		{name: "right brace", input: "}", want: "'}'"},
		{name: "literal octothorpe", input: "#", want: "#"},
		{name: "escaped octothorpe", input: "#", octothorpe: true, want: "'#'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, mf1.Escape(tt.input, tt.octothorpe))
		})
	}
}
