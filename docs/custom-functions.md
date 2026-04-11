# Custom Functions

Custom functions let you extend MessageFormat Go with application-specific formatting logic.

Use them when the built-in functions are not enough and you want to keep formatting rules inside the message template.

## Function Signature

Custom functions implement `messageformat.MessageFunction`:

```go
type MessageFunction func(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue
```

## What the Function Receives

- `ctx`: access to locales, source text, direction, literal option keys, and error reporting
- `options`: resolved function options from the message
- `operand`: the formatted input value or literal operand

Typical context methods:

- `ctx.Locales()`
- `ctx.Source()`
- `ctx.Dir()`
- `ctx.OnError(err)`

## What the Function Returns

Return a `messagevalue.MessageValue`. The most common choices are:

```go
messagevalue.NewStringValue(text, locale, source)
messagevalue.NewNumberValue(number, locale, source, options)
messagevalue.NewFallbackValue(source, locale)
```

`NewFallbackValue` is the right choice when the function cannot produce a meaningful result and you want the formatter to degrade gracefully.

## Basic Example

```go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/kaptinlin/messageformat-go"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func uppercase(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	locales := ctx.Locales()
	locale := "en"
	if len(locales) > 0 {
		locale = locales[0]
	}

	return messagevalue.NewStringValue(
		strings.ToUpper(fmt.Sprint(operand)),
		locale,
		ctx.Source(),
	)
}

func main() {
	mf, err := messageformat.New(
		"en",
		"Hello, {$name :uppercase}!",
		messageformat.WithFunction("uppercase", uppercase),
	)
	if err != nil {
		log.Fatal(err)
	}

	out, err := mf.Format(map[string]any{"name": "world"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
```

Usage in a message:

```text
{$name :uppercase}
```

## Example With Options

```go
func truncate(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	text := fmt.Sprint(operand)
	limit := 20
	suffix := "..."

	if raw, ok := options["length"]; ok {
		switch v := raw.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				limit = parsed
			}
		}
	}

	if raw, ok := options["suffix"]; ok {
		suffix = fmt.Sprint(raw)
	}

	if len(text) > limit {
		cutoff := max(limit-len(suffix), 0)
		text = text[:cutoff] + suffix
	}

	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	return messagevalue.NewStringValue(text, locale, ctx.Source())
}
```

Usage in a message:

```text
{$title :truncate length=12 suffix="..."}
```

## Registration

Register one function:

```go
mf, err := messageformat.New(
	"en",
	"Hello, {$name :uppercase}!",
	messageformat.WithFunction("uppercase", uppercase),
)
```

Register multiple functions:

```go
mf, err := messageformat.New(
	"en",
	template,
	messageformat.WithFunction("uppercase", uppercase),
	messageformat.WithFunction("truncate", truncate),
)
```

Reuse a map of functions:

```go
funcs := map[string]messageformat.MessageFunction{
	"uppercase": uppercase,
	"truncate":  truncate,
}

mf, err := messageformat.New(
	"en",
	template,
	messageformat.WithFunctions(funcs),
)
```

## Error Handling

Custom functions should report recoverable problems through `ctx.OnError` and return a fallback value when appropriate.

```go
func safeDate(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	t, ok := operand.(time.Time)
	if !ok {
		ctx.OnError(fmt.Errorf("safeDate: expected time.Time, got %T", operand))
		return messagevalue.NewFallbackValue(ctx.Source(), locale)
	}

	return messagevalue.NewStringValue(
		t.Format(time.DateOnly),
		locale,
		ctx.Source(),
	)
}
```

Guidelines:

- call `ctx.OnError` for recoverable function-level failures
- return `NewFallbackValue` when you cannot produce a stable result
- prefer deterministic output over partial formatting surprises
- treat `options` as already-resolved values

## Locale-Aware Functions

Use `ctx.Locales()` when behavior should vary by locale:

```go
func relativeLabel(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	label := "just now"
	if locale == "fr" {
		label = "a l'instant"
	}

	return messagevalue.NewStringValue(label, locale, ctx.Source())
}
```

## Testing

For custom functions, test:

- operand type handling
- option parsing
- locale-specific behavior
- fallback behavior on invalid input

The runnable reference example is [`examples/custom-functions/main.go`](../examples/custom-functions/main.go).
