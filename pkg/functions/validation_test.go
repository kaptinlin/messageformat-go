package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOptionKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid simple key",
			key:       "minimumFractionDigits",
			expectErr: false,
		},
		{
			name:      "valid key with underscore",
			key:       "minimum_fraction_digits",
			expectErr: false,
		},
		{
			name:      "valid key with hyphen",
			key:       "minimum-fraction-digits",
			expectErr: false,
		},
		{
			name:      "valid key with numbers",
			key:       "option123",
			expectErr: false,
		},
		{
			name:      "empty key",
			key:       "",
			expectErr: true,
			errMsg:    "cannot be empty",
		},
		{
			name:      "key too long",
			key:       string(make([]byte, MaxOptionKeyLength+1)),
			expectErr: true,
			errMsg:    "too long",
		},
		{
			name:      "key with space",
			key:       "my option",
			expectErr: true,
			errMsg:    "invalid character",
		},
		{
			name:      "key with special character",
			key:       "option$value",
			expectErr: true,
			errMsg:    "invalid character",
		},
		{
			name:      "forbidden key __proto__",
			key:       "__proto__",
			expectErr: true,
			errMsg:    "forbidden",
		},
		{
			name:      "forbidden key constructor",
			key:       "constructor",
			expectErr: true,
			errMsg:    "forbidden",
		},
		{
			name:      "forbidden key prototype",
			key:       "prototype",
			expectErr: true,
			errMsg:    "forbidden",
		},
		{
			name:      "forbidden key case insensitive",
			key:       "__PROTO__",
			expectErr: true,
			errMsg:    "forbidden",
		},
		{
			name:      "key with dot",
			key:       "option.nested",
			expectErr: true,
			errMsg:    "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptionKey(tt.key)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		options := map[string]interface{}{
			"minimumFractionDigits": 2,
			"maximumFractionDigits": 4,
			"signDisplay":           "always",
		}
		err := ValidateOptions(options)
		assert.NoError(t, err)
	})

	t.Run("too many options", func(t *testing.T) {
		options := make(map[string]interface{})
		for i := range MaxOptionsCount + 1 {
			options[string(rune('a'+i))] = i
		}
		err := ValidateOptions(options)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many options")
	})

	t.Run("invalid option key", func(t *testing.T) {
		options := map[string]interface{}{
			"valid_option": 1,
			"__proto__":    "malicious",
		}
		err := ValidateOptions(options)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "forbidden")
	})

	t.Run("empty options map", func(t *testing.T) {
		options := map[string]interface{}{}
		err := ValidateOptions(options)
		assert.NoError(t, err)
	})
}

func TestSanitizeOptions(t *testing.T) {
	t.Run("filters out invalid keys", func(t *testing.T) {
		options := map[string]interface{}{
			"validOption":  1,
			"__proto__":    "malicious",
			"constructor":  "bad",
			"anotherValid": "ok",
			"bad$key":      "invalid",
		}

		sanitized := SanitizeOptions(options)

		assert.Len(t, sanitized, 2)
		assert.Contains(t, sanitized, "validOption")
		assert.Contains(t, sanitized, "anotherValid")
		assert.NotContains(t, sanitized, "__proto__")
		assert.NotContains(t, sanitized, "constructor")
		assert.NotContains(t, sanitized, "bad$key")
	})

	t.Run("nil options", func(t *testing.T) {
		sanitized := SanitizeOptions(nil)
		assert.Nil(t, sanitized)
	})

	t.Run("all valid options", func(t *testing.T) {
		options := map[string]interface{}{
			"option1": 1,
			"option2": 2,
		}
		sanitized := SanitizeOptions(options)
		assert.Equal(t, options, sanitized)
	})
}

func TestIsValidOptionKeyChar(t *testing.T) {
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	for _, ch := range validChars {
		assert.True(t, isValidOptionKeyChar(ch), "Character '%c' should be valid", ch)
	}

	invalidChars := " !@#$%^&*()+=[]{}|\\:;\"'<>,.?/~`"
	for _, ch := range invalidChars {
		assert.False(t, isValidOptionKeyChar(ch), "Character '%c' should be invalid", ch)
	}
}

// Benchmark tests
func BenchmarkValidateOptionKey(b *testing.B) {
	key := "minimumFractionDigits"
	for b.Loop() {
		_ = ValidateOptionKey(key)
	}
}

func BenchmarkValidateOptions(b *testing.B) {
	options := map[string]interface{}{
		"minimumFractionDigits": 2,
		"maximumFractionDigits": 4,
		"signDisplay":           "always",
		"useGrouping":           "auto",
	}
	for b.Loop() {
		_ = ValidateOptions(options)
	}
}
