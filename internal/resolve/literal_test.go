package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

// TestResolveLiteral tests the ResolveLiteral function directly
// TypeScript original code:
//
//	export function resolveLiteral(ctx: Context, lit: Literal) {
//	  const msgCtx = new MessageFunctionContext(ctx, `|${lit.value}|`);
//	  return string(msgCtx, {}, lit.value);
//	}
func TestResolveLiteral(t *testing.T) {
	t.Run("simple quoted literal", func(t *testing.T) {
		// Create a literal with quoted content
		literal := datamodel.NewLiteral("quoted literal")

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		// Resolve the literal
		mv := ResolveLiteral(ctx, literal)
		require.NotNil(t, mv)

		// Check the result
		assert.Equal(t, "string", mv.Type())
		assert.Equal(t, "|quoted literal|", mv.Source())

		// Convert to string
		str, err := mv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "quoted literal", str)

		// Convert to parts
		parts, err := mv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)
		assert.Equal(t, "string", parts[0].Type())
		assert.Equal(t, "quoted literal", parts[0].Value())
	})

	t.Run("literal with spaces, newlines and escapes", func(t *testing.T) {
		// Create a literal with special characters
		literal := datamodel.NewLiteral(" quoted \n \\|literal\\|{}")

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		// Resolve the literal
		mv := ResolveLiteral(ctx, literal)
		require.NotNil(t, mv)

		// Check the result
		assert.Equal(t, "string", mv.Type())

		// Convert to string
		str, err := mv.ToString()
		require.NoError(t, err)
		assert.Equal(t, " quoted \n \\|literal\\|{}", str)

		// Convert to parts
		parts, err := mv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)
		assert.Equal(t, "string", parts[0].Type())
		assert.Equal(t, " quoted \n \\|literal\\|{}", parts[0].Value())
	})

	t.Run("number literals", func(t *testing.T) {
		// Test various number formats
		testValues := []string{
			"0",
			"42",
			"2.5",
			"-1",
			"-0.999",
			"1e3",
			"0.4E+5",
			"11.1e-1",
		}

		for _, value := range testValues {
			t.Run(value, func(t *testing.T) {
				// Create a literal with the number value
				literal := datamodel.NewLiteral(value)

				// Create a basic context
				ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

				// Resolve the literal
				mv := ResolveLiteral(ctx, literal)
				require.NotNil(t, mv)

				// Check the result
				assert.Equal(t, "string", mv.Type())
				assert.Equal(t, "|"+value+"|", mv.Source())

				// Convert to string
				str, err := mv.ToString()
				require.NoError(t, err)
				assert.Equal(t, value, str)

				// Convert to parts
				parts, err := mv.ToParts()
				require.NoError(t, err)
				require.Len(t, parts, 1)
				assert.Equal(t, "string", parts[0].Type())
				assert.Equal(t, value, parts[0].Value())
			})
		}
	})

	t.Run("empty literal", func(t *testing.T) {
		// Create an empty literal
		literal := datamodel.NewLiteral("")

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		// Resolve the literal
		mv := ResolveLiteral(ctx, literal)
		require.NotNil(t, mv)

		// Check the result
		assert.Equal(t, "string", mv.Type())
		assert.Equal(t, "||", mv.Source())

		// Convert to string
		str, err := mv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "", str)

		// Convert to parts
		parts, err := mv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)
		assert.Equal(t, "string", parts[0].Type())
		assert.Equal(t, "", parts[0].Value())
	})

	t.Run("literal with missing string function", func(t *testing.T) {
		// Create a literal
		literal := datamodel.NewLiteral("test")

		// Create a context without string function
		emptyFunctions := make(map[string]functions.MessageFunction)
		ctx := NewContext([]string{"en"}, emptyFunctions, nil, nil)

		// Resolve the literal
		mv := ResolveLiteral(ctx, literal)
		require.NotNil(t, mv)

		// Should fallback to StringValue
		assert.Equal(t, "string", mv.Type())
		assert.Equal(t, "|test|", mv.Source())

		// Convert to string
		str, err := mv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "test", str)
	})
}
