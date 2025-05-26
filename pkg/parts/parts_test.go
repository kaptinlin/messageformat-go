package parts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTextPart(t *testing.T) {
	part := NewTextPart("Hello World")
	assert.NotNil(t, part)
	assert.Equal(t, "text", part.Type())
	assert.Equal(t, "Hello World", part.Value())
}

func TestNewBidiIsolationPart(t *testing.T) {
	part := NewBidiIsolationPart("\u2066") // LRI
	assert.NotNil(t, part)
	assert.Equal(t, "bidiIsolation", part.Type())
	assert.Equal(t, "\u2066", part.Value())
}

func TestNewMarkupPart(t *testing.T) {
	options := map[string]interface{}{
		"class": "highlight",
	}
	part := NewMarkupPart("open", "span", "span", options)
	assert.NotNil(t, part)
	assert.Equal(t, "markup", part.Type())
	assert.Equal(t, "span", part.Value())
}

func TestNewFallbackPart(t *testing.T) {
	part := NewFallbackPart("$unknown", "en")
	assert.NotNil(t, part)
	assert.Equal(t, "fallback", part.Type())
	assert.Equal(t, "{$unknown}", part.Value())
}
