package selector

import (
	"errors"
	"testing"

	"github.com/kaptinlin/messageformat-go/internal/resolve"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectPatternChoosesLiteralBeforeCatchall(t *testing.T) {
	t.Parallel()

	ctx := newSelectorCoverageContext(map[string]any{"tier": "gold"}, nil)
	message := newSelectorCoverageMessage(t, "tier",
		newSelectorCoverageVariant(t, datamodel.NewLiteral("gold"), "literal"),
		newSelectorCoverageVariant(t, datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "literal", selectorCoverageText(t, result))
}

func TestSelectPatternFallsBackToCatchall(t *testing.T) {
	t.Parallel()

	ctx := newSelectorCoverageContext(map[string]any{"tier": "silver"}, nil)
	message := newSelectorCoverageMessage(t, "tier",
		newSelectorCoverageVariant(t, datamodel.NewLiteral("gold"), "literal"),
		newSelectorCoverageVariant(t, datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "catchall", selectorCoverageText(t, result))
}

func TestSelectPatternMatchesNormalizedLiteralKeys(t *testing.T) {
	t.Parallel()

	ctx := newSelectorCoverageContext(map[string]any{"word": "café"}, nil)
	message := newSelectorCoverageMessage(t, "word",
		newSelectorCoverageVariant(t, datamodel.NewLiteral("café"), "normalized"),
		newSelectorCoverageVariant(t, datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "normalized", selectorCoverageText(t, result))
}

func TestSelectPatternReportsNoMatchWhenNoVariantSurvives(t *testing.T) {
	t.Parallel()

	var errs []error
	ctx := newSelectorCoverageContext(map[string]any{"tier": "silver"}, func(err error) {
		errs = append(errs, err)
	})
	message := newSelectorCoverageMessage(t, "tier",
		newSelectorCoverageVariant(t, datamodel.NewLiteral("gold"), "literal"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, 0, result.Len())
	require.Len(t, errs, 1)
	assertSelectorCoverageErrorType(t, errs[0], pkgerrors.ErrorTypeNoMatch)
}

func TestSelectPatternReportsBadSelectorForNonSelectableValues(t *testing.T) {
	t.Parallel()

	var errs []error
	value, err := messagevalue.NewNumberValueWithSelection(
		42,
		"en",
		"$amount",
		bidi.DirAuto,
		map[string]any{"style": "currency", "currency": "USD"},
		false,
	)
	require.NoError(t, err)
	ctx := newSelectorCoverageContext(map[string]any{"amount": value}, func(err error) {
		errs = append(errs, err)
	})
	ctx.LocalVars[value] = true
	message := newSelectorCoverageMessage(t, "amount",
		newSelectorCoverageVariant(t, datamodel.NewCatchallKey("*"), "fallback"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "fallback", selectorCoverageText(t, result))
	require.Len(t, errs, 1)
	assertSelectorCoverageErrorType(t, errs[0], pkgerrors.ErrorTypeBadSelector)
}

func TestSelectPatternRecoversFromSelectorPanic(t *testing.T) {
	t.Parallel()

	var errs []error
	value := panickingSelectorValue{}
	ctx := newSelectorCoverageContext(map[string]any{"value": value}, func(err error) {
		errs = append(errs, err)
	})
	ctx.LocalVars[value] = true
	message := newSelectorCoverageMessage(t, "value",
		newSelectorCoverageVariant(t, datamodel.NewLiteral("literal"), "literal"),
		newSelectorCoverageVariant(t, datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "catchall", selectorCoverageText(t, result))
	require.Len(t, errs, 1)
	assertSelectorCoverageErrorType(t, errs[0], pkgerrors.ErrorTypeBadSelector)
}

func newSelectorCoverageContext(scope map[string]any, onError func(error)) *resolve.Context {
	return resolve.NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{
			"number": functions.NumberFunction,
			"string": functions.StringFunction,
		},
		scope,
		onError, "best fit")
}

func newSelectorCoverageMessage(t *testing.T, selectorName string, variants ...datamodel.Variant) *datamodel.SelectMessage {
	t.Helper()

	return mustSelectMessage(t,
		nil,
		[]datamodel.VariableRef{*datamodel.NewVariableRef(selectorName)},
		variants,
		"")
}

func newSelectorCoverageVariant(t *testing.T, key datamodel.VariantKey, text string) datamodel.Variant {
	t.Helper()

	return *mustVariant(t,
		[]datamodel.VariantKey{key},
		mustPattern(t, []datamodel.PatternElement{datamodel.NewTextElement(text)}))
}

func selectorCoverageText(t *testing.T, pattern datamodel.Pattern) string {
	t.Helper()

	require.Len(t, pattern.Elements(), 1)
	text, ok := pattern.Elements()[0].(*datamodel.TextElement)
	require.True(t, ok, "got %T", pattern.Elements()[0])
	return text.Value()
}

func assertSelectorCoverageErrorType(t *testing.T, err error, want string) {
	t.Helper()

	var selectionErr *pkgerrors.MessageSelectionError
	require.ErrorAs(t, err, &selectionErr)
	assert.Equal(t, want, selectionErr.Type)
}

type panickingSelectorValue struct{}

func (panickingSelectorValue) Type() string                                 { return "panic-selector" }
func (panickingSelectorValue) Source() string                               { return "$value" }
func (panickingSelectorValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (panickingSelectorValue) Locale() string                               { return "en" }
func (panickingSelectorValue) Options() map[string]any                      { return nil }
func (panickingSelectorValue) ToString() (string, error)                    { return "literal", nil }
func (panickingSelectorValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (panickingSelectorValue) ValueOf() (any, error)                        { return "literal", nil }
func (panickingSelectorValue) SelectKeys(keys []string) ([]string, error) {
	if len(keys) == 1 && keys[0] == "test" {
		return nil, nil
	}
	panic(errors.New("selector failed"))
}
