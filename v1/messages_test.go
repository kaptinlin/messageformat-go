package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessages_AccessorsResolveLocaleAndFallback(t *testing.T) {
	t.Parallel()

	messages := NewMessages(map[string]MessageData{
		"en": {
			"greeting": MessageFunction(func(param any) (any, error) {
				props, ok := param.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("unexpected props %T", param)
				}
				return fmt.Sprintf("Hello %s", props["name"]), nil
			}),
			"nested": MessageData{"title": "Dashboard"},
		},
		"fr": {
			"greeting": MessageFunction(func(param any) (any, error) { return "Bonjour", nil }),
		},
		"toString": {"ignored": "value"},
	}, "en-US")

	require.NotNil(t, messages.Locale())
	assert.Equal(t, "en", *messages.Locale())
	require.NotNil(t, messages.DefaultLocale())
	assert.Equal(t, "en", *messages.DefaultLocale())
	assert.ElementsMatch(t, []string{"en", "fr"}, messages.AvailableLocales())

	messages.SetLocale("fr-CA")
	require.NotNil(t, messages.Locale())
	assert.Equal(t, "fr", *messages.Locale())
	assert.Equal(t, []string{"en"}, messages.GetFallback(nil))

	messages.SetFallback("fr", []string{"en"})
	assert.Equal(t, []string{"en"}, messages.GetFallback(nil))
	assert.True(t, messages.HasMessage("greeting", nil))
	assert.True(t, messages.HasObject("nested", nil, true))

	got, err := messages.Get("greeting", map[string]any{"name": "Ada"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "Bonjour", got)

	got, err = messages.Get([]string{"nested", "title"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Dashboard", got)

	got, err = messages.Get("missing", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "missing", got)
}

func TestMessages_AddMessages(t *testing.T) {
	t.Parallel()

	messages := NewMessages(map[string]MessageData{}, "")

	messages.AddMessages(map[string]any{"toString": "ignored", "title": "Welcome"}, "en", nil)
	got, err := messages.Get("title", nil, Ptr("en"))
	require.NoError(t, err)
	assert.Equal(t, "Welcome", got)
	assert.False(t, messages.HasMessage("toString", Ptr("en")))

	messages.AddMessages(MessageFunction(func(param any) (any, error) {
		return "Account", nil
	}), "en", []string{"pages", "account"})

	assert.True(t, messages.HasObject("pages", Ptr("en")))
	assert.True(t, messages.HasMessage([]string{"pages", "account"}, Ptr("en")))

	got, err = messages.Get([]string{"pages", "account"}, nil, Ptr("en"))
	require.NoError(t, err)
	assert.Equal(t, "Account", got)
}

func TestMessages_DefaultFallback(t *testing.T) {
	t.Parallel()

	messages := NewMessages(map[string]MessageData{
		"en": {"title": "Hello"},
		"es": {},
	}, "en")
	messages.SetLocale("es")

	assert.Equal(t, []string{"en"}, messages.GetFallback(nil))

	got, err := messages.Get("title", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello", got)

	messages.SetDefaultLocale(nil)
	assert.Empty(t, messages.GetFallback(nil))
}
