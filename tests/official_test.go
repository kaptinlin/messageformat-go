package tests

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	messageformat "github.com/kaptinlin/messageformat-go"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
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

// fixtureErrorTypes maps the pinned fixture vocabulary to the local stable kinds.
var fixtureErrorTypes = map[string]string{
	"missing-fallback-variant": "missing-fallback",
	"variant-key-mismatch":     "key-mismatch",
}

func TestOfficialSuite(t *testing.T) {
	scenarios, err := mfwg.TestScenarios(officialTestsDir)
	require.NoError(t, err, "load official fixtures")
	require.NotEmpty(t, scenarios, "expected at least one official fixture")

	for _, scenario := range scenarios {
		t.Run(scenarioSubtestName(scenario), func(t *testing.T) {
			for i, tc := range mfwg.TestCases(scenario) {
				name := fmt.Sprintf("%d_%s", i, sanitize(mfwg.TestName(tc)))
				t.Run(name, func(t *testing.T) {
					for _, tag := range tc.Tags {
						require.True(t, supportsOfficialTag(tag), "unsupported fixture capability %q in %s", tag, tc.ID())
					}
					runOfficialTest(t, tc)
				})
			}
		})
	}
}

func runOfficialTest(t *testing.T, tc mfwg.Test) {
	t.Helper()
	switch mfwg.GetTestType(tc) {
	case mfwg.TestTypeSyntaxError:
		_, err := newOfficialMessageFormat(tc)
		assertSyntaxError(t, err)
	case mfwg.TestTypeDataModelError:
		_, err := newOfficialMessageFormat(tc)
		assertDataModelError(t, err, tc)
	default:
		runFormatTest(t, tc)
	}
}

// supportsOfficialTag reports whether the harness implements a fixture capability.
// TypeScript original code:
// // No direct equivalent; the reference runner implements every listed capability.
func supportsOfficialTag(tag string) bool {
	switch tag {
	case ":currency", ":percent", "u:dir", "u:id":
		return true
	default:
		return false
	}
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
	fns := make(map[string]functions.MessageFunction, len(functions.DefaultFunctionMap())+len(functions.DraftFunctionMap())+4)
	maps.Copy(fns, functions.DefaultFunctionMap())
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

	if tc.ExpParts != nil {
		var partErrors []error
		parts, partsErr := mf.FormatToParts(tc.GetParamsMap(), messageformat.WithErrorHandler(func(err error) {
			partErrors = append(partErrors, err)
		}))
		require.NoError(t, partsErr, "FormatToParts should not return a non-recoverable error")
		assert.Equal(t, tc.ExpParts, projectMessageParts(tc.ExpParts, parts), "formatted parts mismatch")
		matchExpectedErrors(t, partErrors, tc)
	}
}

// projectMessageParts projects MessagePart values onto fields declared by the fixture.
// TypeScript original code:
// expect(mf.formatToParts(tc.params, onError)).toMatchObject(tc.expParts);
func projectMessageParts(expected []any, actual []messagevalue.MessagePart) []any {
	projected := make([]any, len(actual))
	for i, part := range actual {
		var shape map[string]any
		if i < len(expected) {
			shape, _ = expected[i].(map[string]any)
		}
		if shape == nil {
			shape = map[string]any{"type": part.Type()}
		}
		projected[i] = projectMessagePart(shape, part)
	}
	return projected
}

// projectMessagePart reads the fixture-visible fields from one MessagePart.
// TypeScript original code:
// expect(actualPart).toMatchObject(expectedPart);
func projectMessagePart(shape map[string]any, part messagevalue.MessagePart) map[string]any {
	projected := make(map[string]any, len(shape))
	for field, expected := range shape {
		switch field {
		case "type":
			projected[field] = part.Type()
		case "value":
			projected[field] = part.Value()
		case "source":
			projected[field] = part.Source()
		case "locale":
			if withLocale, ok := part.(interface{ PartLocale() string }); ok {
				projected[field] = withLocale.PartLocale()
			} else {
				projected[field] = part.Locale()
			}
		case "dir":
			projected[field] = string(part.Dir())
		case "id":
			if withID, ok := part.(interface{ ID() string }); ok {
				projected[field] = withID.ID()
			}
		case "kind":
			if markup, ok := part.(interface{ Kind() string }); ok {
				projected[field] = markup.Kind()
			}
		case "name":
			if markup, ok := part.(interface{ Name() string }); ok {
				projected[field] = markup.Name()
			}
		case "options":
			if optioned, ok := part.(interface{ Options() map[string]any }); ok {
				projected[field] = optioned.Options()
			}
		case "parts":
			expectedParts, _ := expected.([]any)
			if compound, ok := part.(interface {
				Parts() []messagevalue.MessagePart
			}); ok {
				projected[field] = projectMessageParts(expectedParts, compound.Parts())
			}
		}
	}
	return projected
}

// TestProjectMessagePartsPreservesMismatches verifies that projection cannot hide fixture drift.
// TypeScript original code:
// expect(actualParts).toMatchObject(expectedParts);
func TestProjectMessagePartsPreservesMismatches(t *testing.T) {
	t.Parallel()

	text := func(value string) messagevalue.MessagePart {
		return messagevalue.NewTextPart(value, value, "")
	}
	tests := []struct {
		name     string
		expected []any
		actual   []messagevalue.MessagePart
	}{
		{
			name:     "missing part",
			expected: []any{map[string]any{"type": "text", "value": "a"}},
		},
		{
			name:   "extra part",
			actual: []messagevalue.MessagePart{text("a")},
		},
		{
			name:     "different field",
			expected: []any{map[string]any{"type": "text", "value": "a"}},
			actual:   []messagevalue.MessagePart{text("b")},
		},
		{
			name: "different order",
			expected: []any{
				map[string]any{"type": "text", "value": "a"},
				map[string]any{"type": "text", "value": "b"},
			},
			actual: []messagevalue.MessagePart{text("b"), text("a")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEqual(t, tt.expected, projectMessageParts(tt.expected, tt.actual))
		})
	}
}

// matchExpectedErrors enforces the TS spec.test.ts semantics:
//   - expErrors absent or false → no runtime errors allowed.
//   - expErrors === true        → at least one runtime error required.
//   - expErrors is an array     → each expected entry must match the observed
//     error at the same index by type (after alias normalization).
func matchExpectedErrors(t *testing.T, runtime []error, tc mfwg.Test) {
	t.Helper()
	assert.NoError(t, validateExpectedErrors(runtime, tc))
}

// validateExpectedErrors compares observed errors with one normalized fixture case.
// TypeScript original code:
// if (Array.isArray(tc.expErrors)) expect(errors).toMatchObject(tc.expErrors);
func validateExpectedErrors(runtime []error, tc mfwg.Test) error {
	expected := mfwg.ExpectedErrors(tc)
	switch v := tc.ExpErrors.(type) {
	case nil:
		if len(runtime) != 0 {
			return fmt.Errorf("unexpected runtime errors: %v", runtime)
		}
	case bool:
		if v {
			if len(runtime) == 0 {
				return fmt.Errorf("expected at least one runtime error")
			}
		} else if len(runtime) != 0 {
			return fmt.Errorf("unexpected runtime errors: %v", runtime)
		}
	default:
		if len(expected) == 0 {
			return nil
		}
		if len(runtime) != len(expected) {
			return fmt.Errorf("error count mismatch: got %v, want %v", errorTypes(runtime), expectedTypes(expected))
		}
		for i, want := range expected {
			wantType, _ := want["type"].(string)
			gotType := classifyRuntimeError(runtime[i])
			if !errorTypesEqual(gotType, wantType) {
				return fmt.Errorf("error[%d] type mismatch: got %q, want %q", i, gotType, wantType)
			}
		}
	}
	return nil
}

// TestValidateExpectedErrors verifies exact runtime error matching.
// TypeScript original code:
// expect(errors).toMatchObject(tc.expErrors);
func TestValidateExpectedErrors(t *testing.T) {
	t.Parallel()

	badOption := pkgerrors.NewMessageResolutionError(pkgerrors.ErrorTypeBadOption, "bad option", "$a")
	unresolved := pkgerrors.NewMessageResolutionError(pkgerrors.ErrorTypeUnresolvedVariable, "unresolved", "$b")
	keyMismatch := pkgerrors.NewMessageResolutionError(pkgerrors.ErrorTypeKeyMismatch, "key mismatch", "$c")
	tests := []struct {
		name    string
		tc      mfwg.Test
		runtime []error
		wantErr bool
	}{
		{name: "absent accepts none"},
		{name: "absent rejects error", runtime: []error{badOption}, wantErr: true},
		{name: "false rejects error", tc: mfwg.Test{ExpErrors: false}, runtime: []error{badOption}, wantErr: true},
		{name: "true accepts error", tc: mfwg.Test{ExpErrors: true}, runtime: []error{badOption}},
		{name: "true rejects none", tc: mfwg.Test{ExpErrors: true}, wantErr: true},
		{
			name:    "array accepts exact errors",
			tc:      mfwg.Test{ExpErrors: []any{map[string]any{"type": "bad-option"}}},
			runtime: []error{badOption},
		},
		{
			name:    "array rejects additional error",
			tc:      mfwg.Test{ExpErrors: []any{map[string]any{"type": "bad-option"}}},
			runtime: []error{badOption, badOption},
			wantErr: true,
		},
		{
			name:    "array rejects missing error",
			tc:      mfwg.Test{ExpErrors: []any{map[string]any{"type": "bad-option"}}},
			wantErr: true,
		},
		{
			name: "array rejects different order",
			tc: mfwg.Test{ExpErrors: []any{
				map[string]any{"type": "bad-option"},
				map[string]any{"type": "unresolved-variable"},
			}},
			runtime: []error{unresolved, badOption},
			wantErr: true,
		},
		{
			name:    "array rejects different kind",
			tc:      mfwg.Test{ExpErrors: []any{map[string]any{"type": "bad-option"}}},
			runtime: []error{unresolved},
			wantErr: true,
		},
		{
			name:    "fixture alias maps one way",
			tc:      mfwg.Test{ExpErrors: []any{map[string]any{"type": "variant-key-mismatch"}}},
			runtime: []error{keyMismatch},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateExpectedErrors(tt.runtime, tt.tc)
			assert.Equal(t, tt.wantErr, err != nil, "got error %v", err)
		})
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

func assertDataModelError(t *testing.T, err error, tc mfwg.Test) {
	t.Helper()
	require.Error(t, err, "expected data-model error")

	expected := mfwg.ExpectedErrors(tc)
	require.Len(t, expected, 1, "data-model cases must declare one owning error")
	wantType, ok := expected[0]["type"].(string)
	require.True(t, ok, "data-model case must declare an error type")
	gotType := classifyRuntimeError(err)
	require.True(t, errorTypesEqual(gotType, wantType), "error type mismatch: got %q, want %q", gotType, wantType)

	syntaxErr, ok := errors.AsType[*pkgerrors.MessageSyntaxError](err)
	require.True(t, ok, "expected *MessageSyntaxError, got %T: %v", err, err)
	assert.LessOrEqual(t, syntaxErr.Start, syntaxErr.End)

	_, isModelError := errors.AsType[*pkgerrors.MessageDataModelError](err)
	switch wantType {
	case "duplicate-attribute", "duplicate-option-name":
		assert.False(t, isModelError, "conversion error %q must remain a plain syntax error", wantType)
	case "duplicate-declaration", "duplicate-variant", "missing-fallback-variant",
		"missing-selector-annotation", "variant-key-mismatch":
		assert.True(t, isModelError, "validator error %q must use MessageDataModelError", wantType)
	default:
		t.Fatalf("unsupported data-model error expectation %q", wantType)
	}
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
	if local, ok := fixtureErrorTypes[want]; ok {
		want = local
	}
	return got == want
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
