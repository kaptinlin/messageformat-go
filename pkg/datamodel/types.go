// Package datamodel provides message data model types and operations for MessageFormat 2.0
// TypeScript original code: data-model/types.ts module
package datamodel

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var (
	ErrInvalidExpression       = errors.New("invalid expression")
	ErrInvalidInputDeclaration = errors.New("invalid input declaration")
	ErrInvalidMarkupKind       = errors.New("invalid markup kind")
	ErrNilMember               = errors.New("nil data model member")
)

// Options represents the options of FunctionRef and Markup
// TypeScript original code:
// export type Options = Map<string, Literal | VariableRef>;
type Options map[string]OptionValue

// OptionValue represents values that can be used in options
// TypeScript original code: Literal | VariableRef
type OptionValue interface {
	Node
	String() string
	optionValue()
}

// Attributes represents the attributes of Markup
// TypeScript original code:
// export type Attributes = Map<string, true | Literal>;
type Attributes map[string]AttributeValue

// AttributeValue represents values that can be used in attributes
// TypeScript original code: true | Literal
type AttributeValue interface {
	Node
	String() string
	attributeValue()
}

type sourcePosition interface {
	Start() int
	End() int
}

type sourceSpan struct {
	start int
	end   int
}

func unknownSourceSpan() sourceSpan {
	return sourceSpan{start: -1, end: -1}
}

func sourceSpanFromPosition(position sourcePosition) sourceSpan {
	if position == nil {
		return unknownSourceSpan()
	}
	return sourceSpan{start: position.Start(), end: position.End()}
}

func (s sourceSpan) getPosition() (int, int) {
	return s.start, s.end
}

func (s sourceSpan) Start() int {
	return s.start
}

func (s sourceSpan) End() int {
	return s.end
}

func cloneDeclarations(declarations []Declaration) []Declaration {
	if declarations == nil {
		return []Declaration{}
	}
	return slices.Clone(declarations)
}

// validateDeclarations rejects nil members before declarations are stored.
// TypeScript original code:
// const declarations: Declaration[] = values;
func validateDeclarations(declarations []Declaration) error {
	for i, declaration := range declarations {
		switch value := declaration.(type) {
		case nil:
			return fmt.Errorf("%w: declaration %d", ErrNilMember, i)
		case *InputDeclaration:
			if value == nil {
				return fmt.Errorf("%w: declaration %d", ErrNilMember, i)
			}
		case *LocalDeclaration:
			if value == nil {
				return fmt.Errorf("%w: declaration %d", ErrNilMember, i)
			}
		}
	}
	return nil
}

func cloneVariableRefs(refs []VariableRef) []VariableRef {
	if refs == nil {
		return []VariableRef{}
	}
	return slices.Clone(refs)
}

func cloneVariants(variants []Variant) []Variant {
	if variants == nil {
		return []Variant{}
	}
	cloned := make([]Variant, len(variants))
	for i := range variants {
		cloned[i] = cloneVariantValue(variants[i])
	}
	return cloned
}

func cloneVariantKeys(keys []VariantKey) []VariantKey {
	if keys == nil {
		return []VariantKey{}
	}
	return slices.Clone(keys)
}

func clonePattern(pattern Pattern) Pattern {
	if pattern == nil {
		return Pattern{}
	}
	return Pattern(slices.Clone([]PatternElement(pattern)))
}

// validatePatternElements rejects nil members before a pattern is stored.
// TypeScript original code:
// const pattern: Pattern = elements;
func validatePatternElements(elements []PatternElement) error {
	for i, element := range elements {
		switch value := element.(type) {
		case nil:
			return fmt.Errorf("%w: pattern element %d", ErrNilMember, i)
		case *TextElement:
			if value == nil {
				return fmt.Errorf("%w: pattern element %d", ErrNilMember, i)
			}
		case *Expression:
			if value == nil {
				return fmt.Errorf("%w: pattern element %d", ErrNilMember, i)
			}
		case *Markup:
			if value == nil {
				return fmt.Errorf("%w: pattern element %d", ErrNilMember, i)
			}
		}
	}
	return nil
}

func cloneOptions(options Options) Options {
	if options == nil {
		return nil
	}
	return maps.Clone(options)
}

// validateOptions rejects nil members before options are stored.
// TypeScript original code:
// const options: Options = values;
func validateOptions(options Options) error {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		switch value := options[name].(type) {
		case nil:
			return fmt.Errorf("%w: option %q", ErrNilMember, name)
		case *Literal:
			if value == nil {
				return fmt.Errorf("%w: option %q", ErrNilMember, name)
			}
		case *VariableRef:
			if value == nil {
				return fmt.Errorf("%w: option %q", ErrNilMember, name)
			}
		}
	}
	return nil
}

func cloneAttributes(attributes Attributes) Attributes {
	if attributes == nil {
		return nil
	}
	return maps.Clone(attributes)
}

// validateAttributes rejects nil members before attributes are stored.
// TypeScript original code:
// const attributes: Attributes = values;
func validateAttributes(attributes Attributes) error {
	for _, name := range slices.Sorted(maps.Keys(attributes)) {
		switch value := attributes[name].(type) {
		case nil:
			return fmt.Errorf("%w: attribute %q", ErrNilMember, name)
		case *Literal:
			if value == nil {
				return fmt.Errorf("%w: attribute %q", ErrNilMember, name)
			}
		case *BooleanAttribute:
			if value == nil {
				return fmt.Errorf("%w: attribute %q", ErrNilMember, name)
			}
		}
	}
	return nil
}

func cloneVariantValue(variant Variant) Variant {
	return Variant{
		keys:  cloneVariantKeys(variant.keys),
		value: clonePattern(variant.value),
		span:  variant.span,
	}
}

// ExpressionArg represents a literal or variable reference expression argument.
// TypeScript original code: Literal | VariableRef
type ExpressionArg interface {
	Node
	String() string
	expressionArg()
}

// MarkupKind identifies whether markup opens, closes, or stands alone.
// TypeScript original code: 'open' | 'standalone' | 'close'
type MarkupKind string

const (
	MarkupOpen       MarkupKind = "open"
	MarkupStandalone MarkupKind = "standalone"
	MarkupClose      MarkupKind = "close"
)

// VisitContext identifies where a visited node appears.
// TypeScript original code: 'declaration' | 'selector' | 'placeholder'
type VisitContext string

const (
	VisitDeclaration VisitContext = "declaration"
	VisitSelector    VisitContext = "selector"
	VisitPlaceholder VisitContext = "placeholder"
)

// ValuePosition identifies how a visited value is used.
// TypeScript original code: 'arg' | 'option' | 'attribute'
type ValuePosition string

const (
	ValueArg       ValuePosition = "arg"
	ValueOption    ValuePosition = "option"
	ValueAttribute ValuePosition = "attribute"
)

// BooleanAttribute represents a boolean attribute (true value)
// TypeScript original code: true
type BooleanAttribute struct {
	span sourceSpan
}

// NewBooleanAttribute creates a new boolean attribute
func NewBooleanAttribute() *BooleanAttribute {
	return &BooleanAttribute{}
}

// newBooleanAttributeWithCST creates a new boolean attribute with source position from CST.
func newBooleanAttributeWithCST(cst sourcePosition) *BooleanAttribute {
	return &BooleanAttribute{span: sourceSpanFromPosition(cst)}
}

func (ba *BooleanAttribute) Type() string {
	return "boolean"
}

func (ba *BooleanAttribute) String() string {
	return "true"
}

func (ba *BooleanAttribute) GetPosition() (start, end int) {
	return ba.span.getPosition()
}

func (ba *BooleanAttribute) attributeValue() {}

// Node represents a node in a message data model
// TypeScript original code:
// export type Node =
//
//	| Declaration
//	| Variant
//	| CatchallKey
//	| Expression
//	| Literal
//	| VariableRef
//	| FunctionRef
//	| Markup;
type Node interface {
	Type() string
}

// Message represents the root of a message data model
// TypeScript original code:
// export type Message = PatternMessage | SelectMessage;
type Message interface {
	Type() string
	Declarations() []Declaration
	Comment() string
	message()
}

// PatternMessage represents a single message with no variants
// TypeScript original code:
//
//	export interface PatternMessage {
//	  type: 'message';
//	  declarations: Declaration[];
//	  pattern: Pattern;
//	  comment?: string;
//	  /** @private */
//	  [cstKey]?: CST.SimpleMessage | CST.ComplexMessage;
//	}
type PatternMessage struct {
	declarations []Declaration
	pattern      Pattern
	comment      string
	span         sourceSpan // Source position from [cstKey]?: CST.SimpleMessage | CST.ComplexMessage
}

// NewPatternMessage creates a new pattern message
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func NewPatternMessage(declarations []Declaration, pattern Pattern, comment string) (*PatternMessage, error) {
	return newPatternMessageWithCST(declarations, pattern, comment, nil)
}

// newPatternMessageWithCST creates a new pattern message with source position from CST.
func newPatternMessageWithCST(declarations []Declaration, pattern Pattern, comment string, cst sourcePosition) (*PatternMessage, error) {
	if err := validateDeclarations(declarations); err != nil {
		return nil, err
	}
	if err := validatePatternElements(pattern); err != nil {
		return nil, err
	}

	return &PatternMessage{
		declarations: cloneDeclarations(declarations),
		pattern:      clonePattern(pattern),
		comment:      comment,
		span:         sourceSpanFromPosition(cst),
	}, nil
}

func (pm *PatternMessage) Type() string {
	return "message"
}

func (pm *PatternMessage) Declarations() []Declaration {
	return cloneDeclarations(pm.declarations)
}

func (pm *PatternMessage) Pattern() Pattern {
	return clonePattern(pm.pattern)
}

func (pm *PatternMessage) Comment() string {
	return pm.comment
}

func (pm *PatternMessage) GetPosition() (start, end int) {
	return pm.span.getPosition()
}

func (pm *PatternMessage) message() {}

// SelectMessage represents a message with variants for selection
// SelectMessage generalises the plural, selectordinal and select
// argument types of MessageFormat 1.
// Each case is defined by a key of one or more string identifiers,
// and selection between them is made according to
// the values of a corresponding number of selectors.
//
// Pattern Selection picks the best match among the variants.
// The result of the selection is always a single Pattern.
// TypeScript original code:
//
//	export interface SelectMessage {
//	  type: 'select';
//	  declarations: Declaration[];
//	  selectors: VariableRef[];
//	  variants: Variant[];
//	  comment?: string;
//	  /** @private */
//	  [cstKey]?: CST.SelectMessage;
//	}
type SelectMessage struct {
	declarations []Declaration
	selectors    []VariableRef
	variants     []Variant
	comment      string
	span         sourceSpan // Source position from [cstKey]?: CST.SelectMessage
}

// NewSelectMessage creates a new select message
// TypeScript original code:
// const message: SelectMessage = { type: 'select', declarations, selectors, variants };
func NewSelectMessage(declarations []Declaration, selectors []VariableRef, variants []Variant, comment string) (*SelectMessage, error) {
	return newSelectMessageWithCST(declarations, selectors, variants, comment, nil)
}

// newSelectMessageWithCST creates a new select message with source position from CST.
func newSelectMessageWithCST(declarations []Declaration, selectors []VariableRef, variants []Variant, comment string, cst sourcePosition) (*SelectMessage, error) {
	if err := validateDeclarations(declarations); err != nil {
		return nil, err
	}

	return &SelectMessage{
		declarations: cloneDeclarations(declarations),
		selectors:    cloneVariableRefs(selectors),
		variants:     cloneVariants(variants),
		comment:      comment,
		span:         sourceSpanFromPosition(cst),
	}, nil
}

func (sm *SelectMessage) Type() string {
	return "select"
}

func (sm *SelectMessage) Declarations() []Declaration {
	return cloneDeclarations(sm.declarations)
}

func (sm *SelectMessage) Selectors() []VariableRef {
	return cloneVariableRefs(sm.selectors)
}

func (sm *SelectMessage) Variants() []Variant {
	return cloneVariants(sm.variants)
}

func (sm *SelectMessage) Comment() string {
	return sm.comment
}

func (sm *SelectMessage) GetPosition() (start, end int) {
	return sm.span.getPosition()
}

func (sm *SelectMessage) message() {}

// Declaration represents variable declarations
// A message may declare any number of input and local variables,
// each with a value defined by an Expression.
// The name of each declaration must be unique within the Message.
// TypeScript original code:
// export type Declaration = InputDeclaration | LocalDeclaration;
type Declaration interface {
	Node // Extends Node interface
	Name() string
	declaration()
}

// InputDeclaration represents .input declarations
// TypeScript original code:
//
//	export interface InputDeclaration {
//	  type: 'input';
//	  name: string;
//	  value: Expression<VariableRef>;
//	  /** @private */
//	  [cstKey]?: CST.Declaration;
//	}
type InputDeclaration struct {
	value *Expression
	span  sourceSpan // Source position from [cstKey]?: CST.Declaration
}

// NewInputDeclaration creates a new input declaration
// TypeScript original code:
// const name = value.arg.name;
func NewInputDeclaration(value *Expression) (*InputDeclaration, error) {
	return newInputDeclarationWithCST(value, nil)
}

// newInputDeclarationWithCST creates a new input declaration with source position from CST.
func newInputDeclarationWithCST(value *Expression, cst sourcePosition) (*InputDeclaration, error) {
	if value == nil {
		return nil, ErrInvalidInputDeclaration
	}
	arg, ok := value.arg.(*VariableRef)
	if !ok || arg == nil {
		return nil, ErrInvalidInputDeclaration
	}

	return &InputDeclaration{
		value: value,
		span:  sourceSpanFromPosition(cst),
	}, nil
}

func (id *InputDeclaration) Type() string {
	return "input"
}

func (id *InputDeclaration) Name() string {
	return id.value.arg.(*VariableRef).Name()
}

func (id *InputDeclaration) Value() *Expression {
	return id.value
}

func (id *InputDeclaration) GetPosition() (start, end int) {
	return id.span.getPosition()
}

func (id *InputDeclaration) declaration() {}

// LocalDeclaration represents .local declarations
// TypeScript original code:
//
//	export interface LocalDeclaration {
//	  type: 'local';
//	  name: string;
//	  value: Expression;
//	  /** @private */
//	  [cstKey]?: CST.Declaration;
//	}
type LocalDeclaration struct {
	name  string
	value *Expression
	span  sourceSpan // Source position from [cstKey]?: CST.Declaration
}

// NewLocalDeclaration creates a new local declaration
func NewLocalDeclaration(name string, value *Expression) *LocalDeclaration {
	return &LocalDeclaration{
		name:  name,
		value: value,
		span:  unknownSourceSpan(),
	}
}

// newLocalDeclarationWithCST creates a new local declaration with source position from CST.
func newLocalDeclarationWithCST(name string, value *Expression, cst sourcePosition) *LocalDeclaration {
	return &LocalDeclaration{
		name:  name,
		value: value,
		span:  sourceSpanFromPosition(cst),
	}
}

func (ld *LocalDeclaration) Type() string {
	return "local"
}

func (ld *LocalDeclaration) Name() string {
	return ld.name
}

func (ld *LocalDeclaration) Value() *Expression {
	return ld.value
}

func (ld *LocalDeclaration) GetPosition() (start, end int) {
	return ld.span.getPosition()
}

func (ld *LocalDeclaration) declaration() {}

// Variant represents select message variants
// TypeScript original code:
//
//	export interface Variant {
//	  type?: never;
//	  keys: Array<Literal | CatchallKey>;
//	  value: Pattern;
//	  /** @private */
//	  [cstKey]?: CST.Variant;
//	}
type Variant struct {
	keys  []VariantKey
	value Pattern
	span  sourceSpan // Source position from [cstKey]?: CST.Variant
}

// NewVariant creates a new variant
// TypeScript original code:
// const variant: Variant = { keys, value };
func NewVariant(keys []VariantKey, value Pattern) (*Variant, error) {
	return newVariantWithCST(keys, value, nil)
}

// newVariantWithCST creates a new variant with source position from CST.
func newVariantWithCST(keys []VariantKey, value Pattern, cst sourcePosition) (*Variant, error) {
	for i, key := range keys {
		switch value := key.(type) {
		case nil:
			return nil, fmt.Errorf("%w: variant key %d", ErrNilMember, i)
		case *Literal:
			if value == nil {
				return nil, fmt.Errorf("%w: variant key %d", ErrNilMember, i)
			}
		case *CatchallKey:
			if value == nil {
				return nil, fmt.Errorf("%w: variant key %d", ErrNilMember, i)
			}
		}
	}
	if err := validatePatternElements(value); err != nil {
		return nil, err
	}

	return &Variant{
		keys:  cloneVariantKeys(keys),
		value: clonePattern(value),
		span:  sourceSpanFromPosition(cst),
	}, nil
}

func (v *Variant) Keys() []VariantKey {
	return cloneVariantKeys(v.keys)
}

func (v *Variant) Value() Pattern {
	return clonePattern(v.value)
}

func (v Variant) GetPosition() (start, end int) {
	return v.span.getPosition()
}

// VariantKey represents keys in variants (Literal or CatchallKey)
// TypeScript original code: Array<Literal | CatchallKey>
type VariantKey interface {
	Node // Extends Node interface
	String() string
	variantKey()
}

// CatchallKey represents the catch-all key that matches all values
// TypeScript original code:
//
//	export interface CatchallKey {
//	  type: '*';
//	  value?: string;
//	  /** @private */
//	  [cstKey]?: CST.CatchallKey;
//	}
type CatchallKey struct {
	value string
	span  sourceSpan // Source position from [cstKey]?: CST.CatchallKey
}

// NewCatchallKey creates a new catchall key
func NewCatchallKey(value string) *CatchallKey {
	return &CatchallKey{
		value: value,
		span:  unknownSourceSpan(),
	}
}

// newCatchallKeyWithCST creates a new catchall key with source position from CST.
func newCatchallKeyWithCST(value string, cst sourcePosition) *CatchallKey {
	return &CatchallKey{
		value: value,
		span:  sourceSpanFromPosition(cst),
	}
}

func (ck *CatchallKey) Type() string {
	return "*"
}

func (ck *CatchallKey) Value() string {
	return ck.value
}

func (ck *CatchallKey) String() string {
	if ck.value != "" {
		return ck.value
	}
	return "*"
}

func (ck *CatchallKey) GetPosition() (start, end int) {
	return ck.span.getPosition()
}

func (ck *CatchallKey) variantKey() {}

// Pattern represents the body of a message composed of a sequence of parts
// The body of each Message is composed of a sequence of parts,
// some of them fixed (Text),
// others Expression and Markup placeholders
// for values depending on additional data.
// TypeScript original code:
// export type Pattern = Array<string | Expression | Markup>;
type Pattern []PatternElement

// NewPattern creates a new pattern from elements
// TypeScript original code: Pattern array construction
func NewPattern(elements []PatternElement) (Pattern, error) {
	if err := validatePatternElements(elements); err != nil {
		return nil, err
	}
	if elements == nil {
		return Pattern{}, nil
	}
	return Pattern(slices.Clone(elements)), nil
}

// Elements returns the pattern elements
// TypeScript original code: Pattern array access
func (p Pattern) Elements() []PatternElement {
	if p == nil {
		return []PatternElement{}
	}
	return slices.Clone([]PatternElement(p))
}

// Add adds an element to the pattern
// TypeScript original code: Pattern.push() equivalent
func (p *Pattern) Add(element PatternElement) {
	*p = append(*p, element)
}

// Len returns the number of elements in the pattern
// TypeScript original code: Pattern.length
func (p Pattern) Len() int {
	return len(p)
}

// Get returns the element at the specified index
// TypeScript original code: Pattern[index]
func (p Pattern) Get(index int) PatternElement {
	if index < 0 || index >= len(p) {
		return nil
	}
	return p[index]
}

// PatternElement represents elements in a pattern
// TypeScript original code: string | Expression | Markup
type PatternElement interface {
	Node // Extends Node interface
	patternElement()
}

// TextElement represents literal text in patterns
type TextElement struct {
	value string
	span  sourceSpan // Source position for text elements
}

// NewTextElement creates a new text element
func NewTextElement(value string) *TextElement {
	return &TextElement{
		value: value,
		span:  unknownSourceSpan(),
	}
}

// newTextElementWithCST creates a new text element with source position from CST.
func newTextElementWithCST(value string, cst sourcePosition) *TextElement {
	return &TextElement{
		value: value,
		span:  sourceSpanFromPosition(cst),
	}
}

func (te *TextElement) Type() string {
	return "text"
}

func (te *TextElement) Value() string {
	return te.value
}

func (te *TextElement) GetPosition() (start, end int) {
	return te.span.getPosition()
}

func (te *TextElement) patternElement() {}

// Expression represents expressions used in declarations and placeholders
// Expressions are used in declarations and as placeholders.
// Each must include at least an arg or a functionRef, or both.
// TypeScript original code:
// export type Expression<
//
//	A extends Literal | VariableRef | undefined =
//	  | Literal
//	  | VariableRef
//	  | undefined
//
//	> = {
//	  type: 'expression';
//	  attributes?: Attributes;
//	  /** @private */
//	  [cstKey]?: CST.Expression;
//	} & (A extends Literal | VariableRef
//
//	? { arg: A; functionRef?: FunctionRef }
//	: { arg?: never; functionRef: FunctionRef });
type Expression struct {
	arg         ExpressionArg // Literal, VariableRef, or nil
	functionRef *FunctionRef
	attributes  Attributes // Attributes instead of map[string]interface{}
	span        sourceSpan // Source position from [cstKey]?: CST.Expression
}

// NewExpression creates a new expression
// TypeScript original code:
// if (!arg && !functionRef) throw new TypeError('Invalid expression');
func NewExpression(arg ExpressionArg, functionRef *FunctionRef, attributes Attributes) (*Expression, error) {
	return newExpressionWithCST(arg, functionRef, attributes, nil)
}

// newExpressionWithCST creates a new expression with source position from CST.
func newExpressionWithCST(arg ExpressionArg, functionRef *FunctionRef, attributes Attributes, cst sourcePosition) (*Expression, error) {
	switch value := arg.(type) {
	case *Literal:
		if value == nil {
			arg = nil
		}
	case *VariableRef:
		if value == nil {
			arg = nil
		}
	}
	if arg == nil && functionRef == nil {
		return nil, ErrInvalidExpression
	}
	if err := validateAttributes(attributes); err != nil {
		return nil, err
	}

	return &Expression{
		arg:         arg,
		functionRef: functionRef,
		attributes:  cloneAttributes(attributes),
		span:        sourceSpanFromPosition(cst),
	}, nil
}

func (e *Expression) Type() string {
	return "expression"
}

func (e *Expression) Arg() ExpressionArg {
	return e.arg
}

func (e *Expression) FunctionRef() *FunctionRef {
	return e.functionRef
}

func (e *Expression) Attributes() Attributes {
	return cloneAttributes(e.attributes)
}

func (e *Expression) GetPosition() (start, end int) {
	return e.span.getPosition()
}

func (e *Expression) patternElement() {}

// Literal represents an immediately defined literal value
// An immediately defined literal value.
//
// Always contains a string value.
// In FunctionRef arguments and options,
// the expected type of the value may result in the value being
// further parsed as a boolean or a number by the function handler.
// TypeScript original code:
//
//	export interface Literal {
//	  type: 'literal';
//	  value: string;
//	  /** @private */
//	  [cstKey]?: CST.Literal;
//	}
type Literal struct {
	value string
	span  sourceSpan // Source position from [cstKey]?: CST.Literal
}

// NewLiteral creates a new literal
func NewLiteral(value string) *Literal {
	return &Literal{
		value: value,
		span:  unknownSourceSpan(),
	}
}

// newLiteralWithCST creates a new literal with source position from CST.
func newLiteralWithCST(value string, cst sourcePosition) *Literal {
	return &Literal{
		value: value,
		span:  sourceSpanFromPosition(cst),
	}
}

func (l *Literal) Type() string {
	return "literal"
}

func (l *Literal) Value() string {
	return l.value
}

func (l *Literal) String() string {
	return l.value
}

func (l *Literal) GetPosition() (start, end int) {
	return l.span.getPosition()
}

func (l *Literal) expressionArg() {}

func (l *Literal) optionValue() {}

func (l *Literal) attributeValue() {}

func (l *Literal) variantKey() {}

// VariableRef represents a reference to a variable
// The value of a VariableRef is defined by a declaration,
// or by the msgParams argument of a MessageFormat.format or
// MessageFormat.formatToParts call.
// TypeScript original code:
//
//	export interface VariableRef {
//	  type: 'variable';
//	  name: string;
//	  /** @private */
//	  [cstKey]?: CST.VariableRef;
//	}
type VariableRef struct {
	name string
	span sourceSpan // Source position from [cstKey]?: CST.VariableRef
}

// NewVariableRef creates a new variable reference
func NewVariableRef(name string) *VariableRef {
	return &VariableRef{
		name: name,
		span: unknownSourceSpan(),
	}
}

// newVariableRefWithCST creates a new variable reference with source position from CST.
func newVariableRefWithCST(name string, cst sourcePosition) *VariableRef {
	return &VariableRef{
		name: name,
		span: sourceSpanFromPosition(cst),
	}
}

func (vr *VariableRef) Type() string {
	return "variable"
}

func (vr *VariableRef) Name() string {
	return vr.name
}

func (vr *VariableRef) String() string {
	return vr.name
}

func (vr VariableRef) GetPosition() (start, end int) {
	return vr.span.getPosition()
}

func (vr *VariableRef) expressionArg() {}

func (vr *VariableRef) optionValue() {}

// FunctionRef represents a reference to a function
// To resolve a FunctionRef, a MessageFunction is called.
//
// The name identifies one of the default functions,
// or a function included in the MessageFormatOptions.functions.
// TypeScript original code:
//
//	export interface FunctionRef {
//	  type: 'function';
//	  name: string;
//	  options?: Options;
//	  /** @private */
//	  [cstKey]?: CST.FunctionRef;
//	}
type FunctionRef struct {
	name    string
	options Options    // Options instead of map[string]interface{}
	span    sourceSpan // Source position from [cstKey]?: CST.FunctionRef
}

// NewFunctionRef creates a new function reference
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func NewFunctionRef(name string, options Options) (*FunctionRef, error) {
	return newFunctionRefWithCST(name, options, nil)
}

// newFunctionRefWithCST creates a new function reference with source position from CST.
func newFunctionRefWithCST(name string, options Options, cst sourcePosition) (*FunctionRef, error) {
	if err := validateOptions(options); err != nil {
		return nil, err
	}

	return &FunctionRef{
		name:    name,
		options: cloneOptions(options),
		span:    sourceSpanFromPosition(cst),
	}, nil
}

func (fr *FunctionRef) Type() string {
	return "function"
}

func (fr *FunctionRef) Name() string {
	return fr.name
}

func (fr *FunctionRef) Options() Options {
	return cloneOptions(fr.options)
}

func (fr *FunctionRef) GetPosition() (start, end int) {
	return fr.span.getPosition()
}

// Markup represents markup placeholders
// TypeScript original code:
//
//	export interface Markup {
//	  type: 'markup';
//	  kind: 'open' | 'standalone' | 'close';
//	  name: string;
//	  options?: Options;
//	  attributes?: Attributes;
//	  /** @private */
//	  [cstKey]?: CST.Expression;
//	}
type Markup struct {
	kind       MarkupKind
	name       string
	options    Options    // Options instead of map[string]interface{}
	attributes Attributes // Attributes instead of map[string]interface{}
	span       sourceSpan // Source position from [cstKey]?: CST.Expression
}

// NewMarkup creates a new markup element
func NewMarkup(kind MarkupKind, name string, options Options, attributes Attributes) (*Markup, error) {
	return newMarkupWithCST(kind, name, options, attributes, nil)
}

func validMarkupKind(kind MarkupKind) bool {
	switch kind {
	case MarkupOpen, MarkupStandalone, MarkupClose:
		return true
	default:
		return false
	}
}

// newMarkupWithCST creates a new markup element with source position from CST.
func newMarkupWithCST(kind MarkupKind, name string, options Options, attributes Attributes, cst sourcePosition) (*Markup, error) {
	if !validMarkupKind(kind) {
		return nil, ErrInvalidMarkupKind
	}
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	if err := validateAttributes(attributes); err != nil {
		return nil, err
	}
	return &Markup{
		kind:       kind,
		name:       name,
		options:    cloneOptions(options),
		attributes: cloneAttributes(attributes),
		span:       sourceSpanFromPosition(cst),
	}, nil
}

func (m *Markup) Type() string {
	return "markup"
}

func (m *Markup) Kind() MarkupKind {
	return m.kind
}

func (m *Markup) Name() string {
	return m.name
}

func (m *Markup) Options() Options {
	return cloneOptions(m.options)
}

func (m *Markup) Attributes() Attributes {
	return cloneAttributes(m.attributes)
}

func (m *Markup) GetPosition() (start, end int) {
	return m.span.getPosition()
}

func (m *Markup) patternElement() {}

// ConvertMapToOptions converts map[string]interface{} to Options
func ConvertMapToOptions(m map[string]any) Options {
	if m == nil {
		return nil
	}

	options := make(Options)
	for k, v := range m {
		switch val := v.(type) {
		case *Literal:
			options[k] = val
		case *VariableRef:
			options[k] = val
		case string:
			// Convert string to Literal
			options[k] = NewLiteral(val)
		default:
			// Convert other types to string literals
			options[k] = NewLiteral(fmt.Sprintf("%v", val))
		}
	}
	return options
}

// ConvertMapToAttributes converts map[string]interface{} to Attributes
func ConvertMapToAttributes(m map[string]any) Attributes {
	if m == nil {
		return nil
	}

	attributes := make(Attributes)
	for k, v := range m {
		switch val := v.(type) {
		case *Literal:
			attributes[k] = val
		case bool:
			if val {
				attributes[k] = NewBooleanAttribute()
			}
		case string:
			// Convert string to Literal
			attributes[k] = NewLiteral(val)
		default:
			// Convert other types to string literals
			attributes[k] = NewLiteral(fmt.Sprintf("%v", val))
		}
	}
	return attributes
}
