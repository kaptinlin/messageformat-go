package functions

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFunctions(t *testing.T) {
	defaults := DefaultFunctionMap()
	assert.ElementsMatch(t, []string{
		"currency",
		"integer",
		"number",
		"offset",
		"percent",
		"string",
	}, slices.Collect(maps.Keys(defaults)))
}

func TestDraftFunctions(t *testing.T) {
	drafts := DraftFunctionMap()
	assert.ElementsMatch(t, []string{
		"date",
		"datetime",
		"time",
		"unit",
	}, slices.Collect(maps.Keys(drafts)))
}

func TestFunctionMapsReturnSnapshots(t *testing.T) {
	defaults := DefaultFunctionMap()
	drafts := DraftFunctionMap()

	delete(defaults, "string")
	defaults["custom"] = StringFunction
	delete(drafts, "date")
	drafts["custom"] = StringFunction

	freshDefaults := DefaultFunctionMap()
	freshDrafts := DraftFunctionMap()
	assert.Contains(t, freshDefaults, "string")
	assert.NotContains(t, freshDefaults, "custom")
	assert.Contains(t, freshDrafts, "date")
	assert.NotContains(t, freshDrafts, "custom")
}
