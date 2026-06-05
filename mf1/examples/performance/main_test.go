package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageCacheFormatsAndCachesTemplates(t *testing.T) {
	t.Parallel()

	cache, err := NewMessageCache("en", nil)
	require.NoError(t, err)

	template := "Hello {name}, you have {count, plural, one {# message} other {# messages}}."
	got, err := cache.Format(template, map[string]any{"name": "Ada", "count": 2})
	require.NoError(t, err)
	assert.Equal(t, "Hello Ada, you have 2 messages.", got)

	size, memStats := cache.CacheStats()
	assert.Equal(t, 1, size)
	assert.NotZero(t, memStats.Sys)

	_, err = cache.GetCompiledMessage(template)
	require.NoError(t, err)
	size, _ = cache.CacheStats()
	assert.Equal(t, 1, size)
}

func TestMessageCacheReportsCompileErrors(t *testing.T) {
	t.Parallel()

	cache, err := NewMessageCache("en", nil)
	require.NoError(t, err)

	_, err = cache.GetCompiledMessage("{broken")
	assert.Error(t, err)
}

func TestMessageCacheFormatsConcurrently(t *testing.T) {
	t.Parallel()

	cache, err := NewMessageCache("en", nil)
	require.NoError(t, err)

	template := "{user} sent {count, plural, one {# message} other {# messages}} to {recipient}"
	var wg sync.WaitGroup
	errors := make(chan error, 8)

	for worker := range 8 {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			got, err := cache.Format(template, map[string]any{
				"user":      fmt.Sprintf("User%d", worker),
				"count":     worker + 1,
				"recipient": "Team",
			})
			if err != nil {
				errors <- err
				return
			}
			if got == "" {
				errors <- assert.AnError
			}
		}(worker)
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}

	size, _ := cache.CacheStats()
	assert.Equal(t, 1, size)
}
