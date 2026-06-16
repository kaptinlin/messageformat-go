package selector

import (
	"errors"
	"testing"

	"github.com/kaptinlin/messageformat-go/internal/resolve"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
)

var errSelectorSelectionFailed = errors.New("selection failed")

func TestSelectPatternBacktracksToEarlierSelectorChoice(t *testing.T) {
	t.Parallel()

	ctx := newSelectorBacktrackingContext(map[string]any{
		"first":  newMapSelectorValue("primary", "secondary"),
		"second": newMapSelectorValue("only"),
	}, nil)
	message := datamodel.NewSelectMessage(
		nil,
		[]datamodel.VariableRef{
			*datamodel.NewVariableRef("first"),
			*datamodel.NewVariableRef("second"),
		},
		[]datamodel.Variant{
			newSelectorCoverageVariantForKeys("primary misses", datamodel.NewLiteral("primary"), datamodel.NewLiteral("missing")),
			newSelectorCoverageVariantForKeys("secondary matches", datamodel.NewLiteral("secondary"), datamodel.NewLiteral("only")),
			newSelectorCoverageVariantForKeys("catchall", datamodel.NewCatchallKey("*"), datamodel.NewCatchallKey("*")),
		},
		"",
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "secondary matches", selectorCoverageText(t, result))
}

func TestSelectPatternDoesNotProbeSelectorBeforeSelection(t *testing.T) {
	t.Parallel()

	var errs []error
	value := probeRejectingSelectorValue{}
	ctx := newSelectorBacktrackingContext(map[string]any{
		"value": value,
	}, func(err error) {
		errs = append(errs, err)
	})
	message := newSelectorCoverageMessage("value",
		newSelectorCoverageVariant(datamodel.NewLiteral("literal"), "literal"),
		newSelectorCoverageVariant(datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "literal", selectorCoverageText(t, result))
	assert.Empty(t, errs)
}

func TestSelectPatternReportsSelectorErrorsDuringSelection(t *testing.T) {
	t.Parallel()

	var errs []error
	ctx := newSelectorBacktrackingContext(map[string]any{
		"value": erroringSelectorValue{},
	}, func(err error) {
		errs = append(errs, err)
	})
	message := newSelectorCoverageMessage("value",
		newSelectorCoverageVariant(datamodel.NewLiteral("literal"), "literal"),
		newSelectorCoverageVariant(datamodel.NewCatchallKey("*"), "catchall"),
	)

	result := SelectPattern(ctx, message)
	assert.Equal(t, "catchall", selectorCoverageText(t, result))
	if assert.Len(t, errs, 1) {
		assertSelectorCoverageErrorType(t, errs[0], pkgerrors.ErrorTypeBadSelector)
	}
}

func TestSelectPatternPassesCandidateKeysInVariantOrder(t *testing.T) {
	t.Parallel()

	orderedKeys := []string{
		"alpha",
		"bravo",
		"charlie",
		"delta",
		"echo",
		"foxtrot",
		"golf",
		"hotel",
	}
	value := &orderSensitiveSelectorValue{
		want:     orderedKeys,
		selected: "delta",
	}
	ctx := newSelectorBacktrackingContext(map[string]any{"value": value}, nil)
	variants := make([]datamodel.Variant, 0, len(orderedKeys)+1)
	for _, key := range orderedKeys {
		variants = append(variants, newSelectorCoverageVariant(datamodel.NewLiteral(key), key))
	}
	variants = append(variants, newSelectorCoverageVariant(datamodel.NewCatchallKey("*"), "catchall"))
	message := newSelectorCoverageMessage("value", variants...)

	for range 25 {
		result := SelectPattern(ctx, message)
		assert.Equal(t, "delta", selectorCoverageText(t, result))
	}
}

func newSelectorBacktrackingContext(scope map[string]any, onError func(error)) *resolve.Context {
	ctx := newSelectorCoverageContext(scope, onError)
	for _, value := range scope {
		if mv, ok := value.(messagevalue.MessageValue); ok {
			ctx.LocalVars[mv] = true
		}
	}
	return ctx
}

func newSelectorCoverageVariantForKeys(text string, keys ...datamodel.VariantKey) datamodel.Variant {
	return *datamodel.NewVariant(
		keys,
		datamodel.NewPattern([]datamodel.PatternElement{datamodel.NewTextElement(text)}),
	)
}

type mapSelectorValue struct {
	preferred []string
}

func newMapSelectorValue(preferred ...string) *mapSelectorValue {
	return &mapSelectorValue{preferred: preferred}
}

func (*mapSelectorValue) Type() string                                 { return "map-selector" }
func (*mapSelectorValue) Source() string                               { return "$value" }
func (*mapSelectorValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (*mapSelectorValue) Locale() string                               { return "en" }
func (*mapSelectorValue) Options() map[string]any                      { return nil }
func (*mapSelectorValue) ToString() (string, error)                    { return "", nil }
func (*mapSelectorValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (*mapSelectorValue) ValueOf() (any, error)                        { return nil, nil }
func (mv *mapSelectorValue) SelectKeys(keys []string) ([]string, error) {
	for _, preferred := range mv.preferred {
		for _, key := range keys {
			if key == preferred {
				return []string{key}, nil
			}
		}
	}
	return nil, nil
}

type erroringSelectorValue struct{}

func (erroringSelectorValue) Type() string                                 { return "error-selector" }
func (erroringSelectorValue) Source() string                               { return "$value" }
func (erroringSelectorValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (erroringSelectorValue) Locale() string                               { return "en" }
func (erroringSelectorValue) Options() map[string]any                      { return nil }
func (erroringSelectorValue) ToString() (string, error)                    { return "", nil }
func (erroringSelectorValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (erroringSelectorValue) ValueOf() (any, error)                        { return nil, nil }
func (erroringSelectorValue) SelectKeys(keys []string) ([]string, error) {
	if len(keys) == 1 && keys[0] == "test" {
		return []string{"test"}, nil
	}
	return nil, errSelectorSelectionFailed
}

type probeRejectingSelectorValue struct{}

func (probeRejectingSelectorValue) Type() string                                 { return "probe-rejecting-selector" }
func (probeRejectingSelectorValue) Source() string                               { return "$value" }
func (probeRejectingSelectorValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (probeRejectingSelectorValue) Locale() string                               { return "en" }
func (probeRejectingSelectorValue) Options() map[string]any                      { return nil }
func (probeRejectingSelectorValue) ToString() (string, error)                    { return "", nil }
func (probeRejectingSelectorValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (probeRejectingSelectorValue) ValueOf() (any, error)                        { return nil, nil }
func (probeRejectingSelectorValue) SelectKeys(keys []string) ([]string, error) {
	if len(keys) == 1 && keys[0] == "test" {
		return nil, errSelectorSelectionFailed
	}
	return []string{"literal"}, nil
}

type orderSensitiveSelectorValue struct {
	want     []string
	selected string
}

func (*orderSensitiveSelectorValue) Type() string                                 { return "order-sensitive-selector" }
func (*orderSensitiveSelectorValue) Source() string                               { return "$value" }
func (*orderSensitiveSelectorValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (*orderSensitiveSelectorValue) Locale() string                               { return "en" }
func (*orderSensitiveSelectorValue) Options() map[string]any                      { return nil }
func (*orderSensitiveSelectorValue) ToString() (string, error)                    { return "", nil }
func (*orderSensitiveSelectorValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (*orderSensitiveSelectorValue) ValueOf() (any, error)                        { return nil, nil }
func (mv *orderSensitiveSelectorValue) SelectKeys(keys []string) ([]string, error) {
	if len(keys) != len(mv.want) {
		return nil, nil
	}
	for i, key := range keys {
		if key != mv.want[i] {
			return nil, nil
		}
	}
	return []string{mv.selected}, nil
}

var _ messagevalue.MessageValue = (*mapSelectorValue)(nil)
var _ messagevalue.Selector = (*mapSelectorValue)(nil)
var _ messagevalue.MessageValue = erroringSelectorValue{}
var _ messagevalue.Selector = erroringSelectorValue{}
var _ messagevalue.MessageValue = probeRejectingSelectorValue{}
var _ messagevalue.Selector = probeRejectingSelectorValue{}
var _ messagevalue.MessageValue = (*orderSensitiveSelectorValue)(nil)
var _ messagevalue.Selector = (*orderSensitiveSelectorValue)(nil)
