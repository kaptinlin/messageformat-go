package functions

import (
	"testing"

	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertResolutionErrorType(t *testing.T, err error, want string) {
	t.Helper()

	var resolutionErr *pkgerrors.MessageResolutionError
	require.ErrorAs(t, err, &resolutionErr)
	assert.Equal(t, want, resolutionErr.Type)
}

func assertFunctionErrorType(t *testing.T, err error, want string) {
	t.Helper()

	var functionErr *pkgerrors.MessageFunctionError
	require.ErrorAs(t, err, &functionErr)
	assert.Equal(t, want, functionErr.Type)
}

func newTestContext(onError func(error)) MessageFunctionContext {
	return NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		onError,
		nil,
		"",
		"",
	)
}
