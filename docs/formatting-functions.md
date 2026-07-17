# Formatting Functions

Reference for the built-in formatting functions available in MessageFormat Go v2.

Stable default functions are `:number`, `:integer`, `:string`, `:offset`, `:currency`, and `:percent`. Draft functions `:date`, `:datetime`, `:time`, and `:unit` are available by explicitly passing `messageformat.WithFunctions(functions.DraftFunctionMap())`. The `:math` function is an extension and is only available when supplied explicitly with `WithFunction`.

These functions are used inside expressions such as:

```text
{$value :number}
{$date :datetime dateLength=long timePrecision=second}
{$price :currency currency=USD}
```

## Function Model

A formatting expression has three parts:

- operand: `$value` or a literal such as `|hello|`
- annotation: `:number`, `:currency`, `:string`, `:datetime`
- options: `currency=USD`

General form:

```text
{$variable :function option=value}
```

## Stable Default Functions

### `:string`

Formats a value as text and is commonly used in selection.

Example:

```go
mf, err := messageformat.Parse([]string{"en"}, "Message: {$text :string}")
if err != nil {
	log.Fatal(err)
}
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
mf, err := messageformat.Parse([]string{"en"}, "Count: {$count :number}")
if err != nil {
	log.Fatal(err)
}
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

## Draft Functions

Draft functions require explicit opt-in:

```go
mf, err := messageformat.Parse(
	[]string{"en"},
	"Created {$createdAt :datetime dateLength=long timePrecision=minute}",
	messageformat.WithFunctions(functions.DraftFunctionMap()),
)
```

### `:datetime`, `:date`, `:time`

Formats temporal values.

Examples:

```text
{$createdAt :datetime dateLength=long timePrecision=second}
{$createdAt :date fields=year-month-day length=long}
{$createdAt :time precision=second timeZoneStyle=short}
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
{$amount :currency currency=USD fractionDigits=2}
```

## Date and Time Options

Common options for temporal formatting:

| Option | Meaning |
|--------|---------|
| `calendar` | Unicode calendar identifier |
| `numberingSystem` | Unicode numbering-system identifier |
| `timeZone` | IANA name, UTC offset, or `input` when supplied by the operand |
| `hour12` | 12-hour clock selection for `:datetime` and `:time` |
| `dateFields` / `fields` | visible date fields for `:datetime` / `:date` |
| `dateLength` / `length` | `long`, `medium`, or `short` date width |
| `timePrecision` / `precision` | `hour`, `minute`, or `second` precision |
| `timeZoneStyle` | `long` or `short` time-zone name |

Example:

```text
{$createdAt :datetime dateLength=long timePrecision=second timeZone=UTC}
```

## Locale Behavior

Formatting depends on the active locale list passed to `messageformat.Parse(...)` or `messageformat.Compile(...)`.

```go
mf, err := messageformat.Parse(
	[]string{"de-DE", "en"},
	"Price: {$amount :currency currency=EUR}",
)
if err != nil {
	log.Fatal(err)
}
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

Formatting-time issues return fallback output and a non-nil error:

```go
out, err := mf.Format(map[string]any{"amount": "not-a-number"})
if err != nil {
	log.Printf("format diagnostics: %v", err)
}
fmt.Println(out)
```

Recoverable issues degrade through fallback behavior rather than crashing the formatter.

## Custom Functions

When the built-ins are not enough, add your own with `WithFunction(...)` or `WithFunctions(...)`.

See [Custom Functions](custom-functions.md) for the full custom function contract.
