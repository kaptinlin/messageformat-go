package messageformat

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacyPackageRemoved(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	repoRoot := filepath.Dir(file)
	legacyDir := string([]byte{118, 49})
	_, err := os.Stat(filepath.Join(repoRoot, legacyDir))
	assert.ErrorIs(t, err, os.ErrNotExist)
}
