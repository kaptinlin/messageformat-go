// Package cst provides tests for resource option parsing
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMessagesInResources tests message parsing with resource option
// TypeScript original code:
//
//	describe('messages in resources', () => {
//	  test('text character escapes', () => {
//	    const src = '\\\t\\ \\n\\r\\t\\x01\\u0002\\U000003';
//	    const noRes = parseCST(src, { resource: false });
//	    expect(noRes.errors).toHaveLength(8);
//	    const msg = parseCST(src, { resource: true });
//	    expect(msg).toMatchObject<CST.Message>({
//	      type: 'simple',
//	      errors: [],
//	      pattern: {
//	        start: 0,
//	        end: src.length,
//	        body: [
//	          {
//	            type: 'text',
//	            start: 0,
//	            end: src.length,
//	            value: '\t \n\r\t\x01\x02\x03'
//	          }
//	        ]
//	      }
//	    });
//	  });
func TestMessagesInResources(t *testing.T) {
	t.Run("text character escapes", func(t *testing.T) {
		// TypeScript original code: const src = '\\\t\\ \\n\\r\\t\\x01\\u0002\\U000003';
		src := "\\\t\\ \\n\\r\\t\\x01\\u0002\\U000003"

		// TypeScript original code: const noRes = parseCST(src, { resource: false });
		noRes := ParseCST(src, false)
		// TypeScript original code: expect(noRes.errors).toHaveLength(8);
		assert.Len(t, noRes.Errors(), 8)

		// TypeScript original code: const msg = parseCST(src, { resource: true });
		msg := ParseCST(src, true)
		// TypeScript original code: expect(msg).toMatchObject<CST.Message>({
		require.Equal(t, "simple", msg.Type())
		assert.Empty(t, msg.Errors())

		simple, ok := msg.(*SimpleMessage)
		require.True(t, ok)

		pattern := simple.Pattern()
		assert.Equal(t, 0, pattern.Start())
		assert.Equal(t, len(src), pattern.End())

		body := pattern.Body()
		require.Len(t, body, 1)

		text, ok := body[0].(*Text)
		require.True(t, ok)
		assert.Equal(t, "text", text.Type())
		assert.Equal(t, 0, text.Start())
		assert.Equal(t, len(src), text.End())
		// TypeScript original code: value: '\t \n\r\t\x01\x02\x03'
		assert.Equal(t, "\t \n\r\t\x01\x02\x03", text.Value())
	})

	t.Run("quoted literal character escapes", func(t *testing.T) {
		// TypeScript original code: const src = '{|\\\t\\ \\n\\r\\t\\x01\\u0002\\U000003|}';
		src := "{|\\\t\\ \\n\\r\\t\\x01\\u0002\\U000003|}"

		// TypeScript original code: const noRes = parseCST(src, { resource: false });
		noRes := ParseCST(src, false)
		// TypeScript original code: expect(noRes.errors).toHaveLength(8);
		assert.Len(t, noRes.Errors(), 8)

		// TypeScript original code: const msg = parseCST(src, { resource: true });
		msg := ParseCST(src, true)
		// TypeScript original code: expect(msg).toMatchObject<CST.Message>({
		require.Equal(t, "simple", msg.Type())
		assert.Empty(t, msg.Errors())

		simple, ok := msg.(*SimpleMessage)
		require.True(t, ok)

		pattern := simple.Pattern()
		assert.Equal(t, 0, pattern.Start())
		assert.Equal(t, len(src), pattern.End())

		body := pattern.Body()
		require.Len(t, body, 1)

		expr, ok := body[0].(*Expression)
		require.True(t, ok)
		assert.Equal(t, "expression", expr.Type())
		assert.Equal(t, 0, expr.Start())
		assert.Equal(t, len(src), expr.End())

		braces := expr.Braces()
		require.Len(t, braces, 2)
		assert.Equal(t, 0, braces[0].Start())
		assert.Equal(t, 1, braces[0].End())
		assert.Equal(t, "{", braces[0].Value())
		assert.Equal(t, len(src)-1, braces[1].Start())
		assert.Equal(t, len(src), braces[1].End())
		assert.Equal(t, "}", braces[1].Value())

		arg := expr.Arg()
		require.NotNil(t, arg)
		literal, ok := arg.(*Literal)
		require.True(t, ok)
		assert.Equal(t, "literal", literal.Type())
		assert.True(t, literal.Quoted())
		assert.Equal(t, 1, literal.Start())
		assert.Equal(t, len(src)-1, literal.End())
		// TypeScript original code: value: '\t \n\r\t\x01\x02\x03'
		assert.Equal(t, "\t \n\r\t\x01\x02\x03", literal.Value())

		assert.Nil(t, expr.FunctionRef())
		assert.Empty(t, expr.Attributes())
	})

	t.Run("complex pattern with leading .", func(t *testing.T) {
		// TypeScript original code: const src = '{{.local}}';
		src := "{{.local}}"

		// TypeScript original code: const msg = parseCST(src, { resource: false });
		msg := ParseCST(src, false)
		// TypeScript original code: expect(msg).toMatchObject<CST.Message>({
		require.Equal(t, "complex", msg.Type())
		assert.Empty(t, msg.Errors())

		complex, ok := msg.(*ComplexMessage)
		require.True(t, ok)
		assert.Empty(t, complex.Declarations())

		pattern := complex.Pattern()
		braces := pattern.Braces()
		require.Len(t, braces, 2)
		assert.Equal(t, 0, braces[0].Start())
		assert.Equal(t, 2, braces[0].End())
		assert.Equal(t, "{{", braces[0].Value())
		assert.Equal(t, 8, braces[1].Start())
		assert.Equal(t, 10, braces[1].End())
		assert.Equal(t, "}}", braces[1].Value())

		assert.Equal(t, 0, pattern.Start())
		assert.Equal(t, len(src), pattern.End())

		body := pattern.Body()
		require.Len(t, body, 1)

		text, ok := body[0].(*Text)
		require.True(t, ok)
		assert.Equal(t, "text", text.Type())
		assert.Equal(t, 2, text.Start())
		assert.Equal(t, len(src)-2, text.End())
		assert.Equal(t, ".local", text.Value())
	})

	t.Run("newlines in text", func(t *testing.T) {
		// TypeScript original code: const src = '1\n \t2 \n \\ 3\n\\t';
		src := "1\n \t2 \n \\ 3\n\\t"

		// TypeScript original code: const noRes = parseCST(src, { resource: false });
		noRes := ParseCST(src, false)
		// TypeScript original code: expect(noRes).toMatchObject({
		require.Equal(t, "simple", noRes.Type())
		// TypeScript original code: errors: [{ type: 'bad-escape' }, { type: 'bad-escape' }],
		assert.Len(t, noRes.Errors(), 2)
		for _, err := range noRes.Errors() {
			assert.Equal(t, "bad-escape", string(err.Type))
		}

		simple, ok := noRes.(*SimpleMessage)
		require.True(t, ok)
		pattern := simple.Pattern()
		body := pattern.Body()
		require.Len(t, body, 1)
		text, ok := body[0].(*Text)
		require.True(t, ok)
		// TypeScript original code: pattern: { body: [{ type: 'text', value: '1\n \t2 \n \\ 3\n\\t' }] }
		assert.Equal(t, "1\n \t2 \n \\ 3\n\\t", text.Value())

		// TypeScript original code: const msg = parseCST(src, { resource: true });
		msg := ParseCST(src, true)
		// TypeScript original code: expect(msg).toMatchObject({
		require.Equal(t, "simple", msg.Type())
		assert.Empty(t, msg.Errors())

		simple, ok = msg.(*SimpleMessage)
		require.True(t, ok)
		pattern = simple.Pattern()
		body = pattern.Body()
		require.Len(t, body, 1)
		text, ok = body[0].(*Text)
		require.True(t, ok)
		// TypeScript original code: pattern: { body: [{ type: 'text', value: '1\n2 \n 3\n\t' }] }
		assert.Equal(t, "1\n2 \n 3\n\t", text.Value())
	})

	t.Run("newlines in quoted literal", func(t *testing.T) {
		// TypeScript original code: const src = '{|1\n \t2 \n \\ 3\n\\t|}';
		src := "{|1\n \t2 \n \\ 3\n\\t|}"

		// TypeScript original code: const noRes = parseCST(src, { resource: false });
		noRes := ParseCST(src, false)
		// TypeScript original code: expect(noRes).toMatchObject({
		require.Equal(t, "simple", noRes.Type())
		// TypeScript original code: errors: [{ type: 'bad-escape' }, { type: 'bad-escape' }],
		assert.Len(t, noRes.Errors(), 2)
		for _, err := range noRes.Errors() {
			assert.Equal(t, "bad-escape", string(err.Type))
		}

		simple, ok := noRes.(*SimpleMessage)
		require.True(t, ok)
		pattern := simple.Pattern()
		body := pattern.Body()
		require.Len(t, body, 1)
		expr, ok := body[0].(*Expression)
		require.True(t, ok)
		arg := expr.Arg()
		require.NotNil(t, arg)
		literal, ok := arg.(*Literal)
		require.True(t, ok)
		// TypeScript original code: arg: { type: 'literal', value: '1\n \t2 \n \\ 3\n\\t' }
		assert.Equal(t, "1\n \t2 \n \\ 3\n\\t", literal.Value())

		// TypeScript original code: const msg = parseCST(src, { resource: true });
		msg := ParseCST(src, true)
		// TypeScript original code: expect(msg).toMatchObject({
		require.Equal(t, "simple", msg.Type())
		assert.Empty(t, msg.Errors())

		simple, ok = msg.(*SimpleMessage)
		require.True(t, ok)
		pattern = simple.Pattern()
		body = pattern.Body()
		require.Len(t, body, 1)
		expr, ok = body[0].(*Expression)
		require.True(t, ok)
		arg = expr.Arg()
		require.NotNil(t, arg)
		literal, ok = arg.(*Literal)
		require.True(t, ok)
		// TypeScript original code: arg: { type: 'literal', value: '1\n2 \n 3\n\t' }
		assert.Equal(t, "1\n2 \n 3\n\t", literal.Value())
	})
}
