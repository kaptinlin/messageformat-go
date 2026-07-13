package resolve

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFunctionRefAppliesUniversalOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		dir            string
		wantDir        bidi.Direction
		wantPartLocale string
	}{
		{name: "ltr", dir: "ltr", wantDir: bidi.DirLTR, wantPartLocale: ""},
		{name: "rtl", dir: "rtl", wantDir: bidi.DirRTL, wantPartLocale: "en"},
		{name: "auto", dir: "auto", wantDir: bidi.DirAuto, wantPartLocale: "en"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := newResolveCoverageContext(nil)
			result := ResolveFunctionRef(ctx, datamodel.NewLiteral("hello"), datamodel.NewFunctionRef("identity", datamodel.Options{
				"u:dir": datamodel.NewLiteral(tc.dir),
				"u:id":  datamodel.NewLiteral("part-id"),
			}))

			assert.Equal(t, "string", result.Type())
			assert.Equal(t, `|hello|`, result.Source())
			assert.Equal(t, tc.wantDir, result.Dir())
			assert.Equal(t, "en", result.Locale())
			optioned, ok := result.(messagevalue.OptionedValue)
			require.True(t, ok)
			assert.Nil(t, optioned.Options())
			assert.True(t, result.(interface{ HasBidiIsolate() bool }).HasBidiIsolate())
			assert.Equal(t, "part-id", result.(interface{ ID() string }).ID())

			got, err := result.ToString()
			require.NoError(t, err)
			assert.Equal(t, "hello", got)

			value, err := result.ValueOf()
			require.NoError(t, err)
			assert.Equal(t, "hello", value)

			selector, ok := result.(messagevalue.Selector)
			require.True(t, ok)
			keys, err := selector.SelectKeys([]string{"hello", "other"})
			require.NoError(t, err)
			if diff := cmp.Diff([]string{"hello"}, keys); diff != "" {
				t.Errorf("selected keys mismatch (-want +got):\n%s", diff)
			}

			parts, err := result.ToParts()
			require.NoError(t, err)
			require.Len(t, parts, 1)
			assert.Equal(t, "string", parts[0].Type())
			assert.Equal(t, "hello", parts[0].Value())
			assert.Equal(t, `|hello|`, parts[0].Source())
			assert.Equal(t, "en", parts[0].Locale())
			assert.Equal(t, tc.wantDir, parts[0].Dir())
			assert.Equal(t, "part-id", parts[0].(interface{ ID() string }).ID())
			assert.Equal(t, tc.dir, parts[0].(interface{ PartDir() string }).PartDir())
			assert.Equal(t, tc.wantPartLocale, parts[0].(interface{ PartLocale() string }).PartLocale())
		})
	}
}

func TestResolveFunctionRefResolvesOptionsAndReportsFailures(t *testing.T) {
	t.Parallel()

	t.Run("literal and variable options are passed to function", func(t *testing.T) {
		t.Parallel()

		var got map[string]any
		ctx := newResolveCoverageContext(map[string]any{"width": "wide"})
		ctx.Functions["capture"] = func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
			got = options
			return messagevalue.NewStringValue("ok", functions.GetFirstLocale(ctx.Locales()), ctx.Source())
		}

		result := ResolveFunctionRef(ctx, nil, datamodel.NewFunctionRef("capture", datamodel.Options{
			"literal": datamodel.NewLiteral("value"),
			"width":   datamodel.NewVariableRef("width"),
		}))
		require.Equal(t, "string", result.Type())
		if diff := cmp.Diff(map[string]any{"literal": "value", "width": "wide"}, got); diff != "" {
			t.Errorf("resolved options mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("function context and resolved options use the same function ref options", func(t *testing.T) {
		t.Parallel()

		var gotLiteralKeys map[string]bool
		var gotOptions map[string]any
		var gotID string
		var gotDir string
		ctx := newResolveCoverageContext(map[string]any{"width": "wide"})
		ctx.Functions["capture"] = func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
			gotLiteralKeys = ctx.LiteralOptionKeys()
			gotOptions = options
			gotID = ctx.ID()
			gotDir = ctx.Dir()
			return messagevalue.NewStringValue("ok", functions.GetFirstLocale(ctx.Locales()), ctx.Source())
		}

		result := ResolveFunctionRef(ctx, nil, datamodel.NewFunctionRef("capture", datamodel.Options{
			"mode":  datamodel.NewLiteral("compact"),
			"width": datamodel.NewVariableRef("width"),
			"u:id":  datamodel.NewLiteral("part-id"),
			"u:dir": datamodel.NewLiteral("rtl"),
		}))
		require.Equal(t, "string", result.Type())
		assert.Equal(t, "part-id", gotID)
		assert.Equal(t, "rtl", gotDir)
		if diff := cmp.Diff(map[string]any{"mode": "compact", "width": "wide"}, gotOptions); diff != "" {
			t.Errorf("resolved options mismatch (-want +got):\n%s", diff)
		}
		assert.True(t, gotLiteralKeys["mode"])
		assert.True(t, gotLiteralKeys["u:id"])
		assert.True(t, gotLiteralKeys["u:dir"])
		assert.False(t, gotLiteralKeys["width"])
	})

	t.Run("unresolved option variable reports unresolved variable and passes nil", func(t *testing.T) {
		t.Parallel()

		var errs []error
		var got map[string]any
		ctx := newResolveCoverageContext(nil)
		ctx.OnError = func(err error) { errs = append(errs, err) }
		ctx.Functions["capture"] = func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
			got = options
			return messagevalue.NewStringValue("ok", functions.GetFirstLocale(ctx.Locales()), ctx.Source())
		}

		result := ResolveFunctionRef(ctx, nil, datamodel.NewFunctionRef("capture", datamodel.Options{
			"missing": datamodel.NewVariableRef("missing"),
		}))
		require.Equal(t, "string", result.Type())
		assert.Nil(t, got["missing"])
		require.Len(t, errs, 1)
		assertResolveCoverageResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeUnresolvedVariable)
	})

	t.Run("invalid universal direction reports bad option", func(t *testing.T) {
		t.Parallel()

		var errs []error
		ctx := newResolveCoverageContext(nil)
		ctx.OnError = func(err error) { errs = append(errs, err) }

		result := ResolveFunctionRef(ctx, datamodel.NewLiteral("hello"), datamodel.NewFunctionRef("identity", datamodel.Options{
			"u:dir": datamodel.NewLiteral("sideways"),
		}))
		require.Equal(t, "string", result.Type())
		require.Len(t, errs, 1)
		assertResolveCoverageResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOption)
	})
}

func TestResolveFunctionRefPassesOperandValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		scope   map[string]any
		operand datamodel.Node
		want    any
	}{
		{
			name: "nil operand",
			want: nil,
		},
		{
			name:    "variable operand",
			scope:   map[string]any{"name": "Ada"},
			operand: datamodel.NewVariableRef("name"),
			want:    "Ada",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got any
			ctx := newResolveCoverageContext(tc.scope)
			ctx.Functions["capture"] = func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
				got = operand
				return messagevalue.NewStringValue("ok", functions.GetFirstLocale(ctx.Locales()), ctx.Source())
			}

			result := ResolveFunctionRef(ctx, tc.operand, datamodel.NewFunctionRef("capture", nil))
			require.Equal(t, "string", result.Type())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveFunctionRefUsesFallbackSource(t *testing.T) {
	t.Parallel()

	ctx := newResolveCoverageContext(nil)

	literalResult := ResolveFunctionRef(ctx, datamodel.NewLiteral(`a|b\c`), datamodel.NewFunctionRef("missing", nil))
	assert.Equal(t, "fallback", literalResult.Type())
	assert.Equal(t, `|a\|b\\c|`, literalResult.Source())

	nilResult := ResolveFunctionRef(ctx, nil, datamodel.NewFunctionRef("missing", nil))
	assert.Equal(t, "fallback", nilResult.Type())
	assert.Equal(t, ":missing", nilResult.Source())
}

func TestFormatMarkupResolvesOptionsAndReportsUniversalDir(t *testing.T) {
	t.Parallel()

	var errs []error
	ctx := newResolveCoverageContext(map[string]any{
		"class": messagevalue.NewStringValue("primary", "en", "$class"),
	})
	ctx.OnError = func(err error) { errs = append(errs, err) }
	part := FormatMarkup(ctx, mustMarkup(t, "open", "span", datamodel.Options{
		"class": datamodel.NewVariableRef("class"),
		"role":  datamodel.NewLiteral("button"),
		"u:id":  datamodel.NewLiteral("cta"),
		"u:dir": datamodel.NewLiteral("rtl"),
	}, nil))

	markup, ok := part.(*messagevalue.MarkupPart)
	require.True(t, ok, "got %T", part)
	if diff := cmp.Diff(map[string]any{
		"class": "primary",
		"role":  "button",
	}, markup.Options()); diff != "" {
		t.Errorf("markup options mismatch (-want +got):\n%s", diff)
	}
	assert.Equal(t, "cta", markup.ID())
	require.Len(t, errs, 1)
	assertResolveCoverageResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOption)
}

func TestUnresolvedInputValueShortcutsOriginalInput(t *testing.T) {
	t.Parallel()

	t.Run("matching variable input", func(t *testing.T) {
		t.Parallel()

		expr := datamodel.NewExpression(datamodel.NewVariableRef("input"), nil, nil)
		unresolved := NewUnresolvedExpression(expr, map[string]any{"input": "original"})
		got, ok := unresolvedInputValue(unresolved)
		require.True(t, ok)
		assert.Equal(t, "original", got)
	})

	t.Run("single concrete scoped value", func(t *testing.T) {
		t.Parallel()

		expr := datamodel.NewExpression(datamodel.NewVariableRef("missing"), nil, nil)
		unresolved := NewUnresolvedExpression(expr, map[string]any{
			"candidate": "only value",
			"pending":   NewUnresolvedExpression(datamodel.NewExpression(datamodel.NewLiteral("x"), nil, nil), nil),
		})
		got, ok := unresolvedInputValue(unresolved)
		require.True(t, ok)
		assert.Equal(t, "only value", got)
	})
}

func TestContextCloneSharesResolvingVarsOnly(t *testing.T) {
	t.Parallel()

	ctx := NewContext([]string{"en"}, nil, map[string]any{"name": "Ada"}, nil)
	ctx.ResolvingVars["name"] = true
	cloned := ctx.Clone()

	cloned.Scope["name"] = "Grace"
	assert.Equal(t, "Ada", ctx.Scope["name"])

	cloned.ResolvingVars["other"] = true
	assert.True(t, ctx.ResolvingVars["other"])
}

func newResolveCoverageContext(scope map[string]any) *Context {
	ctx := NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{},
		scope,
		nil,
	)
	ctx.Functions["identity"] = func(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
		return messagevalue.NewStringValue(operand.(string), functions.GetFirstLocale(ctx.Locales()), ctx.Source())
	}
	return ctx
}

func assertResolveCoverageResolutionErrorType(t *testing.T, err error, want string) {
	t.Helper()

	var resolutionErr *pkgerrors.MessageResolutionError
	require.ErrorAs(t, err, &resolutionErr)
	assert.Equal(t, want, resolutionErr.Type)
}
