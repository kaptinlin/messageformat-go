package datamodel_test

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustPattern constructs a valid pattern through the public API.
// TypeScript original code:
// const pattern: Pattern = elements;
func mustPattern(t *testing.T, elements []datamodel.PatternElement) datamodel.Pattern {
	t.Helper()

	pattern, err := datamodel.NewPattern(elements)
	require.NoError(t, err)
	return pattern
}

// mustPatternMessage constructs a valid pattern message through the public API.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func mustPatternMessage(t *testing.T, declarations []datamodel.Declaration, pattern datamodel.Pattern, comment string) *datamodel.PatternMessage {
	t.Helper()

	message, err := datamodel.NewPatternMessage(declarations, pattern, comment)
	require.NoError(t, err)
	return message
}

// mustFunctionRef constructs a valid function reference through the public API.
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func mustFunctionRef(t *testing.T, name string, options datamodel.Options) *datamodel.FunctionRef {
	t.Helper()

	functionRef, err := datamodel.NewFunctionRef(name, options)
	require.NoError(t, err)
	return functionRef
}

// TestInputDeclarationDerivesNameFromExpression proves the public constructor owns the input name.
// TypeScript original code:
// const value = { type: 'expression', arg: { type: 'variable', name: 'count' } };
// const input = { type: 'input', name: value.arg.name, value };
func TestInputDeclarationDerivesNameFromExpression(t *testing.T) {
	t.Parallel()

	expr, err := datamodel.NewExpression(datamodel.NewVariableRef("count"), nil, nil)
	require.NoError(t, err)

	input, err := datamodel.NewInputDeclaration(expr)
	require.NoError(t, err)
	require.Equal(t, "count", input.Name())
	require.Same(t, expr, input.Value())
}

// TestExpressionRejectsEmptyState proves callers cannot construct a valueless expression.
// TypeScript original code:
// if (!arg && !functionRef) throw new TypeError('Invalid expression');
func TestExpressionRejectsEmptyState(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilVariable *datamodel.VariableRef
	tests := []struct {
		name string
		arg  datamodel.ExpressionArg
	}{
		{name: "nil"},
		{name: "typed nil literal", arg: nilLiteral},
		{name: "typed nil variable", arg: nilVariable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := datamodel.NewExpression(tt.arg, nil, nil)
			require.Nil(t, expr)
			require.ErrorIs(t, err, datamodel.ErrInvalidExpression)
		})
	}
}

// TestExpressionAcceptsSupportedShapes characterizes the three valid expression forms.
// TypeScript original code:
// type Expression = { arg?: Literal | VariableRef; functionRef?: FunctionRef };
func TestExpressionAcceptsSupportedShapes(t *testing.T) {
	t.Parallel()

	argument := datamodel.NewLiteral("42")
	function := mustFunctionRef(t, "number", nil)
	tests := []struct {
		name        string
		arg         datamodel.ExpressionArg
		functionRef *datamodel.FunctionRef
	}{
		{name: "argument only", arg: argument},
		{name: "function only", functionRef: function},
		{name: "argument and function", arg: argument, functionRef: function},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := datamodel.NewExpression(tt.arg, tt.functionRef, nil)
			require.NoError(t, err)
			require.NotNil(t, expression)
		})
	}
}

// TestInputDeclarationRejectsExpressionsWithoutVariable proves .input cannot own an invalid expression shape.
// TypeScript original code:
// if (value.arg?.type !== 'variable') throw new TypeError('Invalid input declaration');
func TestInputDeclarationRejectsExpressionsWithoutVariable(t *testing.T) {
	t.Parallel()

	literal, err := datamodel.NewExpression(datamodel.NewLiteral("count"), nil, nil)
	require.NoError(t, err)
	functionOnly, err := datamodel.NewExpression(nil, mustFunctionRef(t, "number", nil), nil)
	require.NoError(t, err)
	var typedNil *datamodel.Expression

	tests := []struct {
		name       string
		expression *datamodel.Expression
	}{
		{name: "nil"},
		{name: "typed nil", expression: typedNil},
		{name: "literal argument", expression: literal},
		{name: "missing variable argument", expression: functionOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := datamodel.NewInputDeclaration(tt.expression)
			require.Nil(t, input)
			require.ErrorIs(t, err, datamodel.ErrInvalidInputDeclaration)
		})
	}
}

// TestInputExpressionIsObservedOnceByModelOperations proves every model operation sees the same expression tree.
// TypeScript original code:
// visit(message, { expression, functionRef, options, attributes, value });
func TestInputExpressionIsObservedOnceByModelOperations(t *testing.T) {
	t.Parallel()

	expression, err := datamodel.NewExpression(
		datamodel.NewVariableRef("count"),
		mustFunctionRef(t, "number", datamodel.Options{
			"style": datamodel.NewVariableRef("style"),
		}),
		datamodel.Attributes{
			"label": datamodel.NewLiteral("Count"),
		},
	)
	require.NoError(t, err)
	input, err := datamodel.NewInputDeclaration(expression)
	require.NoError(t, err)
	message := mustPatternMessage(t,
		[]datamodel.Declaration{input},
		mustPattern(t, nil),
		"",
	)

	assert.Equal(t, ".input {$count :number style=$style @label=Count}\n{{}}", datamodel.StringifyMessage(message))

	counts := map[string]int{}
	datamodel.Visit(message, &datamodel.Visitor{
		Expression: func(*datamodel.Expression, datamodel.VisitContext) func() {
			counts["expression"]++
			return nil
		},
		FunctionRef: func(*datamodel.FunctionRef, datamodel.VisitContext, datamodel.ExpressionArg) func() {
			counts["function"]++
			return nil
		},
		Options: func(datamodel.Options, datamodel.VisitContext) func() {
			counts["options"]++
			return nil
		},
		Attributes: func(datamodel.Attributes, datamodel.VisitContext) func() {
			counts["attributes"]++
			return nil
		},
		Value: func(_ datamodel.ExpressionArg, _ datamodel.VisitContext, position datamodel.ValuePosition) {
			counts[string(position)]++
		},
	})
	assert.Equal(t, map[string]int{
		"expression": 1,
		"function":   1,
		"options":    1,
		"attributes": 1,
		"arg":        1,
		"option":     1,
		"attribute":  1,
	}, counts)

	result, err := datamodel.ValidateMessage(message, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"number"}, result.Functions)
	assert.ElementsMatch(t, []string{"count", "style"}, result.Variables)
}

// TestPatternRejectsNilElements proves successful patterns contain only usable union members.
// TypeScript original code:
// const pattern: Pattern = elements;
func TestPatternRejectsNilElements(t *testing.T) {
	t.Parallel()

	var nilText *datamodel.TextElement
	var nilExpression *datamodel.Expression
	var nilMarkup *datamodel.Markup
	tests := []struct {
		name     string
		elements []datamodel.PatternElement
	}{
		{name: "nil", elements: []datamodel.PatternElement{nil}},
		{name: "typed nil text", elements: []datamodel.PatternElement{nilText}},
		{name: "typed nil expression", elements: []datamodel.PatternElement{nilExpression}},
		{name: "typed nil markup", elements: []datamodel.PatternElement{nilMarkup}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := datamodel.NewPattern(tt.elements)
			require.Nil(t, pattern)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestVariantRejectsNilKeys proves successful variants contain only usable key union members.
// TypeScript original code:
// const variant: Variant = { keys, value };
func TestVariantRejectsNilKeys(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilCatchall *datamodel.CatchallKey
	tests := []struct {
		name string
		keys []datamodel.VariantKey
	}{
		{name: "nil", keys: []datamodel.VariantKey{nil}},
		{name: "typed nil literal", keys: []datamodel.VariantKey{nilLiteral}},
		{name: "typed nil catchall", keys: []datamodel.VariantKey{nilCatchall}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant, err := datamodel.NewVariant(tt.keys, mustPattern(t, nil))
			require.Nil(t, variant)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestVariantRejectsNilPatternElements proves raw Pattern values cannot bypass the owning constructor.
// TypeScript original code:
// const variant: Variant = { keys, value };
func TestVariantRejectsNilPatternElements(t *testing.T) {
	t.Parallel()

	var nilText *datamodel.TextElement
	var nilExpression *datamodel.Expression
	var nilMarkup *datamodel.Markup
	tests := []struct {
		name    string
		pattern datamodel.Pattern
	}{
		{name: "nil", pattern: datamodel.Pattern{nil}},
		{name: "typed nil text", pattern: datamodel.Pattern{nilText}},
		{name: "typed nil expression", pattern: datamodel.Pattern{nilExpression}},
		{name: "typed nil markup", pattern: datamodel.Pattern{nilMarkup}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant, err := datamodel.NewVariant(nil, tt.pattern)
			require.Nil(t, variant)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestPatternMessageRejectsNilDeclarations proves successful messages contain only usable declaration union members.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func TestPatternMessageRejectsNilDeclarations(t *testing.T) {
	t.Parallel()

	var nilInput *datamodel.InputDeclaration
	var nilLocal *datamodel.LocalDeclaration
	tests := []struct {
		name         string
		declarations []datamodel.Declaration
	}{
		{name: "nil", declarations: []datamodel.Declaration{nil}},
		{name: "typed nil input", declarations: []datamodel.Declaration{nilInput}},
		{name: "typed nil local", declarations: []datamodel.Declaration{nilLocal}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := datamodel.NewPatternMessage(tt.declarations, mustPattern(t, nil), "")
			require.Nil(t, message)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestPatternMessageRejectsNilPatternElements proves raw Pattern values cannot bypass message construction.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func TestPatternMessageRejectsNilPatternElements(t *testing.T) {
	t.Parallel()

	var nilExpression *datamodel.Expression
	tests := []struct {
		name    string
		pattern datamodel.Pattern
	}{
		{name: "nil", pattern: datamodel.Pattern{nil}},
		{name: "typed nil", pattern: datamodel.Pattern{nilExpression}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := datamodel.NewPatternMessage(nil, tt.pattern, "")
			require.Nil(t, message)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestSelectMessageRejectsNilDeclarations proves select messages share the declaration invariant.
// TypeScript original code:
// const message: SelectMessage = { type: 'select', declarations, selectors, variants };
func TestSelectMessageRejectsNilDeclarations(t *testing.T) {
	t.Parallel()

	var nilInput *datamodel.InputDeclaration
	var nilLocal *datamodel.LocalDeclaration
	tests := []struct {
		name         string
		declarations []datamodel.Declaration
	}{
		{name: "nil", declarations: []datamodel.Declaration{nil}},
		{name: "typed nil input", declarations: []datamodel.Declaration{nilInput}},
		{name: "typed nil local", declarations: []datamodel.Declaration{nilLocal}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := datamodel.NewSelectMessage(tt.declarations, nil, nil, "")
			require.Nil(t, message)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestCompositeConstructorsAcceptEmptySequences characterizes empty sequence semantics across the model.
// TypeScript original code:
// const message = { type: 'select', declarations: [], selectors: [], variants: [{ keys: [], value: [] }] };
func TestCompositeConstructorsAcceptEmptySequences(t *testing.T) {
	t.Parallel()

	pattern, err := datamodel.NewPattern(nil)
	require.NoError(t, err)
	require.Empty(t, pattern)
	variant, err := datamodel.NewVariant(nil, pattern)
	require.NoError(t, err)
	require.Empty(t, variant.Keys())
	patternMessage, err := datamodel.NewPatternMessage(nil, pattern, "")
	require.NoError(t, err)
	selectMessage, err := datamodel.NewSelectMessage(nil, nil, []datamodel.Variant{*variant}, "")
	require.NoError(t, err)

	for _, message := range []datamodel.Message{patternMessage, selectMessage} {
		assert.NotPanics(t, func() {
			datamodel.StringifyMessage(message)
			datamodel.Visit(message, &datamodel.Visitor{})
			_, validateErr := datamodel.ValidateMessage(message, nil)
			require.NoError(t, validateErr)
		})
	}
}

// TestFunctionRefRejectsNilOptions proves successful function references contain only usable option values.
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func TestFunctionRefRejectsNilOptions(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilVariable *datamodel.VariableRef
	tests := []struct {
		name    string
		options datamodel.Options
	}{
		{name: "nil", options: datamodel.Options{"value": nil}},
		{name: "typed nil literal", options: datamodel.Options{"value": nilLiteral}},
		{name: "typed nil variable", options: datamodel.Options{"value": nilVariable}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functionRef, err := datamodel.NewFunctionRef("number", tt.options)
			require.Nil(t, functionRef)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestExpressionRejectsNilAttributes proves successful expressions contain only usable attribute values.
// TypeScript original code:
// const expression: Expression = { type: 'expression', arg, attributes };
func TestExpressionRejectsNilAttributes(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilBoolean *datamodel.BooleanAttribute
	tests := []struct {
		name       string
		attributes datamodel.Attributes
	}{
		{name: "nil", attributes: datamodel.Attributes{"value": nil}},
		{name: "typed nil literal", attributes: datamodel.Attributes{"value": nilLiteral}},
		{name: "typed nil boolean", attributes: datamodel.Attributes{"value": nilBoolean}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := datamodel.NewExpression(datamodel.NewLiteral("value"), nil, tt.attributes)
			require.Nil(t, expression)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestMarkupRejectsNilOptions proves markup shares the function option invariant.
// TypeScript original code:
// const markup: Markup = { type: 'markup', kind, name, options, attributes };
func TestMarkupRejectsNilOptions(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilVariable *datamodel.VariableRef
	tests := []struct {
		name    string
		options datamodel.Options
	}{
		{name: "nil", options: datamodel.Options{"value": nil}},
		{name: "typed nil literal", options: datamodel.Options{"value": nilLiteral}},
		{name: "typed nil variable", options: datamodel.Options{"value": nilVariable}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markup, err := datamodel.NewMarkup(datamodel.MarkupOpen, "strong", tt.options, nil)
			require.Nil(t, markup)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}

// TestMarkupRejectsNilAttributes proves markup shares the expression attribute invariant.
// TypeScript original code:
// const markup: Markup = { type: 'markup', kind, name, options, attributes };
func TestMarkupRejectsNilAttributes(t *testing.T) {
	t.Parallel()

	var nilLiteral *datamodel.Literal
	var nilBoolean *datamodel.BooleanAttribute
	tests := []struct {
		name       string
		attributes datamodel.Attributes
	}{
		{name: "nil", attributes: datamodel.Attributes{"value": nil}},
		{name: "typed nil literal", attributes: datamodel.Attributes{"value": nilLiteral}},
		{name: "typed nil boolean", attributes: datamodel.Attributes{"value": nilBoolean}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markup, err := datamodel.NewMarkup(datamodel.MarkupOpen, "strong", nil, tt.attributes)
			require.Nil(t, markup)
			require.ErrorIs(t, err, datamodel.ErrNilMember)
		})
	}
}
