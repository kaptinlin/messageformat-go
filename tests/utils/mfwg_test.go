package mfwg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenariosLoadsJSONAndAppliesDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("not json"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.json"), []byte(`{
		"scenario": "defaults",
		"defaultTestProperties": {
			"locale": "fr",
			"bidiIsolation": "default",
			"expCleanSrc": "clean default",
			"expErrors": [{"type": "bad-operand"}]
		},
		"tests": [
			{
				"src": "hello {$name}",
				"exp": "hello Ada",
				"params": [{"name": "name", "value": "Ada"}],
				"tags": ["smoke"],
				"description": "formats a name"
			},
			{
				"src": "explicit",
				"locale": "en",
				"bidiIsolation": false,
				"expErrors": false,
				"expParts": [{"type": "text", "value": "explicit"}]
			}
		]
	}`), 0o600))

	scenarios, err := TestScenarios(dir)
	require.NoError(t, err)
	require.Len(t, scenarios, 1)
	assert.Equal(t, "defaults", scenarios[0].Scenario)

	cases := TestCases(scenarios[0])
	require.Len(t, cases, 2)

	first := cases[0]
	assert.Equal(t, "formats a name", TestName(first))
	assert.Equal(t, "fr", first.Locale)
	assert.True(t, GetBidiIsolation(first))
	require.NotNil(t, first.ExpCleanSrc)
	assert.Equal(t, "clean default", *first.ExpCleanSrc)
	assert.Contains(t, first.Tags, "smoke")

	wantParams := map[string]any{"name": "Ada"}
	if diff := cmp.Diff(wantParams, first.GetParamsMap()); diff != "" {
		t.Fatalf("params mismatch (-want +got):\n%s", diff)
	}

	gotString, ok := ExpectedString(first)
	require.True(t, ok)
	assert.Equal(t, "hello Ada", gotString)

	wantErrors := []map[string]any{{"type": "bad-operand"}}
	if diff := cmp.Diff(wantErrors, ExpectedErrors(first)); diff != "" {
		t.Fatalf("errors mismatch (-want +got):\n%s", diff)
	}

	second := cases[1]
	assert.Equal(t, "en", second.Locale)
	assert.False(t, GetBidiIsolation(second))
	assert.NotContains(t, second.Tags, "smoke")
	assert.Empty(t, ExpectedErrors(second))

	wantParts := []any{map[string]any{"type": "text", "value": "explicit"}}
	if diff := cmp.Diff(wantParts, ExpectedParts(second)); diff != "" {
		t.Fatalf("parts mismatch (-want +got):\n%s", diff)
	}
}

func TestScenariosAppliesBooleanBidiDefault(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.json"), []byte(`{
		"scenario": "bidi",
		"defaultTestProperties": {"bidiIsolation": false},
		"tests": [{"src": "plain", "exp": "plain"}]
	}`), 0o600))

	scenarios, err := TestScenarios(dir)
	require.NoError(t, err)
	require.Len(t, scenarios, 1)
	require.Len(t, scenarios[0].Tests, 1)
	assert.False(t, GetBidiIsolation(scenarios[0].Tests[0]))
}

func TestScenariosMergesAllDefaultProperties(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.json"), []byte(`{
		"scenario": "all defaults",
		"defaultTestProperties": {
			"src": "default source",
			"locale": "fr",
			"params": [{"name": "name", "value": "Ada"}],
			"tags": ["default-tag"],
			"exp": "default output",
			"expParts": [{"type": "literal", "value": "default output"}],
			"expErrors": [{"type": "bad-operand"}],
			"bidiIsolation": "default"
		},
		"tests": [
			{},
			{
				"src": "",
				"locale": "",
				"params": [],
				"tags": [],
				"exp": "",
				"expParts": [],
				"expErrors": false,
				"bidiIsolation": false
			}
		]
	}`), 0o600))

	scenarios, err := TestScenarios(dir)
	require.NoError(t, err)
	require.Len(t, scenarios, 1)
	require.Len(t, scenarios[0].Tests, 2)

	inherited := scenarios[0].Tests[0]
	assert.Equal(t, "default source", inherited.Src)
	assert.Equal(t, "fr", inherited.Locale)
	assert.Equal(t, map[string]any{"name": "Ada"}, inherited.GetParamsMap())
	assert.Equal(t, []string{"default-tag"}, inherited.Tags)
	assert.Equal(t, "default output", inherited.Exp)
	assert.Len(t, inherited.ExpParts, 1)
	assert.Len(t, ExpectedErrors(inherited), 1)
	assert.True(t, GetBidiIsolation(inherited))

	overridden := scenarios[0].Tests[1]
	assert.Empty(t, overridden.Src)
	assert.Empty(t, overridden.Locale)
	assert.Empty(t, overridden.Params)
	assert.Empty(t, overridden.Tags)
	assert.Equal(t, "", overridden.Exp)
	assert.Empty(t, overridden.ExpParts)
	assert.Empty(t, ExpectedErrors(overridden))
	assert.False(t, GetBidiIsolation(overridden))
}

func TestScenariosConvertsDatetimeParams(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "datetime.json"), []byte(`{
		"scenario": "datetime params",
		"defaultTestProperties": {
			"src": "{$when :datetime}",
			"locale": "en",
			"params": [{"name": "when", "type": "datetime", "value": "2020-01-02T03:04:05Z"}]
		},
		"tests": [{}]
	}`), 0o600))

	scenarios, err := TestScenarios(dir)
	require.NoError(t, err)
	require.Len(t, scenarios, 1)
	require.Len(t, scenarios[0].Tests, 1)

	params := scenarios[0].Tests[0].GetParamsMap()
	assert.Equal(t, time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC), params["when"])
}

func TestScenariosAssignsStablePathIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested"), 0o700))
	fixture := []byte(`{
		"tests": [{"src": "same", "locale": "en", "description": "duplicate"}]
	}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nested", "a.json"), fixture, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nested", "b.json"), fixture, 0o600))

	scenarios, err := TestScenarios(dir)
	require.NoError(t, err)
	require.Len(t, scenarios, 2)

	assert.Equal(t, "nested/a.json", scenarios[0].Path)
	assert.Equal(t, "nested/a.json", scenarios[0].Scenario)
	assert.Equal(t, "nested/a.json#0", scenarios[0].Tests[0].ID())
	assert.Equal(t, "nested/b.json#0", scenarios[1].Tests[0].ID())
	assert.NotEqual(t, scenarios[0].Tests[0].ID(), scenarios[1].Tests[0].ID())
}

func TestScenariosReportsErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{`), 0o600))

		_, err := TestScenarios(dir)
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to walk test directory")
		assert.ErrorContains(t, err, "failed to parse JSON file")
	})

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()

		_, err := TestScenarios(filepath.Join(t.TempDir(), "missing"))
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to walk test directory")
	})

	t.Run("only marker", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "only.json"), []byte(`{
			"scenario": "focused",
			"tests": [{"src": "plain", "locale": "en", "only": true}]
		}`), 0o600))

		_, err := TestScenarios(dir)
		require.Error(t, err)
		assert.ErrorContains(t, err, "only.json#0")
		assert.ErrorContains(t, err, "only: true")
	})
}

func TestGetTestType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		test Test
		want TestType
	}{
		{
			name: "syntax error slice",
			test: Test{ExpErrors: []any{map[string]any{"type": "syntax-error"}}},
			want: TestTypeSyntaxError,
		},
		{
			name: "data model error subtype",
			test: Test{ExpErrors: []any{map[string]any{"type": "duplicate-declaration"}}},
			want: TestTypeDataModelError,
		},
		{
			name: "syntax error takes precedence over data model error",
			test: Test{ExpErrors: []any{
				map[string]any{"type": "duplicate-declaration"},
				map[string]any{"type": "syntax-error"},
			}},
			want: TestTypeSyntaxError,
		},
		{
			name: "runtime error is format test",
			test: Test{ExpErrors: []any{map[string]any{"type": "bad-option"}}},
			want: TestTypeFormat,
		},
		{
			name: "true error marker is format test",
			test: Test{ExpErrors: true},
			want: TestTypeFormat,
		},
		{
			name: "false error marker is format test",
			test: Test{ExpErrors: false},
			want: TestTypeFormat,
		},
		{
			name: "missing expected errors is format test",
			test: Test{},
			want: TestTypeFormat,
		},
		{
			name: "unknown error shape is format test",
			test: Test{ExpErrors: "anything"},
			want: TestTypeFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, GetTestType(tc.test))
		})
	}
}

func TestExpectedHelpers(t *testing.T) {
	t.Parallel()

	assert.Nil(t, Test{}.GetParamsMap())

	name := TestName(Test{
		Src:    "{$name}",
		Locale: "en",
		Params: []TestParam{{Name: "name", Value: "Ada"}},
	})
	assert.Contains(t, name, "src: {$name}")
	assert.Contains(t, name, "locale: en")
	assert.Contains(t, name, "params:")

	_, ok := ExpectedString(Test{Exp: 1})
	assert.False(t, ok)

	wantErrors := []map[string]any{{"type": "unknown"}}
	if diff := cmp.Diff(wantErrors, ExpectedErrors(Test{ExpErrors: true})); diff != "" {
		t.Fatalf("errors mismatch (-want +got):\n%s", diff)
	}

	assert.Nil(t, ExpectedErrors(Test{}))
	assert.True(t, GetBidiIsolation(Test{}))
}

func TestCasesReturnsDetachedSlice(t *testing.T) {
	t.Parallel()

	scenario := TestScenario{Tests: []Test{{Src: "original"}}}
	cases := TestCases(scenario)
	require.Len(t, cases, 1)
	cases[0].Src = "changed"

	assert.Equal(t, "original", scenario.Tests[0].Src)
}
