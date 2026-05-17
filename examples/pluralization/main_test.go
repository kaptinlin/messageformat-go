package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluralizationExampleRuns(t *testing.T) {
	silenceStdout(t)

	main()
}

func silenceStdout(t *testing.T) {
	t.Helper()

	originalStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)

	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = originalStdout
		require.NoError(t, devNull.Close())
	})
}
