package v1_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/kaptinlin/messageformat-go/mf1"
)

func TestTypedConstructors(t *testing.T) {
	t.Parallel()

	t.Run("locale constructor", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			locale     string
			wantLocale string
			wantErr    error
		}{
			{name: "supported", locale: "en", wantLocale: "en"},
			{name: "valid unsupported", locale: "eo", wantLocale: "en"},
			{name: "empty", locale: "", wantErr: v1.ErrInvalidLocale},
			{name: "malformed", locale: "x", wantErr: v1.ErrInvalidLocale},
			{name: "unknown language", locale: "xx", wantErr: v1.ErrInvalidLocale},
			{name: "unknown malformed", locale: "lawlz", wantErr: v1.ErrInvalidLocale},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				mf, err := v1.New(tt.locale, nil)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.wantLocale, mf.ResolvedOptions().Locale)
			})
		}
	})

	t.Run("custom plural constructor", func(t *testing.T) {
		t.Parallel()

		plural := v1.PluralFunction(func(any, ...bool) (v1.PluralCategory, error) {
			return v1.PluralMany, nil
		})
		mf, err := v1.NewWithPlural(plural, nil)
		require.NoError(t, err)
		compiled, err := mf.Compile("{count, plural, many {many} other {other}}")
		require.NoError(t, err)
		got, err := compiled(map[string]any{"count": 2})
		require.NoError(t, err)
		assert.Equal(t, "many", got)

		_, err = v1.NewWithPlural(nil, nil)
		require.ErrorIs(t, err, v1.ErrInvalidPluralFunction)
	})
}
