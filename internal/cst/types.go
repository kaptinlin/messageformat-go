// Package cst provides concrete syntax tree types for MessageFormat 2.0
// TypeScript original code: cst/types.ts module
package cst

import (
	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// Message represents the CST root node
// TypeScript original code:
// export type Message = SimpleMessage | ComplexMessage | SelectMessage;
type Message interface {
	Type() string
	Errors() []*errors.MessageSyntaxError
}

// SimpleMessage represents a simple message without declarations
// TypeScript original code:
//
//	export interface SimpleMessage {
//	  type: 'simple';
//	  declarations?: never;
//	  pattern: Pattern;
//	  errors: MessageSyntaxError[];
//	}
type SimpleMessage struct {
	pattern Pattern
	errors  []*errors.MessageSyntaxError
}

// Type returns the message type
// TypeScript original code: type: 'simple'
func (m *SimpleMessage) Type() string { return "simple" }

// Errors returns syntax errors
// TypeScript original code: errors: MessageSyntaxError[]
func (m *SimpleMessage) Errors() []*errors.MessageSyntaxError { return m.errors }

// Pattern returns the message pattern
// TypeScript original code: pattern: Pattern
func (m *SimpleMessage) Pattern() Pattern { return m.pattern }

// Declarations returns nil for simple messages
// TypeScript original code: declarations?: never
func (m *SimpleMessage) Declarations() []Declaration { return nil }

// ComplexMessage represents a complex message with declarations
// TypeScript original code:
//
//	export interface ComplexMessage {
//	  type: 'complex';
//	  declarations: Declaration[];
//	  pattern: Pattern;
//	  errors: MessageSyntaxError[];
//	}
type ComplexMessage struct {
	declarations []Declaration
	pattern      Pattern
	errors       []*errors.MessageSyntaxError
}

// Type returns the message type
// TypeScript original code: type: 'complex'
func (m *ComplexMessage) Type() string { return "complex" }

// Errors returns syntax errors
// TypeScript original code: errors: MessageSyntaxError[]
func (m *ComplexMessage) Errors() []*errors.MessageSyntaxError { return m.errors }

// Pattern returns the message pattern
// TypeScript original code: pattern: Pattern
func (m *ComplexMessage) Pattern() Pattern { return m.pattern }

// Declarations returns the message declarations
// TypeScript original code: declarations: Declaration[]
func (m *ComplexMessage) Declarations() []Declaration { return m.declarations }

// SelectMessage represents a select message with variants
// TypeScript original code:
//
//	export interface SelectMessage {
//	  type: 'select';
//	  declarations: Declaration[];
//	  match: Syntax<'.match'>;
//	  selectors: VariableRef[];
//	  variants: Variant[];
//	  errors: MessageSyntaxError[];
//	}
type SelectMessage struct {
	declarations []Declaration
	match        Syntax
	selectors    []VariableRef
	variants     []Variant
	errors       []*errors.MessageSyntaxError
}

// Type returns the message type
// TypeScript original code: type: 'select'
func (m *SelectMessage) Type() string { return "select" }

// Errors returns syntax errors
// TypeScript original code: errors: MessageSyntaxError[]
func (m *SelectMessage) Errors() []*errors.MessageSyntaxError { return m.errors }

// Declarations returns the message declarations
// TypeScript original code: declarations: Declaration[]
func (m *SelectMessage) Declarations() []Declaration { return m.declarations }

// Match returns the match syntax
// TypeScript original code: match: Syntax<'.match'>
func (m *SelectMessage) Match() Syntax { return m.match }

// Selectors returns the selector variable references
// TypeScript original code: selectors: VariableRef[]
func (m *SelectMessage) Selectors() []VariableRef { return m.selectors }

// Variants returns the select variants
// TypeScript original code: variants: Variant[]
func (m *SelectMessage) Variants() []Variant { return m.variants }

// Declaration represents a message declaration
// TypeScript original code:
// export type Declaration = InputDeclaration | LocalDeclaration | Junk;
type Declaration interface {
	Type() string
	Start() int
	End() int
}

// InputDeclaration represents an input declaration
// TypeScript original code:
//
//	export interface InputDeclaration {
//	  type: 'input';
//	  start: number;
//	  end: number;
//	  keyword: Syntax<'.input'>;
//	  value: Expression | Junk;
//	}
type InputDeclaration struct {
	start   int
	end     int
	keyword Syntax
	value   Node // Expression | Junk
}

// Type returns the declaration type
// TypeScript original code: type: 'input'
func (d *InputDeclaration) Type() string { return "input" }

// Start returns the start position
// TypeScript original code: start: number
func (d *InputDeclaration) Start() int { return d.start }

// End returns the end position
// TypeScript original code: end: number
func (d *InputDeclaration) End() int { return d.end }

// Keyword returns the keyword syntax
// TypeScript original code: keyword: Syntax<'.input'>
func (d *InputDeclaration) Keyword() Syntax { return d.keyword }

// Value returns the declaration value
// TypeScript original code: value: Expression | Junk
func (d *InputDeclaration) Value() Node { return d.value }

// LocalDeclaration represents a local declaration
// TypeScript original code:
//
//	export interface LocalDeclaration {
//	  type: 'local';
//	  start: number;
//	  end: number;
//	  keyword: Syntax<'.local'>;
//	  target: VariableRef | Junk;
//	  equals?: Syntax<'='>;
//	  value: Expression | Junk;
//	}
type LocalDeclaration struct {
	start   int
	end     int
	keyword Syntax
	target  Node // VariableRef | Junk
	equals  *Syntax
	value   Node // Expression | Junk
}

// Type returns the declaration type
// TypeScript original code: type: 'local'
func (d *LocalDeclaration) Type() string { return "local" }

// Start returns the start position
// TypeScript original code: start: number
func (d *LocalDeclaration) Start() int { return d.start }

// End returns the end position
// TypeScript original code: end: number
func (d *LocalDeclaration) End() int { return d.end }

// Keyword returns the keyword syntax
// TypeScript original code: keyword: Syntax<'.local'>
func (d *LocalDeclaration) Keyword() Syntax { return d.keyword }

// Target returns the declaration target
// TypeScript original code: target: VariableRef | Junk
func (d *LocalDeclaration) Target() Node { return d.target }

// Equals returns the equals syntax
// TypeScript original code: equals?: Syntax<'='>
func (d *LocalDeclaration) Equals() *Syntax { return d.equals }

// Value returns the declaration value
// TypeScript original code: value: Expression | Junk
func (d *LocalDeclaration) Value() Node { return d.value }

// Variant represents a select variant
// TypeScript original code:
//
//	export interface Variant {
//	  start: number;
//	  end: number;
//	  keys: Array<Literal | CatchallKey>;
//	  value: Pattern;
//	}
type Variant struct {
	start int
	end   int
	keys  []Key // Literal | CatchallKey
	value Pattern
}

// Start returns the start position
// TypeScript original code: start: number
func (v *Variant) Start() int { return v.start }

// End returns the end position
// TypeScript original code: end: number
func (v *Variant) End() int { return v.end }

// Keys returns the variant keys
// TypeScript original code: keys: Array<Literal | CatchallKey>
func (v *Variant) Keys() []Key { return v.keys }

// Value returns the variant pattern
// TypeScript original code: value: Pattern
func (v *Variant) Value() Pattern { return v.value }

// Key represents a variant key
// TypeScript original code: Literal | CatchallKey union type
type Key interface {
	Type() string
	Start() int
	End() int
}

// CatchallKey represents a catchall key (*)
// TypeScript original code:
//
//	export interface CatchallKey {
//	  type: '*';
//	  /** position of the `*` */
//	  start: number;
//	  end: number;
//	}
type CatchallKey struct {
	start int
	end   int
}

// Type returns the key type
// TypeScript original code: type: '*'
func (k *CatchallKey) Type() string { return "*" }

// Start returns the start position
// TypeScript original code: start: number
func (k *CatchallKey) Start() int { return k.start }

// End returns the end position
// TypeScript original code: end: number
func (k *CatchallKey) End() int { return k.end }

// Pattern represents a message pattern
// TypeScript original code:
//
//	export interface Pattern {
//	  start: number;
//	  end: number;
//	  body: Array<Text | Expression>;
//	  braces?: [Syntax<'{{'>] | [Syntax<'{{'>, Syntax<'}}'>];
//	}
type Pattern struct {
	start  int
	end    int
	body   []Node   // Text | Expression
	braces []Syntax // optional braces for quoted patterns
}

func (p *Pattern) Start() int       { return p.start }
func (p *Pattern) End() int         { return p.end }
func (p *Pattern) Body() []Node     { return p.body }
func (p *Pattern) Braces() []Syntax { return p.braces }

// Node represents any CST node
type Node interface {
	Type() string
	Start() int
	End() int
}

// Text represents literal text
// TypeScript original code:
//
//	export interface Text {
//	  type: 'text';
//	  start: number;
//	  end: number;
//	  value: string;
//	}
type Text struct {
	start int
	end   int
	value string
}

func (t *Text) Type() string  { return "text" }
func (t *Text) Start() int    { return t.start }
func (t *Text) End() int      { return t.end }
func (t *Text) Value() string { return t.value }

// Expression represents a placeholder expression
// TypeScript original code:
//
//	export interface Expression {
//	  type: 'expression';
//	  start: number;
//	  end: number;
//	  braces: [Syntax<'{'>] | [Syntax<'{'>, Syntax<'}'>];
//	  arg?: Literal | VariableRef;
//	  functionRef?: FunctionRef | Junk;
//	  markup?: Markup;
//	  attributes: Attribute[];
//	}
type Expression struct {
	start       int
	end         int
	braces      []Syntax
	arg         Node // Literal | VariableRef | nil
	functionRef Node // FunctionRef | Junk | nil
	markup      *Markup
	attributes  []Attribute
}

func (e *Expression) Type() string            { return "expression" }
func (e *Expression) Start() int              { return e.start }
func (e *Expression) End() int                { return e.end }
func (e *Expression) Braces() []Syntax        { return e.braces }
func (e *Expression) Arg() Node               { return e.arg }
func (e *Expression) FunctionRef() Node       { return e.functionRef }
func (e *Expression) Markup() *Markup         { return e.markup }
func (e *Expression) Attributes() []Attribute { return e.attributes }

// Junk represents unparseable content
// TypeScript original code:
//
//	export interface Junk {
//	  type: 'junk';
//	  start: number;
//	  end: number;
//	  source: string;
//	  name?: never;
//	}
type Junk struct {
	start  int
	end    int
	source string
}

func (j *Junk) Type() string   { return "junk" }
func (j *Junk) Start() int     { return j.start }
func (j *Junk) End() int       { return j.end }
func (j *Junk) Source() string { return j.source }

// Literal represents a literal value
// TypeScript original code:
//
//	export interface Literal {
//	  type: 'literal';
//	  quoted: boolean;
//	  start: number;
//	  end: number;
//	  open?: Syntax<'|'>;
//	  value: string;
//	  close?: Syntax<'|'>;
//	}
type Literal struct {
	start  int
	end    int
	quoted bool
	open   *Syntax
	value  string
	close  *Syntax
}

func (l *Literal) Type() string   { return "literal" }
func (l *Literal) Start() int     { return l.start }
func (l *Literal) End() int       { return l.end }
func (l *Literal) Quoted() bool   { return l.quoted }
func (l *Literal) Open() *Syntax  { return l.open }
func (l *Literal) Value() string  { return l.value }
func (l *Literal) Close() *Syntax { return l.close }

// VariableRef represents a variable reference
// TypeScript original code:
//
//	export interface VariableRef {
//	  type: 'variable';
//	  start: number;
//	  end: number;
//	  open: Syntax<'$'>;
//	  name: string;
//	}
type VariableRef struct {
	start int
	end   int
	open  Syntax
	name  string
}

func (v *VariableRef) Type() string { return "variable" }
func (v *VariableRef) Start() int   { return v.start }
func (v *VariableRef) End() int     { return v.end }
func (v *VariableRef) Open() Syntax { return v.open }
func (v *VariableRef) Name() string { return v.name }

// FunctionRef represents a function reference
// TypeScript original code:
//
//	export interface FunctionRef {
//	  type: 'function';
//	  start: number;
//	  end: number;
//	  open: Syntax<':'>;
//	  name: Identifier;
//	  options: Option[];
//	}
type FunctionRef struct {
	start   int
	end     int
	open    Syntax
	name    Identifier
	options []Option
}

func (f *FunctionRef) Type() string      { return "function" }
func (f *FunctionRef) Start() int        { return f.start }
func (f *FunctionRef) End() int          { return f.end }
func (f *FunctionRef) Open() Syntax      { return f.open }
func (f *FunctionRef) Name() Identifier  { return f.name }
func (f *FunctionRef) Options() []Option { return f.options }

// Markup represents markup content
// TypeScript original code:
//
//	export interface Markup {
//	  type: 'markup';
//	  start: number;
//	  end: number;
//	  open: Syntax<'#' | '/'>;
//	  name: Identifier;
//	  options: Option[];
//	  close?: Syntax<'/'>;
//	}
type Markup struct {
	start   int
	end     int
	open    Syntax
	name    Identifier
	options []Option
	close   *Syntax
}

func (m *Markup) Type() string      { return "markup" }
func (m *Markup) Start() int        { return m.start }
func (m *Markup) End() int          { return m.end }
func (m *Markup) Open() Syntax      { return m.open }
func (m *Markup) Name() Identifier  { return m.name }
func (m *Markup) Options() []Option { return m.options }
func (m *Markup) Close() *Syntax    { return m.close }

// Option represents a function or markup option
// TypeScript original code:
//
//	export interface Option {
//	  /** position at the start of the name */
//	  start: number;
//	  end: number;
//	  name: Identifier;
//	  equals?: Syntax<'='>;
//	  value: Literal | VariableRef;
//	}
type Option struct {
	start  int
	end    int
	name   Identifier
	equals *Syntax
	value  Node // Literal | VariableRef
}

func (o *Option) Start() int       { return o.start }
func (o *Option) End() int         { return o.end }
func (o *Option) Name() Identifier { return o.name }
func (o *Option) Equals() *Syntax  { return o.equals }
func (o *Option) Value() Node      { return o.value }

// Attribute represents an expression attribute
// TypeScript original code:
//
//	export interface Attribute {
//	  /** position at the start of the name */
//	  start: number;
//	  end: number;
//	  open: Syntax<'@'>;
//	  name: Identifier;
//	  equals?: Syntax<'='>;
//	  value?: Literal;
//	}
type Attribute struct {
	start  int
	end    int
	open   Syntax
	name   Identifier
	equals *Syntax
	value  *Literal
}

func (a *Attribute) Start() int       { return a.start }
func (a *Attribute) End() int         { return a.end }
func (a *Attribute) Open() Syntax     { return a.open }
func (a *Attribute) Name() Identifier { return a.name }
func (a *Attribute) Equals() *Syntax  { return a.equals }
func (a *Attribute) Value() *Literal  { return a.value }

// Identifier represents an identifier (name or namespace:name)
// TypeScript original code:
// export type Identifier =
//
//	| [name: Syntax<string>]
//	| [namespace: Syntax<string>, separator: Syntax<':'>]
//	| [namespace: Syntax<string>, separator: Syntax<':'>, name: Syntax<string>];
type Identifier []Syntax

// String returns the string representation of the identifier
func (i Identifier) String() string {
	var result string
	for _, part := range i {
		result += part.Value()
	}
	return result
}

// Namespace returns the namespace part if present
func (i Identifier) Namespace() *Syntax {
	if len(i) >= 2 && i[1].Value() == ":" {
		return &i[0]
	}
	return nil
}

// Name returns the name part
func (i Identifier) Name() *Syntax {
	if len(i) == 1 {
		return &i[0]
	} else if len(i) == 3 {
		return &i[2]
	}
	return nil
}

// Separator returns the separator if present
func (i Identifier) Separator() *Syntax {
	if len(i) >= 2 && i[1].Value() == ":" {
		return &i[1]
	}
	return nil
}

// Syntax represents a syntax token with position
// TypeScript original code:
//
//	export interface Syntax<T extends string> {
//	  start: number;
//	  end: number;
//	  value: T;
//	}
type Syntax struct {
	start int
	end   int
	value string
}

func (s *Syntax) Start() int    { return s.start }
func (s *Syntax) End() int      { return s.end }
func (s *Syntax) Value() string { return s.value }

// Constructor functions for creating CST nodes

// NewSimpleMessage creates a new simple message
func NewSimpleMessage(pattern Pattern, errors []*errors.MessageSyntaxError) *SimpleMessage {
	return &SimpleMessage{
		pattern: pattern,
		errors:  errors,
	}
}

// NewComplexMessage creates a new complex message
func NewComplexMessage(declarations []Declaration, pattern Pattern, errors []*errors.MessageSyntaxError) *ComplexMessage {
	return &ComplexMessage{
		declarations: declarations,
		pattern:      pattern,
		errors:       errors,
	}
}

// NewSelectMessage creates a new select message
func NewSelectMessage(declarations []Declaration, match Syntax, selectors []VariableRef, variants []Variant, errors []*errors.MessageSyntaxError) *SelectMessage {
	return &SelectMessage{
		declarations: declarations,
		match:        match,
		selectors:    selectors,
		variants:     variants,
		errors:       errors,
	}
}

// NewInputDeclaration creates a new input declaration
func NewInputDeclaration(start, end int, keyword Syntax, value Node) *InputDeclaration {
	return &InputDeclaration{
		start:   start,
		end:     end,
		keyword: keyword,
		value:   value,
	}
}

// NewLocalDeclaration creates a new local declaration
func NewLocalDeclaration(start, end int, keyword Syntax, target Node, equals *Syntax, value Node) *LocalDeclaration {
	return &LocalDeclaration{
		start:   start,
		end:     end,
		keyword: keyword,
		target:  target,
		equals:  equals,
		value:   value,
	}
}

// NewVariant creates a new variant
func NewVariant(start, end int, keys []Key, value Pattern) *Variant {
	return &Variant{
		start: start,
		end:   end,
		keys:  keys,
		value: value,
	}
}

// NewCatchallKey creates a new catchall key
func NewCatchallKey(start, end int) *CatchallKey {
	return &CatchallKey{
		start: start,
		end:   end,
	}
}

// NewPattern creates a new pattern
func NewPattern(start, end int, body []Node, braces []Syntax) *Pattern {
	return &Pattern{
		start:  start,
		end:    end,
		body:   body,
		braces: braces,
	}
}

// NewText creates a new text node
func NewText(start, end int, value string) *Text {
	return &Text{
		start: start,
		end:   end,
		value: value,
	}
}

// NewExpression creates a new expression
func NewExpression(start, end int, braces []Syntax, arg Node, functionRef Node, markup *Markup, attributes []Attribute) *Expression {
	return &Expression{
		start:       start,
		end:         end,
		braces:      braces,
		arg:         arg,
		functionRef: functionRef,
		markup:      markup,
		attributes:  attributes,
	}
}

// NewJunk creates a new junk node
func NewJunk(start, end int, source string) *Junk {
	return &Junk{
		start:  start,
		end:    end,
		source: source,
	}
}

// NewLiteral creates a new literal
func NewLiteral(start, end int, quoted bool, open *Syntax, value string, close *Syntax) *Literal {
	return &Literal{
		start:  start,
		end:    end,
		quoted: quoted,
		open:   open,
		value:  value,
		close:  close,
	}
}

// NewVariableRef creates a new variable reference
func NewVariableRef(start, end int, open Syntax, name string) *VariableRef {
	return &VariableRef{
		start: start,
		end:   end,
		open:  open,
		name:  name,
	}
}

// NewFunctionRef creates a new function reference
func NewFunctionRef(start, end int, open Syntax, name Identifier, options []Option) *FunctionRef {
	return &FunctionRef{
		start:   start,
		end:     end,
		open:    open,
		name:    name,
		options: options,
	}
}

// NewMarkup creates a new markup
func NewMarkup(start, end int, open Syntax, name Identifier, options []Option, close *Syntax) *Markup {
	return &Markup{
		start:   start,
		end:     end,
		open:    open,
		name:    name,
		options: options,
		close:   close,
	}
}

// NewOption creates a new option
func NewOption(start, end int, name Identifier, equals *Syntax, value Node) *Option {
	return &Option{
		start:  start,
		end:    end,
		name:   name,
		equals: equals,
		value:  value,
	}
}

// NewAttribute creates a new attribute
func NewAttribute(start, end int, open Syntax, name Identifier, equals *Syntax, value *Literal) *Attribute {
	return &Attribute{
		start:  start,
		end:    end,
		open:   open,
		name:   name,
		equals: equals,
		value:  value,
	}
}

// NewSyntax creates a new syntax token
func NewSyntax(start, end int, value string) Syntax {
	return Syntax{
		start: start,
		end:   end,
		value: value,
	}
}
