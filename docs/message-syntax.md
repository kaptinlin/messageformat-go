# Message Syntax

Syntax reference for the MessageFormat Go v2 package, based on Unicode MessageFormat 2.0.

This page focuses on the syntax the package accepts and the defaults that affect what formatted output looks like.

## Message Shapes

The package supports two broad message shapes:

- simple messages: plain text with inline expressions
- complex messages: declarations followed by `.match` selection

Simple message:

```text
Hello, {$name}!
```

Complex message:

```text
.input {$count :number}
.match $count
one {{You have one item}}
*   {{You have {$count} items}}
```

## Expressions

Expressions are enclosed in `{...}`.

Common forms:

```text
{$name}
{$count :number}
{$price :number style=currency currency=USD}
{|hello| :string}
```

Expression parts:

- operand: `$name`, `$count`, or a literal such as `|hello|`
- annotation: `:number`, `:string`, `:datetime`
- options: `style=currency`, `currency=USD`

## Variables

Variables always use the `$` prefix:

```text
{$name}
{$user}
{$count}
```

Example:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"name": "Alice"})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

By default, the package returns clean output without bidi isolation markers. If you need isolation markers, opt in with `WithBidiIsolation(messageformat.BidiDefault)`.

## Literals

Use pipe-delimited literals inside expressions:

```text
{|literal text|}
{|42|}
{|true|}
```

Literals are useful when you want a function call without referencing an external variable:

```text
{|hello| :string}
```

## Declarations

### `.input`

Use `.input` to declare and annotate external values before selection:

```text
.input {$count :number}
.input {$status :string}
```

### `.local`

Use `.local` to derive local values:

```text
.input {$price :number}
.local $formatted = {$price :number style=currency currency=USD}
```

Declarations are processed in order, so later declarations can depend on earlier ones.

## Selection With `.match`

Use `.match` with declared selectors:

```text
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
```

Important rule:

- selectors in `.match` are written as `$count`, not `{$count}`

Incorrect:

```text
.match {$count}
```

Correct:

```text
.match $count
```

Example:

```go
source := `
.input {$count :number}
.match $count
0   {{No messages}}
one {{One message}}
*   {{{$count} messages}}
`

mf, err := messageformat.Parse([]string{"en"}, source)
if err != nil {
	log.Fatal(err)
}

for _, count := range []int{0, 1, 5} {
	out, err := mf.Format(map[string]any{"count": count})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}
```

## Multiple Selectors

You can match on more than one selector:

```text
.input {$gender :string}
.input {$count :number}
.match $gender $count
male one   {{He has one item}}
male *     {{He has {$count} items}}
*    one   {{They have one item}}
*    *     {{They have {$count} items}}
```

Each variant key position corresponds to one selector position.

## Built-in Functions

Common built-in annotations:

- `:number`
- `:integer`
- `:string`
- `:datetime`
- `:date`
- `:time`
- `:currency`
- `:percent`
- `:offset`
- `:unit`

Examples:

```text
{$amount :number style=currency currency=USD}
{$count :integer}
{$createdAt :datetime dateStyle=full timeStyle=short}
```

See [Formatting Functions](formatting-functions.md) for details and option behavior.

## Markup

The parser also supports markup expressions:

```text
{#link}
{/link}
{#link /}
```

These are useful when you want structured output through `FormatToParts`.

## Whitespace

Whitespace inside patterns is preserved:

```text
one {{ You have one item }}
*   {{ You have {$count} items }}
```

Whitespace between selectors and variant keys is significant in `.match` blocks and must be present.

## Escaping and Reserved Characters

Characters with syntax meaning include:

- `{` and `}`
- `$`
- `:`
- `|`
- `*`
- `.`

When you need literal values inside expressions, prefer pipe literals:

```text
{|{not syntax}|}
```

## Bidirectional Text

The package supports bidirectional text, but the default is intentionally simple:

- default: `messageformat.BidiNone`
- opt-in isolation: `messageformat.BidiDefault`

Default behavior:

```go
mf, err := messageformat.Parse([]string{"ar"}, "مرحبا {$name}!")
```

Opt-in isolation:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithBidiIsolation(messageformat.BidiDefault),
)
```

Explicit direction:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithDir(messageformat.DirRTL),
)
```

Use typed options instead of old string literals. Prefer:

- `messageformat.BidiNone`
- `messageformat.BidiDefault`
- `messageformat.DirLTR`
- `messageformat.DirRTL`
- `messageformat.DirAuto`

## Error Behavior

Syntax errors preserve specific categories such as:

- `missing-syntax`
- `bad-selector`
- `extra-content`
- `bad-input-expression`

That means malformed `.match` selectors and missing syntax are reported with more precise error types instead of being flattened into a generic parse error.

## Examples

### Currency in a message

```text
Your balance is {$amount :number style=currency currency=EUR}
```

### Status selection

```text
.input {$status :string}
.match $status
online  {{User is online}}
offline {{User is offline}}
*       {{Status unknown}}
```

### Count and gender selection

```text
.input {$gender :string}
.input {$count :number}
.match $gender $count
male one {{He sent one message}}
male *   {{He sent {$count} messages}}
*    one {{They sent one message}}
*    *   {{They sent {$count} messages}}
```
