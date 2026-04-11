# Formatting Functions

Reference for the built-in formatting functions available in MessageFormat Go v2.

These functions are used inside expressions such as:

```text
{$value :number}
{$date :datetime dateStyle=full}
{$price :number style=currency currency=USD}
```

## Function Model

A formatting expression has three parts:

- operand: `$value` or a literal such as `|hello|`
- annotation: `:number`, `:string`, `:datetime`
- options: `style=currency`, `currency=USD`

General form:

```text
{$variable :function option=value}
```

## Common Built-in Functions

### `:string`

Formats a value as text and is commonly used in selection.

Example:

```go
mf := messageformat.MustNew("en", "Message: {$text :string}")
out, err := mf.Format(map[string]any{"text": "Hello"})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

Selection example:

```text
.input {$status :string}
.match $status
online  {{User is online}}
offline {{User is offline}}
*       {{Status unknown}}
```

### `:number`

Formats numeric values and is also used for plural/category selection.

Example:

```go
mf := messageformat.MustNew("en", "Count: {$count :number}")
out, err := mf.Format(map[string]any{"count": 1234.56})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

Plural-style matching:

```text
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
```

### `:integer`

Formats numeric values as integers.

```text
{$count :integer}
```

### `:datetime`, `:date`, `:time`

Formats temporal values.

Examples:

```text
{$createdAt :datetime dateStyle=full timeStyle=short}
{$createdAt :date style=long}
{$createdAt :time style=short}
```

### `:currency`

Formats money values.

```text
{$amount :currency currency=USD}
```

### `:percent`

Formats percentages.

```text
{$progress :percent}
```

### `:offset`

Formats offset-like values used in message formatting scenarios.

```text
{$count :offset}
```

### `:unit`

Formats values with unit semantics.

```text
{$distance :unit unit=kilometer}
```

## Number Options

Common options for numeric formatting:

| Option | Meaning |
|--------|---------|
| `style` | `decimal`, `currency`, or `percent` depending on formatter behavior |
| `currency` | ISO currency code such as `USD` or `EUR` |
| `currencyDisplay` | Presentation mode for currency |
| `useGrouping` | Whether grouping separators should be used |
| `minimumIntegerDigits` | Minimum integer digits |
| `minimumFractionDigits` | Minimum fraction digits |
| `maximumFractionDigits` | Maximum fraction digits |
| `minimumSignificantDigits` | Minimum significant digits |
| `maximumSignificantDigits` | Maximum significant digits |

Example:

```text
{$amount :number style=currency currency=USD minimumFractionDigits=2}
```

## Date and Time Options

Common options for temporal formatting:

| Option | Meaning |
|--------|---------|
| `style` | high-level style for `:date` or `:time` |
| `dateStyle` | date formatting preset |
| `timeStyle` | time formatting preset |

Example:

```text
{$createdAt :datetime dateStyle=full timeStyle=short}
```

## Locale Behavior

Formatting depends on the active locale list passed to `messageformat.New(...)`.

```go
mf := messageformat.MustNew(
	[]string{"de-DE", "en"},
	"Price: {$amount :number style=currency currency=EUR}",
)
```

The package defaults to `messageformat.LocaleBestFit`. Use `WithLocaleMatcher(...)` if you want to override that behavior.

## Selection Notes

For `.match` selectors:

- declare selector inputs with `.input`
- use `.match $selector`, not `.match {$selector}`
- `:number` and `:string` are common selector annotations

Example:

```text
.input {$count :number}
.match $count
one {{One item}}
*   {{{$count} items}}
```

## Error Handling

Formatting-time issues can be surfaced through `WithErrorHandler(...)`:

```go
out, err := mf.Format(
	map[string]any{"amount": "not-a-number"},
	messageformat.WithErrorHandler(func(err error) {
		log.Printf("format warning: %v", err)
	}),
)
```

Recoverable issues degrade through fallback behavior rather than crashing the formatter.

## Custom Functions

When the built-ins are not enough, add your own with `WithFunction(...)` or `WithFunctions(...)`.

See [Custom Functions](custom-functions.md) for the full custom function contract.
