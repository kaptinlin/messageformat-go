package tests

import (
	"testing"

	"github.com/kaptinlin/messageformat-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicFunctionality(t *testing.T) {
	t.Run("simple text", func(t *testing.T) {
		mf, err := messageformat.New("en", "Hello world!", nil)
		require.NoError(t, err)

		result, err := mf.Format(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello world!", result)
	})

	t.Run("variable substitution", func(t *testing.T) {
		options := &messageformat.MessageFormatOptions{
			BidiIsolation: messageformat.BidiNone,
		}
		mf, err := messageformat.New("en", "Hello {$name}!", options)
		require.NoError(t, err)

		params := map[string]interface{}{
			"name": "world",
		}
		result, err := mf.Format(params, nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello world!", result)
	})

	t.Run("test function", func(t *testing.T) {
		options := &messageformat.MessageFormatOptions{
			Functions:     TestFunctions,
			BidiIsolation: messageformat.BidiNone,
		}
		mf, err := messageformat.New("en", "{42 :test:function}", options)
		require.NoError(t, err)

		result, err := mf.Format(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "42", result) // test:function should format the number
	})
}
