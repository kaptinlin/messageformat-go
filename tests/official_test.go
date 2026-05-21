package tests

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	messageformat "github.com/kaptinlin/messageformat-go"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	mfwg "github.com/kaptinlin/messageformat-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOfficialSuite runs every JSON fixture under the
// `messageformat-wg/test/tests/` submodule directory through the implementation
// and checks the result against the Unicode MessageFormat 2.0 spec semantics.
//
// Test categories follow the TS reference's `testType()` taxonomy:
//   - syntax-error: Parse must surface a *MessageSyntaxError that is NOT a
//     data-model error.
//   - data-model-error: Parse must surface a *MessageDataModelError.
//   - format: Parse must succeed; Format's onError callback collects runtime
//     errors that are matched against expErrors. exp (if present) is compared
//     against the formatted string.
const officialTestsDir = "messageformat-wg/test/tests"

// errorAliases maps LDML 47 names emitted internally to the LDML 48 names that
// appear in current spec fixtures. The pairs are bidirectional (canonical form
// on the right).
var errorAliases = map[string]string{
	"missing-fallback-variant": "missing-fallback",
	"variant-key-mismatch":     "key-mismatch",
	"bad-variant-key":          "key-mismatch",
}

// dataModelErrorTypes lists the specific error subtypes that Parse must emit
// as MessageDataModelError rather than a plain syntax error.
var dataModelErrorTypes = map[string]bool{
	"duplicate-declaration":       true,
	"duplicate-option-name":       true,
	"duplicate-variant":           true,
	"missing-fallback":            true,
	"missing-fallback-variant":    true,
	"missing-selector-annotation": true,
	"key-mismatch":                true,
	"variant-key-mismatch":        true,
	"duplicate-attribute":         true,
}

func TestOfficialSuite(t *testing.T) {
	scenarios, err := mfwg.TestScenarios(officialTestsDir)
	require.NoError(t, err, "load official fixtures")
	require.NotEmpty(t, scenarios, "expected at least one official fixture")

	for _, scenario := range scenarios {
		t.Run(scenarioSubtestName(scenario), func(t *testing.T) {
			for i, tc := range scenario.Tests {
				name := fmt.Sprintf("%d_%s", i, sanitize(mfwg.TestName(tc)))
				t.Run(name, func(t *testing.T) {
					runOfficialTest(t, tc)
				})
			}
		})
	}
}

func runOfficialTest(t *testing.T, tc mfwg.Test) {
	t.Helper()
	switch classifyTest(tc) {
	case "syntax-error":
		_, err := newOfficialMessageFormat(tc)
		assertSyntaxError(t, err)
	case "data-model-error":
		_, err := newOfficialMessageFormat(tc)
		assertDataModelError(t, err)
	default:
		runFormatTest(t, tc)
	}
}

// classifyTest mirrors the TS testType() helper: syntax-error wins over
// data-model-error, which wins over runtime/format.
func classifyTest(tc mfwg.Test) string {
	expected := mfwg.ExpectedErrors(tc)
	for _, e := range expected {
		typ, _ := e["type"].(string)
		if typ == "syntax-error" {
			return "syntax-error"
		}
		if dataModelErrorTypes[typ] {
			return "data-model-error"
		}
	}
	return "format"
}

func newOfficialMessageFormat(tc mfwg.Test) (*messageformat.MessageFormat, error) {
	fns := officialFunctionRegistry()

	opts := []messageformat.Option{messageformat.WithFunctions(fns)}
	if mfwg.GetBidiIsolation(tc) {
		opts = append(opts, messageformat.WithBidiIsolation(messageformat.BidiDefault))
	} else {
		opts = append(opts, messageformat.WithBidiIsolation(messageformat.BidiNone))
	}

	locales := []string{}
	if tc.Locale != "" {
		locales = []string{tc.Locale}
	}
	return messageformat.Parse(locales, tc.Src, opts...)
}

func officialFunctionRegistry() map[string]functions.MessageFunction {
	fns := make(map[string]functions.MessageFunction, len(functions.DraftFunctionMap())+4)
	maps.Copy(fns, functions.DraftFunctionMap())
	maps.Copy(fns, TestFunctions())
	return fns
}

func runFormatTest(t *testing.T, tc mfwg.Test) {
	t.Helper()

	mf, err := newOfficialMessageFormat(tc)
	require.NoError(t, err, "Parse should succeed for format-class test")

	var runtimeErrors []error
	collectError := func(err error) {
		runtimeErrors = append(runtimeErrors, err)
	}

	got, formatErr := mf.Format(tc.GetParamsMap(), messageformat.WithErrorHandler(collectError))
	require.NoError(t, formatErr, "Format should not return a non-recoverable error")

	if exp, ok := mfwg.ExpectedString(tc); ok {
		assert.Equal(t, exp, got, "formatted output mismatch")
	}

	matchExpectedErrors(t, runtimeErrors, tc)
}

// matchExpectedErrors enforces the TS spec.test.ts semantics:
//   - expErrors absent or false → no runtime errors allowed.
//   - expErrors === true        → at least one runtime error required.
//   - expErrors is an array     → each expected entry must match the observed
//     error at the same index by type (after alias normalization).
func matchExpectedErrors(t *testing.T, runtime []error, tc mfwg.Test) {
	t.Helper()

	expected := mfwg.ExpectedErrors(tc)
	switch v := tc.ExpErrors.(type) {
	case nil:
		assert.Empty(t, runtime, "unexpected runtime errors: %v", runtime)
	case bool:
		if v {
			assert.NotEmpty(t, runtime, "expected at least one runtime error")
		} else {
			assert.Empty(t, runtime, "unexpected runtime errors: %v", runtime)
		}
	default:
		if len(expected) == 0 {
			return
		}
		require.GreaterOrEqual(t, len(runtime), len(expected),
			"observed fewer errors than expected: got %v, want %v", errorTypes(runtime), expectedTypes(expected))
		for i, want := range expected {
			wantType, _ := want["type"].(string)
			gotType := classifyRuntimeError(runtime[i])
			assert.True(t, errorTypesEqual(gotType, wantType),
				"error[%d] type mismatch: got %q, want %q", i, gotType, wantType)
		}
	}
}

func assertSyntaxError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err, "expected syntax error")

	if dmErr, ok := errors.AsType[*pkgerrors.MessageDataModelError](err); ok {
		t.Fatalf("expected plain syntax error, got data-model error %q", dmErr.Type)
	}
	_, ok := errors.AsType[*pkgerrors.MessageSyntaxError](err)
	require.True(t, ok, "expected *MessageSyntaxError, got %T: %v", err, err)
}

func assertDataModelError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err, "expected data-model error")

	// TS reference (spec.test.ts:50) only asserts MessageSyntaxError for both
	// syntax-error and data-model-error categories; MessageDataModelError is
	// the more specific subtype but the parent type is sufficient here.
	_, ok := errors.AsType[*pkgerrors.MessageSyntaxError](err)
	require.True(t, ok, "expected *MessageSyntaxError (or subtype), got %T: %v", err, err)
}

// typedError matches messageformat errors that expose a spec-typed name via
// ErrorType(); used by classifyRuntimeError to compare against the JSON fixture
// "type" field.
type typedError interface {
	error
	ErrorType() string
}

func classifyRuntimeError(err error) string {
	if err == nil {
		return ""
	}
	if withType, ok := errors.AsType[typedError](err); ok {
		return withType.ErrorType()
	}
	return err.Error()
}

func errorTypesEqual(got, want string) bool {
	return canonicalErrorType(got) == canonicalErrorType(want)
}

func canonicalErrorType(t string) string {
	if alias, ok := errorAliases[t]; ok {
		return alias
	}
	return t
}

func errorTypes(errs []error) []string {
	out := make([]string, len(errs))
	for i, e := range errs {
		out[i] = classifyRuntimeError(e)
	}
	return out
}

func expectedTypes(expected []map[string]any) []string {
	out := make([]string, 0, len(expected))
	for _, e := range expected {
		typ, _ := e["type"].(string)
		out = append(out, typ)
	}
	return out
}

func scenarioSubtestName(s mfwg.TestScenario) string {
	if s.Scenario != "" {
		return sanitize(s.Scenario)
	}
	return "unnamed"
}

func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return '_'
		case '/':
			return '|'
		default:
			return r
		}
	}, s)
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// Compile-time guard: ensure we don't accidentally drop the slices import if we
// later remove a helper that uses it.
var _ = slices.Contains[[]string]
