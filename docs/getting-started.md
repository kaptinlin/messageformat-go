# Getting Started

This guide covers the fastest path to using MessageFormat Go v2 in a real application.

## Requirements

- Go 1.26.1 or newer
- `github.com/kaptinlin/messageformat-go`

Install the package:

```bash
go get github.com/kaptinlin/messageformat-go
```

## First Message

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/messageformat-go"
)

func main() {
	mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
	if err != nil {
		log.Fatal(err)
	}

	out, err := mf.Format(map[string]any{"name": "World"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
```

## Core Concepts

MessageFormat Go v2 revolves around three ideas:

- a message is parsed once with `messageformat.Parse(...)`
- variables are passed to `Format(...)` or `FormatToParts(...)`
- selection logic is expressed in the template, not in surrounding application code

## Variable Syntax

Variables use the MessageFormat 2.0 form with a `$` prefix:

```text
Hello, {$name}!
```

Multiple variables:

```go
mf, err := messageformat.Parse(
	[]string{"en"},
	"Hello {$firstName} {$lastName}! You have {$count :number} messages.",
)
```

## Formatting Values

Built-in functions are attached inline:

```text
{$count :number}
{$amount :number style=currency currency=USD}
{$createdAt :datetime dateStyle=full}
```

Example:

```go
mf, err := messageformat.Parse(
	[]string{"en"},
	"Total: {$amount :number style=currency currency=USD}",
)
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"amount": 42.50})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

## Pattern Matching

Use `.input` and `.match` for conditional output:

```go
source := `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`

mf, err := messageformat.Parse([]string{"en"}, source)
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"count": 5})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

Key rule:

- write selectors as `.match $count`, not `.match {$count}`

## Defaults

The package keeps the default path simple:

- `BidiIsolation` defaults to `messageformat.BidiNone`
- `LocaleMatcher` defaults to `messageformat.LocaleBestFit`
- instances are safe for concurrent use after construction

That means basic formatting does not insert bidi isolation markers unless you opt in.

If you need explicit isolation for mixed-direction output:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithBidiIsolation(messageformat.BidiDefault),
)
```

## Error Handling

Construction errors are returned immediately:

```go
mf, err := messageformat.Parse([]string{"en"}, ".match {$count}")
if err != nil {
	log.Fatal(err)
}
```

Runtime issues are reported through fallbacks and optional error handlers:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello {$name}")
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(
	map[string]any{},
	messageformat.WithErrorHandler(func(err error) {
		log.Printf("format warning: %v", err)
	}),
)
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

Use `Parse(...)` for source text and `Compile(...)` for an existing data model. Handle invalid templates through the returned error.

## Next Steps

After this page, the most useful follow-ups are:

1. [Message Syntax](message-syntax.md)
2. [Formatting Functions](formatting-functions.md)
3. [Custom Functions](custom-functions.md)
4. [Error Handling](error-handling.md)
