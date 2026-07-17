package datamodel_test

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseMessagePreservesRootAndVariantSpans proves public parsed nodes retain CST provenance.
// TypeScript original code:
// return { ...message, [cstKey]: msg };
func TestParseMessagePreservesRootAndVariantSpans(t *testing.T) {
	t.Parallel()

	patternSource := ".input {$name}\n{{Hello}}"
	pattern, err := datamodel.ParseMessage(patternSource)
	require.NoError(t, err)
	patternMessage, ok := pattern.(*datamodel.PatternMessage)
	require.True(t, ok, "got %T", pattern)
	assertNodeSpan(t, patternMessage, 0, len(patternSource))

	selectSource := ".input {$count :number}\n.match $count\none {{One}}\n* {{Other}}"
	selected, err := datamodel.ParseMessage(selectSource)
	require.NoError(t, err)
	selectMessage, ok := selected.(*datamodel.SelectMessage)
	require.True(t, ok, "got %T", selected)
	assertNodeSpan(t, selectMessage, 0, len(selectSource))

	variants := selectMessage.Variants()
	require.Len(t, variants, 2)
	firstStart := strings.Index(selectSource, "one {{One}}")
	secondStart := strings.Index(selectSource, "* {{Other}}")
	assertNodeSpan(t, variants[0], firstStart, firstStart+len("one {{One}}"))
	assertNodeSpan(t, variants[1], secondStart, secondStart+len("* {{Other}}"))
}

// TestVariantValidationErrorUsesParsedSpan proves variant-owned model errors retain source provenance.
// TypeScript original code:
// onError(new MessageDataModelError('key-mismatch', variant));
func TestVariantValidationErrorUsesParsedSpan(t *testing.T) {
	t.Parallel()

	source := ".input {$count :number}\n.match $count\none extra {{Bad}}\n* {{Fallback}}"
	message, err := datamodel.ParseMessage(source)
	require.NoError(t, err)

	_, err = datamodel.ValidateMessage(message, nil)
	require.Error(t, err)
	var modelErr *pkgerrors.MessageDataModelError
	require.True(t, stderrors.As(err, &modelErr))
	start := strings.Index(source, "one extra {{Bad}}")
	assert.Equal(t, start, modelErr.Start)
	assert.Equal(t, start+len("one extra {{Bad}}"), modelErr.End)
}

// TestProgrammaticCompositeNodesUseUnknownSpans characterizes provenance for caller-built models.
// TypeScript original code:
// const position = node[cstKey] ?? { start: -1, end: -1 };
func TestProgrammaticCompositeNodesUseUnknownSpans(t *testing.T) {
	t.Parallel()

	pattern := mustPattern(t, nil)
	message := mustPatternMessage(t, nil, pattern, "")
	variant, err := datamodel.NewVariant(nil, pattern)
	require.NoError(t, err)

	assertNodeSpan(t, message, -1, -1)
	assertNodeSpan(t, *variant, -1, -1)
}

// assertNodeSpan compares one public model node's source span.
// TypeScript original code:
// expect(node[cstKey]).toMatchObject({ start, end });
func assertNodeSpan(t *testing.T, node interface {
	GetPosition() (start, end int)
}, wantStart, wantEnd int) {
	t.Helper()

	start, end := node.GetPosition()
	assert.Equal(t, wantStart, start)
	assert.Equal(t, wantEnd, end)
}
