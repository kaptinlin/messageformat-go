package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
)

func TestParseMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		source   string
		wantType string
	}{
		{
			name:     "pattern message",
			source:   "Hello {$name}!",
			wantType: "message",
		},
		{
			name:     "select message",
			source:   ".input {$count :number}\n.match $count\none {{One item}}\n* {{Many items}}",
			wantType: "select",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			message, err := ParseMessage(tt.source)

			require.NoError(t, err)
			require.NotNil(t, message)
			assert.Equal(t, tt.wantType, message.Type())
		})
	}
}

func TestParseMessage_SyntaxError(t *testing.T) {
	t.Parallel()

	message, err := ParseMessage("Hello {$name")

	require.Error(t, err)
	assert.Nil(t, message)

	var syntaxErr *pkgerrors.MessageSyntaxError
	require.ErrorAs(t, err, &syntaxErr)
	assert.Equal(t, pkgerrors.ErrorTypeParseError, syntaxErr.ErrorType())
	assert.GreaterOrEqual(t, syntaxErr.Start, 0)
	assert.GreaterOrEqual(t, syntaxErr.End, syntaxErr.Start)
}
