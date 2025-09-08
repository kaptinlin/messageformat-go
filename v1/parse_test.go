package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlainStrings tests parsing of plain string content
// TypeScript original code:
//
//	describe('Plain strings', () => {
//	  run({
//	    'should accept string only input': {
//	      'This is a string': 'This is a string',
//	      '☺☺☺☺': '☺☺☺☺',
//	      ...
//	    }
//	  })
//	})
func TestPlainStrings(t *testing.T) {
	t.Run("should accept string only input", func(t *testing.T) {
		tests := map[string]string{
			"This is a string":           "This is a string",
			"☺☺☺☺":                       "☺☺☺☺",
			"This is \n a string":        "This is \n a string",
			"中国话不用彁字。":                   "中国话不用彁字。",
			" \t leading whitspace":      " \t leading whitspace",
			"trailing whitespace   \n  ": "trailing whitespace   \n  ",
		}

		for input, expected := range tests {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				content, ok := result[0].(*Content)
				require.True(t, ok, "Expected Content token", nil)
				assert.Equal(t, expected, content.Value)
			})
		}
	})

	t.Run("should allow you to escape { and } characters", func(t *testing.T) {
		tests := map[string]string{
			"'{'test":  "{test",
			"test'}'":  "test}",
			"'{test}'": "{test}",
		}

		for input, expected := range tests {
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				content, ok := result[0].(*Content)
				require.True(t, ok, "Expected Content token", nil)
				assert.Equal(t, expected, content.Value)
			})
		}
	})

	t.Run("should gracefully handle quotes", func(t *testing.T) {
		tests := map[string]string{
			"This is a dbl quote: \"":   "This is a dbl quote: \"",
			"This is a single quote: '": "This is a single quote: '",
		}

		for input, expected := range tests {
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				content, ok := result[0].(*Content)
				require.True(t, ok, "Expected Content token", nil)
				assert.Equal(t, expected, content.Value)
			})
		}
	})

	t.Run("should allow you to use extension keywords for plural formats everywhere except where they go", func(t *testing.T) {
		tests := map[string][]interface{}{
			"select select, ":          {"select select, "},
			"select offset, offset:1 ": {"select offset, offset:1 "},
			"one other, =1 ":           {"one other, =1 "},
			"one {select} ":            {"one ", "select", " "},
			"one {plural} ":            {"one ", "plural", " "},
		}

		for input, expected := range tests {
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, len(expected))
				for i, exp := range expected {
					if expStr, ok := exp.(string); ok {
						// For this specific test case pattern: "one {select} " -> ["one ", "select", " "]
						// Even indices (0, 2, 4...) are Content tokens
						// Odd indices (1, 3, 5...) are PlainArg token arg values
						if i%2 == 0 {
							content, ok := result[i].(*Content)
							require.True(t, ok, "Expected Content token at index %d", i)
							assert.Equal(t, expStr, content.Value)
						} else {
							arg, ok := result[i].(*PlainArg)
							require.True(t, ok, "Expected PlainArg token at index %d", i)
							assert.Equal(t, expStr, arg.Arg)
						}
					}
				}
			})
		}
	})

	t.Run("should correctly handle apostrophes", func(t *testing.T) {
		tests := map[string]string{
			"I see '{many}'":      "I see {many}",
			"I said '{''Wow!''}'": "I said {'Wow!'}",
			"I don't know":        "I don't know",
			"I don''t know":       "I don't know",
			"A'a''a'A":            "A'a'a'A",
			"A'{a''a}'A":          "A{a'a}A",
			"A '#' A":             "A '#' A",
			"A '|' A":             "A '|' A",
		}

		for input, expected := range tests {
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				content, ok := result[0].(*Content)
				require.True(t, ok, "Expected Content token", nil)
				assert.Equal(t, expected, content.Value)
			})
		}
	})
}

// TestSimpleArguments tests parsing of simple variable arguments
// TypeScript original code:
//
//	describe('Simple arguments', () => {
//	  run({
//	    'should accept only a variable': {
//	      '{test}': [{ type: 'argument', arg: 'test' }],
//	      '{0}': [{ type: 'argument', arg: '0' }]
//	    }
//	  })
//	})
func TestSimpleArguments(t *testing.T) {
	t.Run("should accept only a variable", func(t *testing.T) {
		tests := map[string]string{
			"{test}": "test",
			"{0}":    "0",
		}

		for input, expected := range tests {
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				arg, ok := result[0].(*PlainArg)
				require.True(t, ok, "Expected PlainArg token", nil)
				assert.Equal(t, expected, arg.Arg)
			})
		}
	})

	t.Run("should not care about white space in a variable", func(t *testing.T) {
		tests := map[string]string{
			"{test }":           "test",
			"{ test}":           "test",
			"{test  }":          "test",
			"{  \ttest}":        "test",
			"{test}":            "test",
			"{ \n  test  \n\n}": "test",
		}

		for input, expected := range tests {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				arg, ok := result[0].(*PlainArg)
				require.True(t, ok, "Expected PlainArg token", nil)
				assert.Equal(t, expected, arg.Arg)
			})
		}
	})

	t.Run("should maintain exact strings - not affected by variables", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []interface{}
		}{
			{"x{test}", []interface{}{"x", "test"}},
			{"\n{test}", []interface{}{"\n", "test"}},
			{" {test}", []interface{}{" ", "test"}},
			{"x { test}", []interface{}{"x ", "test"}},
			{"x{test} x ", []interface{}{"x", "test", " x "}},
			{"x\n{test}\n", []interface{}{"x\n", "test", "\n"}},
		}

		for _, tc := range testCases {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(tc.input, "\n", "\\n"), "\t", "\\t"), func(t *testing.T) {
				result, err := Parse(tc.input, nil)
				require.NoError(t, err)
				require.Len(t, result, len(tc.expected))
				for i, exp := range tc.expected {
					if expStr, ok := exp.(string); ok {
						// Pattern: "x{test}" -> ["x", "test"]
						// Even indices (0, 2, 4...) are Content tokens
						// Odd indices (1, 3, 5...) are PlainArg token arg values
						if i%2 == 0 {
							content, ok := result[i].(*Content)
							require.True(t, ok, "Expected Content token at index %d", i)
							assert.Equal(t, expStr, content.Value)
						} else {
							arg, ok := result[i].(*PlainArg)
							require.True(t, ok, "Expected PlainArg token at index %d", i)
							assert.Equal(t, expStr, arg.Arg)
						}
					}
				}
			})
		}
	})

	t.Run("should handle extended character literals", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []interface{}
		}{
			{"☺{test}", []interface{}{"☺", "test"}},
			{"中{test}中国话不用彁字。", []interface{}{"中", "test", "中国话不用彁字。"}},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result, err := Parse(tc.input, nil)
				require.NoError(t, err)
				require.Len(t, result, len(tc.expected))
				for i, exp := range tc.expected {
					if expStr, ok := exp.(string); ok {
						// Pattern: "☺{test}" -> ["☺", "test"]
						// Even indices (0, 2, 4...) are Content tokens
						// Odd indices (1, 3, 5...) are PlainArg token arg values
						if i%2 == 0 {
							content, ok := result[i].(*Content)
							require.True(t, ok, "Expected Content token at index %d", i)
							assert.Equal(t, expStr, content.Value)
						} else {
							arg, ok := result[i].(*PlainArg)
							require.True(t, ok, "Expected PlainArg token at index %d", i)
							assert.Equal(t, expStr, arg.Arg)
						}
					}
				}
			})
		}
	})

	t.Run("should not matter if it has html or something in it", func(t *testing.T) {
		result, err := Parse("<div class=\"test\">content: {TEST}</div>", nil)
		require.NoError(t, err)
		require.Len(t, result, 3)

		content1, ok := result[0].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, "<div class=\"test\">content: ", content1.Value)

		arg, ok := result[1].(*PlainArg)
		require.True(t, ok, "Expected PlainArg token", nil)
		assert.Equal(t, "TEST", arg.Arg)

		content2, ok := result[2].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, "</div>", content2.Value)
	})
}

// TestSelect tests parsing of select statements
// TypeScript original code:
//
//	describe('Select', () => {
//	  describe('should be very whitespace agnostic', () => {
//	    const exp = [{
//	      type: 'select',
//	      arg: 'VAR',
//	      cases: [
//	        { key: 'key', tokens: [{ type: 'content', value: 'a' }] },
//	        { key: 'other', tokens: [{ type: 'content', value: 'b' }] }
//	      ]
//	    }];
//	  })
//	})
func TestSelect(t *testing.T) {
	t.Run("should be very whitespace agnostic", func(t *testing.T) {
		testInputs := []string{
			"{VAR,select,key{a}other{b}}",
			"{    VAR   ,    select   ,    key      {a}   other    {b}    }",
			"{ \n   VAR  \n , \n   select  \n\n , \n \n  key \n    \n {a}  \n other \n   {b} \n  \n }",
			"{ \t  VAR  \n , \n\t\r  select  \n\t , \t \n  key \n    \t {a}  \n other \t   {b} \t  \t }",
		}

		for _, input := range testInputs {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), "\r", "\\r"), func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				sel, ok := result[0].(*Select)
				require.True(t, ok, "Expected Select token", nil)
				assert.Equal(t, "VAR", sel.Arg)
				require.Len(t, sel.Cases, 2)
				assert.Equal(t, "key", sel.Cases[0].Key)
				assert.Equal(t, "other", sel.Cases[1].Key)
				require.Len(t, sel.Cases[0].Tokens, 1)
				require.Len(t, sel.Cases[1].Tokens, 1)
				content0, ok := sel.Cases[0].Tokens[0].(*Content)
				require.True(t, ok, "Expected Content in first case", nil)
				assert.Equal(t, "a", content0.Value)
				content1, ok := sel.Cases[1].Tokens[0].(*Content)
				require.True(t, ok, "Expected Content in second case", nil)
				assert.Equal(t, "b", content1.Value)
			})
		}
	})

	t.Run("should allow MessageFormat extension keywords in select keys", func(t *testing.T) {
		testCases := []struct {
			input       string
			expectedKey string
		}{
			{"x {TEST, select, select{a} other{b} }", "select"},
			{"x {TEST, select, offset{a} other{b} }", "offset"},
			{"x {TEST, select, TEST{a} other{b} }", "TEST"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result, err := Parse(tc.input, nil)
				require.NoError(t, err)
				require.Len(t, result, 2)
				content, ok := result[0].(*Content)
				require.True(t, ok, "Expected Content token", nil)
				assert.Equal(t, "x ", content.Value)
				sel, ok := result[1].(*Select)
				require.True(t, ok, "Expected Select token", nil)
				assert.Equal(t, "TEST", sel.Arg)
				assert.Equal(t, tc.expectedKey, sel.Cases[0].Key)
				assert.Equal(t, "other", sel.Cases[1].Key)
			})
		}
	})

	t.Run("should be case-sensitive", func(t *testing.T) {
		caseSensitiveKeys := []string{"Select", "SELECT", "selecT"}
		for _, key := range caseSensitiveKeys {
			input := "{TEST, " + key + ", a{a} other{b}}"
			t.Run(input, func(t *testing.T) {
				// Should throw error for case-sensitive keywords
				_, err := Parse(input, nil)
				require.Error(t, err)
			})
		}
	})

	t.Run("numerical keys", func(t *testing.T) {
		t.Run("should accept numerical keys", func(t *testing.T) {
			assert.NotPanics(t, func() {
				_, _ = Parse("{TEST, select, 0{a} other{b}}", nil) // Explicitly ignore return values in test
			})
		})

		t.Run("should reject = prefixed keys", func(t *testing.T) {
			_, err := Parse("{TEST, select, =0{a} other{b}}", nil)
			require.Error(t, err)
		})
	})
}

// TestPlural tests parsing of plural statements
// TypeScript original code:
//
//	describe('Plurals', () => {
//	  it('should accept a variable, no offset, and plural keys', function () {
//	    expect(function () {
//	      parse('{NUM, plural, one{1} other{2}}');
//	    }).not.toThrow();
//	  });
//	})
func TestPlural(t *testing.T) {
	t.Run("should accept a variable, no offset, and plural keys", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_, _ = Parse("{NUM, plural, one{1} other{2}}", nil)
		})
	})

	t.Run("should accept exact values with = prefixes", func(t *testing.T) {
		result, err := Parse("{NUM, plural, =0{e0} =1{e1} =2{e2} other{2}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		plural, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		assert.Equal(t, "NUM", plural.Arg)
		require.Len(t, plural.Cases, 4)
		assert.Equal(t, "=0", plural.Cases[0].Key)
		assert.Equal(t, "=1", plural.Cases[1].Key)
		assert.Equal(t, "=2", plural.Cases[2].Key)
		assert.Equal(t, "other", plural.Cases[3].Key)

		_, err2 := Parse("{NUM, plural, =a{e1} other{2}}", nil)
		require.Error(t, err2)
	})

	t.Run("should accept the 6 official keywords", func(t *testing.T) {
		result, err := Parse("{NUM, plural, zero{0} one{1} two{2} few{5} many{100} other{101}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		plural, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		require.Len(t, plural.Cases, 6)
		assert.Equal(t, "zero", plural.Cases[0].Key)
		assert.Equal(t, "one", plural.Cases[1].Key)
		assert.Equal(t, "two", plural.Cases[2].Key)
		assert.Equal(t, "few", plural.Cases[3].Key)
		assert.Equal(t, "many", plural.Cases[4].Key)
		assert.Equal(t, "other", plural.Cases[5].Key)
	})

	t.Run("should be gracious with whitespace", func(t *testing.T) {
		expected, err := Parse("{NUM, plural, one{1} other{2}}", nil)
		require.NoError(t, err)
		whitespaceInputs := []string{
			"{ NUM, plural, one{1} other{2} }",
			"{NUM,plural,one{1}other{2}}",
			"{\nNUM,   \nplural,\n   one\n\n{1}\n other {2}\n\n\n}",
			"{\tNUM\t,\t\t\r plural\t\n, \tone\n{1}    other\t\n{2}\n\n\n}",
		}

		for _, input := range whitespaceInputs {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), "\r", "\\r"), func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				assert.Equal(t, len(expected), len(result))
				if len(result) > 0 && len(expected) > 0 {
					if expPlural, ok := expected[0].(*Select); ok {
						if resPlural, ok := result[0].(*Select); ok {
							assert.Equal(t, expPlural.Arg, resPlural.Arg)
							assert.Equal(t, len(expPlural.Cases), len(resPlural.Cases))
						}
					}
				}
			})
		}
	})

	t.Run("Plural offsets", func(t *testing.T) {
		t.Run("should accept a valid offset", func(t *testing.T) {
			offsetInputs := []string{
				"{NUM, plural, offset:4 other{a}}",
				"{NUM,plural,offset:4other{a}}",
				"{NUM, plural, offset:4 other{a}}",
				"{NUM, plural, offset\n\t\r : \t\n\r4 other{a}}",
			}

			for _, input := range offsetInputs {
				t.Run(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), "\r", "\\r"), func(t *testing.T) {
					result, err := Parse(input, nil)
					require.NoError(t, err)
					require.Len(t, result, 1)
					plural, ok := result[0].(*Select)
					require.True(t, ok, "Expected Select token for plural")
					assert.Equal(t, "NUM", plural.Arg)
					require.NotNil(t, plural.PluralOffset)
					assert.Equal(t, 4, *plural.PluralOffset)
					require.Len(t, plural.Cases, 1)
					assert.Equal(t, "other", plural.Cases[0].Key)
				})
			}
		})

		t.Run("should require offset before cases", func(t *testing.T) {
			_, err := Parse("{NUM, plural, other{a} offset:4}", nil)
			require.Error(t, err)
		})
	})

	t.Run("should support quoting", func(t *testing.T) {
		// Test complex quoting with date function and octothorpe
		result, err := Parse("{NUM, plural, one{{x,date,y-M-dd # '#'}} two{two}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		plural, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		require.Len(t, plural.Cases, 2)
		require.Len(t, plural.Cases[0].Tokens, 1)
		fnArg, ok := plural.Cases[0].Tokens[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "x", fnArg.Arg)
		assert.Equal(t, "date", fnArg.Key)
		require.Len(t, fnArg.Param, 3)

		// Test simple octothorpe quoting
		result2, _ := Parse("{NUM, plural, one{# '#'} two{two}}", nil)
		plural2, ok := result2[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		require.Len(t, plural2.Cases[0].Tokens, 2)
		_, isOcto := plural2.Cases[0].Tokens[0].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token", nil)
		content, ok := plural2.Cases[0].Tokens[1].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, " #", content.Value)

		// Test octothorpe at end
		result3, _ := Parse("{NUM, plural, one{one#} two{two}}", nil)
		plural3, ok := result3[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		require.Len(t, plural3.Cases[0].Tokens, 2)
		content, ok = plural3.Cases[0].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, "one", content.Value)
		_, isOcto = plural3.Cases[0].Tokens[1].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token", nil)

		// Test complex nested case with quotes
		result4, _ := Parse("'{' {S, plural, other{# is a '#'}} '}'", nil)
		require.Len(t, result4, 3)
		content1, ok := result4[0].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, "{ ", content1.Value)
		plural4, ok := result4[1].(*Select)
		require.True(t, ok, "Expected Select token", nil)
		assert.Equal(t, "S", plural4.Arg)
		require.Len(t, plural4.Cases, 1)
		assert.Equal(t, "other", plural4.Cases[0].Key)
		require.Len(t, plural4.Cases[0].Tokens, 2)
		_, isOcto = plural4.Cases[0].Tokens[0].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token", nil)
		content, ok = plural4.Cases[0].Tokens[1].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, " is a #", content.Value)
		content2, ok := result4[2].(*Content)
		require.True(t, ok, "Expected Content token", nil)
		assert.Equal(t, " }", content2.Value)
	})

	t.Run("should handle octothorpes with nested plurals", func(t *testing.T) {
		result, err := Parse("{x, plural, one{{y, plural, other{}}} other{#}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		plural, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for plural")
		require.Len(t, plural.Cases, 2)
		require.Len(t, plural.Cases[1].Tokens, 1)
		_, isOcto := plural.Cases[1].Tokens[0].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token in outer plural", nil)
	})
}

// TestSelectordinal tests parsing of selectordinal statements
// TypeScript original code:
//
//	describe('Ordinals', () => {
//	  it('should accept a variable, offset, and keys', () => {
//	    expect(
//	      parse('{NUM, selectordinal, offset:1 one{1} other{2}}')
//	    ).toMatchObject([{
//	      type: 'selectordinal',
//	      arg: 'NUM',
//	      pluralOffset: 1,
//	      cases: [
//	        { key: 'one', tokens: [{ type: 'content', value: '1' }] },
//	        { key: 'other', tokens: [{ type: 'content', value: '2' }] }
//	      ]
//	    }]);
//	  });
//	})
func TestSelectordinal(t *testing.T) {
	t.Run("should accept a variable, offset, and keys", func(t *testing.T) {
		result, err := Parse("{NUM, selectordinal, offset:1 one{1} other{2}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		ordinal, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for selectordinal", nil)
		assert.Equal(t, "NUM", ordinal.Arg)
		require.NotNil(t, ordinal.PluralOffset)
		assert.Equal(t, 1, *ordinal.PluralOffset)
		require.Len(t, ordinal.Cases, 2)
		assert.Equal(t, "one", ordinal.Cases[0].Key)
		assert.Equal(t, "other", ordinal.Cases[1].Key)
		require.Len(t, ordinal.Cases[0].Tokens, 1)
		require.Len(t, ordinal.Cases[1].Tokens, 1)
		content0, ok := ordinal.Cases[0].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content in first case", nil)
		assert.Equal(t, "1", content0.Value)
		content1, ok := ordinal.Cases[1].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content in second case", nil)
		assert.Equal(t, "2", content1.Value)
	})

	t.Run("should accept exact values with = prefixes", func(t *testing.T) {
		result, err := Parse("{NUM, selectordinal, =0{e0} =1{e1} =2{e2} other{2}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		ordinal, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for selectordinal", nil)
		require.Len(t, ordinal.Cases, 4)
		assert.Equal(t, "=0", ordinal.Cases[0].Key)
		assert.Equal(t, "=1", ordinal.Cases[1].Key)
		assert.Equal(t, "=2", ordinal.Cases[2].Key)
		assert.Equal(t, "other", ordinal.Cases[3].Key)

		_, err2 := Parse("{NUM, selectordinal, =a{e1} other{2}}", nil)
		require.Error(t, err2)
	})
}

// TestFunctions tests parsing of function calls
// TypeScript original code:
//
//	describe('Functions', function () {
//	  it('should allow upper-case type, except for built-ins', function () {
//	    for (const date of ['date', 'Date', '9ate']) {
//	      expect(parse(`{var,${date}}`)).toMatchObject([
//	        { type: 'function', arg: 'var', key: date }
//	      ]);
//	    }
//	  });
//	})
func TestFunctions(t *testing.T) {
	t.Run("should allow upper-case type, except for built-ins", func(t *testing.T) {
		dateKeys := []string{"date", "Date", "9ate"}
		for _, dateKey := range dateKeys {
			input := "{var," + dateKey + "}"
			t.Run(input, func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				func_, ok := result[0].(*FunctionArg)
				require.True(t, ok, "Expected FunctionArg token", nil)
				assert.Equal(t, "var", func_.Arg)
				assert.Equal(t, dateKey, func_.Key)
			})
		}

		_, err := Parse("{var,Select}", nil)
		require.Error(t, err)
	})

	t.Run("should be gracious with whitespace around arg and key", func(t *testing.T) {
		whitespaceInputs := []string{
			"{var,date}",
			"{var, date}",
			"{ var, date }",
			"{\nvar,   \ndate\n}",
		}

		for _, input := range whitespaceInputs {
			t.Run(strings.ReplaceAll(strings.ReplaceAll(input, "\n", "\\n"), "\t", "\\t"), func(t *testing.T) {
				result, err := Parse(input, nil)
				require.NoError(t, err)
				require.Len(t, result, 1)
				func_, ok := result[0].(*FunctionArg)
				require.True(t, ok, "Expected FunctionArg token", nil)
				assert.Equal(t, "var", func_.Arg)
				assert.Equal(t, "date", func_.Key)
			})
		}
	})

	t.Run("should accept parameters", func(t *testing.T) {
		result, err := Parse("{var,date,long}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "var", func_.Arg)
		assert.Equal(t, "date", func_.Key)
		require.Len(t, func_.Param, 1)
		content, ok := func_.Param[0].(*Content)
		require.True(t, ok, "Expected Content token in param", nil)
		assert.Equal(t, "long", content.Value)

		result2, _ := Parse("{var,date,long,short}", nil)
		func2, ok := result2[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		require.Len(t, func2.Param, 1)
		content, ok = func2.Param[0].(*Content)
		require.True(t, ok, "Expected Content token in param", nil)
		assert.Equal(t, "long,short", content.Value)
	})

	t.Run("should accept parameters with whitespace", func(t *testing.T) {
		result, _ := Parse("{var,date,y-M-d HH:mm:ss zzzz}", nil)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d HH:mm:ss zzzz", func_.Param[0].(*Content).Value)

		result2, err := Parse("{var,date,   y-M-d HH:mm:ss zzzz    }", nil)
		require.NoError(t, err)
		func2, ok := result2[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "   y-M-d HH:mm:ss zzzz    ", func2.Param[0].(*Content).Value)
	})

	t.Run("should accept parameters with special characters", func(t *testing.T) {
		result, err := Parse("{var,date,y-M-d '{,}' '' HH:mm:ss zzzz}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d {,} ' HH:mm:ss zzzz", func_.Param[0].(*Content).Value)

		result2, _ := Parse("{var,date,y-M-d '{,}' '' HH:mm:ss zzzz'}'}", nil)
		func2, ok := result2[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d {,} ' HH:mm:ss zzzz}", func2.Param[0].(*Content).Value)

		result3, _ := Parse("{var,date,y-M-d # HH:mm:ss zzzz}", nil)
		func3, ok := result3[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d # HH:mm:ss zzzz", func3.Param[0].(*Content).Value)

		result4, _ := Parse("{var,date,y-M-d '#' HH:mm:ss zzzz}", nil)
		func4, ok := result4[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d '#' HH:mm:ss zzzz", func4.Param[0].(*Content).Value)

		result5, _ := Parse("{var,date,y-M-d, HH:mm:ss zzzz}", nil)
		func5, ok := result5[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "y-M-d, HH:mm:ss zzzz", func5.Param[0].(*Content).Value)
	})

	t.Run("should accept parameters containing a basic variable", func(t *testing.T) {
		result, err := Parse("{foo, date, {bar}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "foo", func_.Arg)
		assert.Equal(t, "date", func_.Key)
		require.Len(t, func_.Param, 2)
		content, ok := func_.Param[0].(*Content)
		require.True(t, ok, "Expected Content token in param", nil)
		assert.Equal(t, " ", content.Value)
		arg, ok := func_.Param[1].(*PlainArg)
		require.True(t, ok, "Expected PlainArg token in param", nil)
		assert.Equal(t, "bar", arg.Arg)
	})

	t.Run("should accept parameters containing a select", func(t *testing.T) {
		result, err := Parse("{foo, date, {bar, select, other{baz}}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "foo", func_.Arg)
		assert.Equal(t, "date", func_.Key)
		require.Len(t, func_.Param, 2)
		content, ok := func_.Param[0].(*Content)
		require.True(t, ok, "Expected Content token in param", nil)
		assert.Equal(t, " ", content.Value)
		sel, ok := func_.Param[1].(*Select)
		require.True(t, ok, "Expected Select token in param", nil)
		assert.Equal(t, "bar", sel.Arg)
		require.Len(t, sel.Cases, 1)
		assert.Equal(t, "other", sel.Cases[0].Key)
		require.Len(t, sel.Cases[0].Tokens, 1)
		content, ok = sel.Cases[0].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content token in select case", nil)
		assert.Equal(t, "baz", content.Value)
	})

	t.Run("should accept parameters containing a plural", func(t *testing.T) {
		result, err := Parse("{foo, date, {bar, plural, other{#}}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		func_, ok := result[0].(*FunctionArg)
		require.True(t, ok, "Expected FunctionArg token", nil)
		assert.Equal(t, "foo", func_.Arg)
		assert.Equal(t, "date", func_.Key)
		require.Len(t, func_.Param, 2)
		content, ok := func_.Param[0].(*Content)
		require.True(t, ok, "Expected Content token in param", nil)
		assert.Equal(t, " ", content.Value)
		plural, ok := func_.Param[1].(*Select)
		require.True(t, ok, "Expected Select token (plural) in param", nil)
		assert.Equal(t, "bar", plural.Arg)
		require.Len(t, plural.Cases, 1)
		assert.Equal(t, "other", plural.Cases[0].Key)
		require.Len(t, plural.Cases[0].Tokens, 1)
		_, isOcto := plural.Cases[0].Tokens[0].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token in plural case", nil)
	})
}

// TestNestedBlocks tests parsing of nested select and plural statements
// TypeScript original code:
//
//	describe('Nested blocks', function () {
//	  it('should allow nested select statements', function () {
//	    expect(
//	      parse(
//	        '{NUM1, select, other{{NUM2, select, one{a} other{{NUM3, select, other{b}}}}}}'
//	      )
//	    ).toMatchObject([{
//	      arg: 'NUM1',
//	      cases: [{
//	        tokens: [{
//	          arg: 'NUM2',
//	          cases: [
//	            { key: 'one', tokens: [{ value: 'a' }] },
//	            {
//	              key: 'other',
//	              tokens: [
//	                { arg: 'NUM3', cases: [{ tokens: [{ value: 'b' }] }] }
//	              ]
//	            }
//	          ]
//	        }]
//	      }]
//	    }]);
//	  });
//	})
func TestNestedBlocks(t *testing.T) {
	t.Run("should allow nested select statements", func(t *testing.T) {
		result, err := Parse("{NUM1, select, other{{NUM2, select, one{a} other{{NUM3, select, other{b}}}}}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		select1, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for NUM1", nil)
		assert.Equal(t, "NUM1", select1.Arg)
		require.Len(t, select1.Cases, 1)
		assert.Equal(t, "other", select1.Cases[0].Key)
		require.Len(t, select1.Cases[0].Tokens, 1)

		select2, ok := select1.Cases[0].Tokens[0].(*Select)
		require.True(t, ok, "Expected nested Select token for NUM2", nil)
		assert.Equal(t, "NUM2", select2.Arg)
		require.Len(t, select2.Cases, 2)
		assert.Equal(t, "one", select2.Cases[0].Key)
		assert.Equal(t, "other", select2.Cases[1].Key)

		require.Len(t, select2.Cases[1].Tokens, 1)
		select3, ok := select2.Cases[1].Tokens[0].(*Select)
		require.True(t, ok, "Expected deeply nested Select token for NUM3", nil)
		assert.Equal(t, "NUM3", select3.Arg)
		require.Len(t, select3.Cases, 1)
		assert.Equal(t, "other", select3.Cases[0].Key)
		require.Len(t, select3.Cases[0].Tokens, 1)
		content, ok := select3.Cases[0].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content token in deepest level", nil)
		assert.Equal(t, "b", content.Value)
	})

	t.Run("should allow nested plural statements", func(t *testing.T) {
		result, err := Parse("{NUM1, plural, other{{NUM2, plural, offset:1 one{#} other{{NUM3, plural, other{b}}}}}}", nil)
		require.NoError(t, err)
		require.Len(t, result, 1)
		plural1, ok := result[0].(*Select)
		require.True(t, ok, "Expected Select token for NUM1", nil)
		assert.Equal(t, "NUM1", plural1.Arg)
		require.Len(t, plural1.Cases, 1)
		assert.Equal(t, "other", plural1.Cases[0].Key)
		require.Len(t, plural1.Cases[0].Tokens, 1)

		plural2, ok := plural1.Cases[0].Tokens[0].(*Select)
		require.True(t, ok, "Expected nested Select token for NUM2", nil)
		assert.Equal(t, "NUM2", plural2.Arg)
		require.NotNil(t, plural2.PluralOffset)
		assert.Equal(t, 1, *plural2.PluralOffset)
		require.Len(t, plural2.Cases, 2)
		assert.Equal(t, "one", plural2.Cases[0].Key)
		assert.Equal(t, "other", plural2.Cases[1].Key)

		require.Len(t, plural2.Cases[0].Tokens, 1)
		_, isOcto := plural2.Cases[0].Tokens[0].(*Octothorpe)
		assert.True(t, isOcto, "Expected Octothorpe token in nested plural", nil)

		require.Len(t, plural2.Cases[1].Tokens, 1)
		plural3, ok := plural2.Cases[1].Tokens[0].(*Select)
		require.True(t, ok, "Expected deeply nested Select token for NUM3", nil)
		assert.Equal(t, "NUM3", plural3.Arg)
		require.Len(t, plural3.Cases, 1)
		assert.Equal(t, "other", plural3.Cases[0].Key)
		require.Len(t, plural3.Cases[0].Tokens, 1)
		content, ok := plural3.Cases[0].Tokens[0].(*Content)
		require.True(t, ok, "Expected Content token in deepest level", nil)
		assert.Equal(t, "b", content.Value)
	})
}

// TestErrors tests parsing error conditions
// TypeScript original code:
//
//	describe('Errors', () => {
//	  describe('Should require matched braces', () => {
//	    const expectedError = /invalid syntax|Unexpected message end/;
//	    it('{foo', () => {
//	      expect(() => parse('{foo')).toThrow(expectedError);
//	    });
//	  });
//	})
func TestErrors(t *testing.T) {
	t.Run("Should require matched braces", func(t *testing.T) {
		unmatchedInputs := []string{
			"{foo",
			"{foo,",
			"{foo,bar",
			"{foo,bar,",
			"{foo, date, {bar{}",
		}

		for _, input := range unmatchedInputs {
			t.Run(input, func(t *testing.T) {
				_, err := Parse(input, nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Unexpected message end")
			})
		}
	})

	t.Run("should not allow an offset for selects", func(t *testing.T) {
		_, err := Parse("{NUM, select, offset:1 test { 1 } test2 { 2 }}", nil)
		require.Error(t, err)
	})

	t.Run("strictPluralKeys", func(t *testing.T) {
		t.Run("should not allow invalid keys for plurals by default", func(t *testing.T) {
			_, err := Parse("{NUM, plural, one { 1 } invalid { error } other { 2 }}", nil)
			require.Error(t, err)

			_, err = Parse("{NUM, plural, one { 1 } some { error } other { 2 }}", &ParseOptions{
				Cardinal: []PluralCategory{"one", "other"},
			})
			require.Error(t, err)
		})

		t.Run("should allow invalid keys for plurals if strictPluralKeys is false", func(t *testing.T) {
			strictPluralKeys := false
			_, err := Parse("{NUM, plural, one { 1 } invalid { error } other { 2 }}", &ParseOptions{
				StrictPluralKeys: &strictPluralKeys,
			})
			require.NoError(t, err)

			_, err = Parse("{NUM, plural, one { 1 } some { error } other { 2 }}", &ParseOptions{
				Cardinal:         []PluralCategory{"one", "other"},
				StrictPluralKeys: &strictPluralKeys,
			})
			require.NoError(t, err)
		})
	})

	t.Run("should not allow invalid keys for selectordinals", func(t *testing.T) {
		_, err := Parse("{NUM, selectordinal, one { 1 } invalid { error } other { 2 }}", nil)
		require.Error(t, err)

		_, err = Parse("{NUM, selectordinal, one { 1 } some { error } other { 2 }}", &ParseOptions{
			Ordinal: []PluralCategory{"one", "other"},
		})
		require.Error(t, err)
	})

	t.Run("should allow an offset for selectordinals", func(t *testing.T) {
		_, err := Parse("{NUM, selectordinal, offset:1 one { 1 } other { 2 }}", nil)
		require.NoError(t, err)
	})

	t.Run("should allow characters in variables that are valid ICU identifiers", func(t *testing.T) {
		_, err := Parse("{ű\u3000á}", nil)
		require.NoError(t, err)
	})

	t.Run("should allow positional variables", func(t *testing.T) {
		_, err := Parse("{0}", nil)
		require.NoError(t, err)
	})

	t.Run("should throw errors on negative offsets", func(t *testing.T) {
		_, err := Parse("{NUM, plural, offset:-4 other{a}}", nil)
		require.Error(t, err)
	})

	t.Run("should require closing bracket", func(t *testing.T) {
		_, err := Parse("{count, plural, one {car} other {cars}", nil)
		require.Error(t, err)
	})

	t.Run("should complain about unnecessarily quoted #{ outside plural", func(t *testing.T) {
		_, err := Parse("foo '#{' bar", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unsupported escape pattern")
	})
}
