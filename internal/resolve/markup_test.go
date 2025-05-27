package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestSimpleOpenClose tests simple open/close markup elements
// TypeScript original code:
//
//	describe('Simple open/close', () => {
//	  test('no options, literal body', () => {
//	    const mf = new MessageFormat(undefined, '{#b}foo{/b}');
//	    expect(mf.formatToParts()).toEqual([
//	      { type: 'markup', kind: 'open', name: 'b' },
//	      { type: 'text', value: 'foo' },
//	      { type: 'markup', kind: 'close', name: 'b' }
//	    ]);
//	    expect(mf.format()).toBe('foo');
//	  });
func TestSimpleOpenClose(t *testing.T) {
	t.Run("no options, literal body", func(t *testing.T) {
		// Create markup elements directly for testing
		openMarkup := datamodel.NewMarkup("open", "b", nil, nil)
		textElement := datamodel.NewTextElement("foo")
		closeMarkup := datamodel.NewMarkup("close", "b", nil, nil)

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		// Test open markup
		openPart := FormatMarkup(ctx, openMarkup)
		assert.Equal(t, "markup", openPart.Type())
		if mp, ok := openPart.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "open", mp.Kind())
			assert.Equal(t, "b", mp.Name())
		}

		// Test text element
		textPart := messagevalue.NewTextPart(textElement.Value(), textElement.Value(), "")
		assert.Equal(t, "text", textPart.Type())
		assert.Equal(t, "foo", textPart.Value())

		// Test close markup
		closePart := FormatMarkup(ctx, closeMarkup)
		assert.Equal(t, "markup", closePart.Type())
		if mp, ok := closePart.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "close", mp.Kind())
			assert.Equal(t, "b", mp.Name())
		}
	})

	t.Run("options", func(t *testing.T) {
		// Create markup with options
		options := datamodel.Options{
			"foo": datamodel.NewLiteral("42"),
		}

		openMarkup := datamodel.NewMarkup("open", "b", options, nil)

		// Create a basic context with variables
		values := map[string]interface{}{
			"foo": "foo bar",
		}
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, values, nil)

		// Test markup with options
		openPart := FormatMarkup(ctx, openMarkup)
		assert.Equal(t, "markup", openPart.Type())
		if mp, ok := openPart.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "open", mp.Kind())
			assert.Equal(t, "b", mp.Name())
			assert.NotNil(t, mp.Options())
		}
	})
}

// TestMultipleOpenClose tests multiple open/close markup elements
// TypeScript original code:
//
//	describe('Multiple open/close', () => {
//	  test('adjacent', () => {
//	    const mf = new MessageFormat(undefined, '{#b}foo{/b}{#a}bar{/a}');
//	    expect(mf.formatToParts()).toEqual([
//	      { type: 'markup', kind: 'open', name: 'b' },
//	      { type: 'text', value: 'foo' },
//	      { type: 'markup', kind: 'close', name: 'b' },
//	      { type: 'markup', kind: 'open', name: 'a' },
//	      { type: 'text', value: 'bar' },
//	      { type: 'markup', kind: 'close', name: 'a' }
//	    ]);
//	    expect(mf.format()).toBe('foobar');
//	  });
func TestMultipleOpenClose(t *testing.T) {
	t.Run("adjacent", func(t *testing.T) {
		// Create adjacent markup elements
		elements := []struct {
			kind string
			name string
			text string
		}{
			{"open", "b", ""},
			{"", "", "foo"},
			{"close", "b", ""},
			{"open", "a", ""},
			{"", "", "bar"},
			{"close", "a", ""},
		}

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		var parts []messagevalue.MessagePart
		for _, elem := range elements {
			if elem.kind != "" {
				markup := datamodel.NewMarkup(elem.kind, elem.name, nil, nil)
				part := FormatMarkup(ctx, markup)
				parts = append(parts, part)
			} else if elem.text != "" {
				textPart := messagevalue.NewTextPart(elem.text, elem.text, "")
				parts = append(parts, textPart)
			}
		}

		// Verify the parts
		require.Len(t, parts, 6)
		assert.Equal(t, "markup", parts[0].Type())
		assert.Equal(t, "text", parts[1].Type())
		assert.Equal(t, "foo", parts[1].Value())
		assert.Equal(t, "markup", parts[2].Type())
		assert.Equal(t, "markup", parts[3].Type())
		assert.Equal(t, "text", parts[4].Type())
		assert.Equal(t, "bar", parts[4].Value())
		assert.Equal(t, "markup", parts[5].Type())
	})

	t.Run("nested", func(t *testing.T) {
		// Create nested markup elements
		elements := []struct {
			kind string
			name string
			text string
		}{
			{"open", "b", ""},
			{"", "", "foo"},
			{"open", "a", ""},
			{"", "", "bar"},
			{"close", "a", ""},
			{"close", "b", ""},
		}

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		var parts []messagevalue.MessagePart
		for _, elem := range elements {
			if elem.kind != "" {
				markup := datamodel.NewMarkup(elem.kind, elem.name, nil, nil)
				part := FormatMarkup(ctx, markup)
				parts = append(parts, part)
			} else if elem.text != "" {
				textPart := messagevalue.NewTextPart(elem.text, elem.text, "")
				parts = append(parts, textPart)
			}
		}

		// Verify the parts
		require.Len(t, parts, 6)
		assert.Equal(t, "markup", parts[0].Type())
		assert.Equal(t, "text", parts[1].Type())
		assert.Equal(t, "foo", parts[1].Value())
		assert.Equal(t, "markup", parts[2].Type())
		assert.Equal(t, "text", parts[3].Type())
		assert.Equal(t, "bar", parts[3].Value())
		assert.Equal(t, "markup", parts[4].Type())
		assert.Equal(t, "markup", parts[5].Type())
	})

	t.Run("overlapping", func(t *testing.T) {
		// Create overlapping markup elements
		elements := []struct {
			kind string
			name string
			text string
		}{
			{"open", "b", ""},
			{"", "", "foo"},
			{"open", "a", ""},
			{"", "", "bar"},
			{"close", "b", ""},
			{"", "", "baz"},
			{"close", "a", ""},
		}

		// Create a basic context
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		var parts []messagevalue.MessagePart
		for _, elem := range elements {
			if elem.kind != "" {
				markup := datamodel.NewMarkup(elem.kind, elem.name, nil, nil)
				part := FormatMarkup(ctx, markup)
				parts = append(parts, part)
			} else if elem.text != "" {
				textPart := messagevalue.NewTextPart(elem.text, elem.text, "")
				parts = append(parts, textPart)
			}
		}

		// Verify the parts
		require.Len(t, parts, 7)
		assert.Equal(t, "markup", parts[0].Type())
		assert.Equal(t, "text", parts[1].Type())
		assert.Equal(t, "foo", parts[1].Value())
		assert.Equal(t, "markup", parts[2].Type())
		assert.Equal(t, "text", parts[3].Type())
		assert.Equal(t, "bar", parts[3].Value())
		assert.Equal(t, "markup", parts[4].Type())
		assert.Equal(t, "text", parts[5].Type())
		assert.Equal(t, "baz", parts[5].Value())
		assert.Equal(t, "markup", parts[6].Type())
	})
}

// TestFormatMarkup tests the FormatMarkup function directly
func TestFormatMarkup(t *testing.T) {
	t.Run("open markup", func(t *testing.T) {
		markup := datamodel.NewMarkup("open", "div", nil, nil)
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		part := FormatMarkup(ctx, markup)
		assert.Equal(t, "markup", part.Type())
		if mp, ok := part.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "open", mp.Kind())
			assert.Equal(t, "div", mp.Name())
			assert.Equal(t, "div", mp.Value())
		}
	})

	t.Run("close markup", func(t *testing.T) {
		markup := datamodel.NewMarkup("close", "div", nil, nil)
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		part := FormatMarkup(ctx, markup)
		assert.Equal(t, "markup", part.Type())
		if mp, ok := part.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "close", mp.Kind())
			assert.Equal(t, "div", mp.Name())
			assert.Equal(t, "div", mp.Value())
		}
	})

	t.Run("standalone markup", func(t *testing.T) {
		markup := datamodel.NewMarkup("standalone", "br", nil, nil)
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		part := FormatMarkup(ctx, markup)
		assert.Equal(t, "markup", part.Type())
		if mp, ok := part.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "standalone", mp.Kind())
			assert.Equal(t, "br", mp.Name())
			assert.Equal(t, "br", mp.Value())
		}
	})

	t.Run("markup with options", func(t *testing.T) {
		options := datamodel.Options{
			"class": datamodel.NewLiteral("bold"),
			"id":    datamodel.NewLiteral("main"),
		}

		markup := datamodel.NewMarkup("open", "span", options, nil)
		ctx := NewContext([]string{"en"}, functions.DefaultFunctions, nil, nil)

		part := FormatMarkup(ctx, markup)
		assert.Equal(t, "markup", part.Type())
		if mp, ok := part.(*messagevalue.MarkupPart); ok {
			assert.Equal(t, "open", mp.Kind())
			assert.Equal(t, "span", mp.Name())
			assert.NotNil(t, mp.Options())
			assert.Contains(t, mp.Options(), "class")
			assert.Contains(t, mp.Options(), "id")
		}
	})
}
