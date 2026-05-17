package mfwg

import (
	"os"
	"path/filepath"
	"testing"

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
			"expErrors": [{"type": "bad-operand"}]
		},
		"tests": [
			{
				"src": "hello {$name}",
				"exp": "hello Ada",
				"params": [{"name": "name", "value": "Ada"}],
				"tags": ["smoke"],
				"only": true,
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
	assert.Equal(t, "default", first.BidiIsolationRaw)
	assert.True(t, IsOnlyTest(first))
	assert.True(t, HasTag(first, "smoke"))
	assert.True(t, ShouldSkip(first, map[string]bool{"smoke": true}))
	assert.False(t, ShouldSkip(first, map[string]bool{"slow": true}))

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
	assert.False(t, HasTag(second, "smoke"))
	assert.False(t, IsOnlyTest(second))
	assert.Empty(t, ExpectedErrors(second))

	wantParts := []any{map[string]any{"type": "text", "value": "explicit"}}
	if diff := cmp.Diff(wantParts, ExpectedParts(second)); diff != "" {
		t.Fatalf("parts mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, HasOnlyTests(scenarios[0]))
	if diff := cmp.Diff([]Test{first}, FilterOnlyTests(scenarios[0])); diff != "" {
		t.Fatalf("only tests mismatch (-want +got):\n%s", diff)
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
	assert.Equal(t, false, scenarios[0].Tests[0].BidiIsolationRaw)
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
			name: "data model error map",
			test: Test{ExpErrors: map[string]any{"type": "data-model-error"}},
			want: TestTypeDataModelError,
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
	assert.False(t, HasOnlyTests(TestScenario{}))
	assert.Empty(t, FilterOnlyTests(TestScenario{}))
}
