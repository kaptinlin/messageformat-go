// Package datamodel provides message data model types and operations for MessageFormat 2.0
// TypeScript original code: data-model/types.ts module
package datamodel

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/internal/cst"
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
}

// BooleanAttribute represents a boolean attribute (true value)
// TypeScript original code: true
type BooleanAttribute struct {
	cst cst.Node
}

// NewBooleanAttribute creates a new boolean attribute
func NewBooleanAttribute() *BooleanAttribute {
	return &BooleanAttribute{cst: nil}
}

// NewBooleanAttributeWithCST creates a new boolean attribute with CST reference
func NewBooleanAttributeWithCST(cst cst.Node) *BooleanAttribute {
	return &BooleanAttribute{cst: cst}
}

func (ba *BooleanAttribute) Type() string {
	return "boolean"
}

func (ba *BooleanAttribute) String() string {
	return "true"
}

func (ba *BooleanAttribute) CST() cst.Node {
	return ba.cst
}

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
	// CST returns the associated CST node if available
	CST() cst.Node
}

// Message represents the root of a message data model
// TypeScript original code:
// export type Message = PatternMessage | SelectMessage;
type Message interface {
	Type() string
	Declarations() []Declaration
	Comment() string
	// CST returns the associated CST node if available
	CST() cst.Node
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
	cst          cst.Node // [cstKey]?: CST.SimpleMessage | CST.ComplexMessage
}

// NewPatternMessage creates a new pattern message
func NewPatternMessage(declarations []Declaration, pattern Pattern, comment string) *PatternMessage {
	if declarations == nil {
		declarations = []Declaration{}
	}
	return &PatternMessage{
		declarations: declarations,
		pattern:      pattern,
		comment:      comment,
		cst:          nil,
	}
}

// NewPatternMessageWithCST creates a new pattern message with CST reference
func NewPatternMessageWithCST(declarations []Declaration, pattern Pattern, comment string, cst cst.Node) *PatternMessage {
	if declarations == nil {
		declarations = []Declaration{}
	}
	return &PatternMessage{
		declarations: declarations,
		pattern:      pattern,
		comment:      comment,
		cst:          cst,
	}
}

func (pm *PatternMessage) Type() string {
	return "message"
}

func (pm *PatternMessage) Declarations() []Declaration {
	return pm.declarations
}

func (pm *PatternMessage) Pattern() Pattern {
	return pm.pattern
}

func (pm *PatternMessage) Comment() string {
	return pm.comment
}

func (pm *PatternMessage) CST() cst.Node {
	return pm.cst
}

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
	cst          cst.Node // [cstKey]?: CST.SelectMessage
}

// NewSelectMessage creates a new select message
func NewSelectMessage(declarations []Declaration, selectors []VariableRef, variants []Variant, comment string) *SelectMessage {
	if declarations == nil {
		declarations = []Declaration{}
	}
	if selectors == nil {
		selectors = []VariableRef{}
	}
	if variants == nil {
		variants = []Variant{}
	}
	return &SelectMessage{
		declarations: declarations,
		selectors:    selectors,
		variants:     variants,
		comment:      comment,
		cst:          nil,
	}
}

// NewSelectMessageWithCST creates a new select message with CST reference
func NewSelectMessageWithCST(declarations []Declaration, selectors []VariableRef, variants []Variant, comment string, cst cst.Node) *SelectMessage {
	if declarations == nil {
		declarations = []Declaration{}
	}
	if selectors == nil {
		selectors = []VariableRef{}
	}
	if variants == nil {
		variants = []Variant{}
	}
	return &SelectMessage{
		declarations: declarations,
		selectors:    selectors,
		variants:     variants,
		comment:      comment,
		cst:          cst,
	}
}

func (sm *SelectMessage) Type() string {
	return "select"
}

func (sm *SelectMessage) Declarations() []Declaration {
	return sm.declarations
}

func (sm *SelectMessage) Selectors() []VariableRef {
	return sm.selectors
}

func (sm *SelectMessage) Variants() []Variant {
	return sm.variants
}

func (sm *SelectMessage) Comment() string {
	return sm.comment
}

func (sm *SelectMessage) CST() cst.Node {
	return sm.cst
}

// Declaration represents variable declarations
// A message may declare any number of input and local variables,
// each with a value defined by an Expression.
// The name of each declaration must be unique within the Message.
// TypeScript original code:
// export type Declaration = InputDeclaration | LocalDeclaration;
type Declaration interface {
	Node // Extends Node interface
	Name() string
	// Value returns the expression value - type varies by declaration type
	// InputDeclaration: VariableRefExpression, LocalDeclaration: Expression
	GetValue() any
	// Value provides backward compatibility - returns interface{} to handle both types
	Value() any
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
	name  string
	value *VariableRefExpression // Expression<VariableRef> in TypeScript
	cst   cst.Node               // [cstKey]?: CST.Declaration
}

// VariableRefExpression represents an Expression with VariableRef constraint
// This ensures the expression has a VariableRef arg, matching TypeScript constraint
// TypeScript original code: Expression<VariableRef>
type VariableRefExpression struct {
	arg         *VariableRef // Must be VariableRef (not optional)
	functionRef *FunctionRef // Optional function reference
	attributes  Attributes   // Optional attributes
	cst         cst.Node     // [cstKey]?: CST.Expression
}

// NewVariableRefExpression creates a new variable reference expression
func NewVariableRefExpression(arg *VariableRef, functionRef *FunctionRef, attributes Attributes) *VariableRefExpression {
	return &VariableRefExpression{
		arg:         arg,
		functionRef: functionRef,
		attributes:  attributes,
		cst:         nil,
	}
}

// NewVariableRefExpressionWithCST creates a new variable reference expression with CST reference
func NewVariableRefExpressionWithCST(arg *VariableRef, functionRef *FunctionRef, attributes Attributes, cst cst.Node) *VariableRefExpression {
	return &VariableRefExpression{
		arg:         arg,
		functionRef: functionRef,
		attributes:  attributes,
		cst:         cst,
	}
}

func (vre *VariableRefExpression) Type() string {
	return "expression"
}

func (vre *VariableRefExpression) Arg() *VariableRef {
	return vre.arg
}

func (vre *VariableRefExpression) FunctionRef() *FunctionRef {
	return vre.functionRef
}

func (vre *VariableRefExpression) Attributes() Attributes {
	return vre.attributes
}

func (vre *VariableRefExpression) CST() cst.Node {
	return vre.cst
}

// NewInputDeclaration creates a new input declaration
func NewInputDeclaration(name string, value *VariableRefExpression) *InputDeclaration {
	return &InputDeclaration{
		name:  name,
		value: value,
		cst:   nil,
	}
}

// NewInputDeclarationWithCST creates a new input declaration with CST reference
func NewInputDeclarationWithCST(name string, value *VariableRefExpression, cst cst.Node) *InputDeclaration {
	return &InputDeclaration{
		name:  name,
		value: value,
		cst:   cst,
	}
}

func (id *InputDeclaration) Type() string {
	return "input"
}

func (id *InputDeclaration) Name() string {
	return id.name
}

func (id *InputDeclaration) Value() any {
	return id.value
}

func (id *InputDeclaration) GetValue() any {
	return id.value
}

func (id *InputDeclaration) CST() cst.Node {
	return id.cst
}

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
	cst   cst.Node // [cstKey]?: CST.Declaration
}

// NewLocalDeclaration creates a new local declaration
func NewLocalDeclaration(name string, value *Expression) *LocalDeclaration {
	return &LocalDeclaration{
		name:  name,
		value: value,
		cst:   nil,
	}
}

// NewLocalDeclarationWithCST creates a new local declaration with CST reference
func NewLocalDeclarationWithCST(name string, value *Expression, cst cst.Node) *LocalDeclaration {
	return &LocalDeclaration{
		name:  name,
		value: value,
		cst:   cst,
	}
}

func (ld *LocalDeclaration) Type() string {
	return "local"
}

func (ld *LocalDeclaration) Name() string {
	return ld.name
}

func (ld *LocalDeclaration) Value() any {
	return ld.value
}

func (ld *LocalDeclaration) GetValue() any {
	return ld.value
}

func (ld *LocalDeclaration) CST() cst.Node {
	return ld.cst
}

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
	cst   cst.Node // [cstKey]?: CST.Variant
}

// NewVariant creates a new variant
func NewVariant(keys []VariantKey, value Pattern) *Variant {
	if keys == nil {
		keys = []VariantKey{}
	}
	return &Variant{
		keys:  keys,
		value: value,
		cst:   nil,
	}
}

// NewVariantWithCST creates a new variant with CST reference
func NewVariantWithCST(keys []VariantKey, value Pattern, cst cst.Node) *Variant {
	if keys == nil {
		keys = []VariantKey{}
	}
	return &Variant{
		keys:  keys,
		value: value,
		cst:   cst,
	}
}

func (v *Variant) Keys() []VariantKey {
	return v.keys
}

func (v *Variant) Value() Pattern {
	return v.value
}

func (v *Variant) CST() cst.Node {
	return v.cst
}

// VariantKey represents keys in variants (Literal or CatchallKey)
// TypeScript original code: Array<Literal | CatchallKey>
type VariantKey interface {
	Node // Extends Node interface
	String() string
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
	cst   cst.Node // [cstKey]?: CST.CatchallKey
}

// NewCatchallKey creates a new catchall key
func NewCatchallKey(value string) *CatchallKey {
	return &CatchallKey{
		value: value,
		cst:   nil,
	}
}

// NewCatchallKeyWithCST creates a new catchall key with CST reference
func NewCatchallKeyWithCST(value string, cst cst.Node) *CatchallKey {
	return &CatchallKey{
		value: value,
		cst:   cst,
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

func (ck *CatchallKey) CST() cst.Node {
	return ck.cst
}

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
func NewPattern(elements []PatternElement) Pattern {
	if elements == nil {
		elements = []PatternElement{}
	}
	return Pattern(elements)
}

// Elements returns the pattern elements
// TypeScript original code: Pattern array access
func (p Pattern) Elements() []PatternElement {
	return []PatternElement(p)
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
}

// TextElement represents literal text in patterns
type TextElement struct {
	value string
	cst   cst.Node // CST reference for text elements
}

// NewTextElement creates a new text element
func NewTextElement(value string) *TextElement {
	return &TextElement{
		value: value,
		cst:   nil,
	}
}

// NewTextElementWithCST creates a new text element with CST reference
func NewTextElementWithCST(value string, cst cst.Node) *TextElement {
	return &TextElement{
		value: value,
		cst:   cst,
	}
}

func (te *TextElement) Type() string {
	return "text"
}

func (te *TextElement) Value() string {
	return te.value
}

func (te *TextElement) CST() cst.Node {
	return te.cst
}

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
	arg         any // Literal, VariableRef, or nil
	functionRef *FunctionRef
	attributes  Attributes // Attributes instead of map[string]interface{}
	cst         cst.Node   // [cstKey]?: CST.Expression
}

// NewExpression creates a new expression
func NewExpression(arg any, functionRef *FunctionRef, attributes Attributes) *Expression {
	return &Expression{
		arg:         arg,
		functionRef: functionRef,
		attributes:  attributes,
		cst:         nil,
	}
}

// NewExpressionWithCST creates a new expression with CST reference
func NewExpressionWithCST(arg any, functionRef *FunctionRef, attributes Attributes, cst cst.Node) *Expression {
	return &Expression{
		arg:         arg,
		functionRef: functionRef,
		attributes:  attributes,
		cst:         cst,
	}
}

func (e *Expression) Type() string {
	return "expression"
}

func (e *Expression) Arg() any {
	return e.arg
}

func (e *Expression) FunctionRef() *FunctionRef {
	return e.functionRef
}

func (e *Expression) Attributes() Attributes {
	return e.attributes
}

func (e *Expression) CST() cst.Node {
	return e.cst
}

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
	cst   cst.Node // [cstKey]?: CST.Literal
}

// NewLiteral creates a new literal
func NewLiteral(value string) *Literal {
	return &Literal{
		value: value,
		cst:   nil,
	}
}

// NewLiteralWithCST creates a new literal with CST reference
func NewLiteralWithCST(value string, cst cst.Node) *Literal {
	return &Literal{
		value: value,
		cst:   cst,
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

func (l *Literal) CST() cst.Node {
	return l.cst
}

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
	cst  cst.Node // [cstKey]?: CST.VariableRef
}

// NewVariableRef creates a new variable reference
func NewVariableRef(name string) *VariableRef {
	return &VariableRef{
		name: name,
		cst:  nil,
	}
}

// NewVariableRefWithCST creates a new variable reference with CST reference
func NewVariableRefWithCST(name string, cst cst.Node) *VariableRef {
	return &VariableRef{
		name: name,
		cst:  cst,
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

func (vr *VariableRef) CST() cst.Node {
	return vr.cst
}

// FunctionRef represents a reference to a function
// To resolve a FunctionRef, a MessageFunction is called.
//
// The name identifies one of the DefaultFunctions,
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
	options Options  // Options instead of map[string]interface{}
	cst     cst.Node // [cstKey]?: CST.FunctionRef
}

// NewFunctionRef creates a new function reference
func NewFunctionRef(name string, options Options) *FunctionRef {
	return &FunctionRef{
		name:    name,
		options: options,
		cst:     nil,
	}
}

// NewFunctionRefWithCST creates a new function reference with CST reference
func NewFunctionRefWithCST(name string, options Options, cst cst.Node) *FunctionRef {
	return &FunctionRef{
		name:    name,
		options: options,
		cst:     cst,
	}
}

func (fr *FunctionRef) Type() string {
	return "function"
}

func (fr *FunctionRef) Name() string {
	return fr.name
}

func (fr *FunctionRef) Options() Options {
	return fr.options
}

func (fr *FunctionRef) CST() cst.Node {
	return fr.cst
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
	kind       string // "open", "standalone", "close"
	name       string
	options    Options    // Options instead of map[string]interface{}
	attributes Attributes // Attributes instead of map[string]interface{}
	cst        cst.Node   // [cstKey]?: CST.Expression
}

// NewMarkup creates a new markup element
func NewMarkup(kind, name string, options Options, attributes Attributes) *Markup {
	return &Markup{
		kind:       kind,
		name:       name,
		options:    options,
		attributes: attributes,
		cst:        nil,
	}
}

// NewMarkupWithCST creates a new markup element with CST reference
func NewMarkupWithCST(kind, name string, options Options, attributes Attributes, cst cst.Node) *Markup {
	return &Markup{
		kind:       kind,
		name:       name,
		options:    options,
		attributes: attributes,
		cst:        cst,
	}
}

func (m *Markup) Type() string {
	return "markup"
}

func (m *Markup) Kind() string {
	return m.kind
}

func (m *Markup) Name() string {
	return m.name
}

func (m *Markup) Options() Options {
	return m.options
}

func (m *Markup) Attributes() Attributes {
	return m.attributes
}

func (m *Markup) CST() cst.Node {
	return m.cst
}

// Helper functions for type conversion and compatibility

// ConvertExpressionToVariableRefExpression converts Expression to VariableRefExpression if possible
// This is used for backward compatibility with fromcst.go
func ConvertExpressionToVariableRefExpression(expr *Expression) *VariableRefExpression {
	if expr == nil {
		return nil
	}

	// Check if the arg is a VariableRef
	if varRef, ok := expr.arg.(*VariableRef); ok {
		return NewVariableRefExpressionWithCST(varRef, expr.functionRef, expr.attributes, expr.cst)
	}

	// If not a VariableRef, we can't convert - this should not happen for InputDeclaration
	return nil
}

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
