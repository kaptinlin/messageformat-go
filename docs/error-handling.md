# Error Handling

Error handling in MessageFormat Go follows a simple rule:

- fail early at construction time
- degrade gracefully at formatting time

That split is intentional. Invalid templates should be rejected immediately. Runtime issues should stay observable without turning every missing value into a hard failure.

## Construction-Time Errors

`messageformat.Parse(...)` parses and validates source text before returning a `MessageFormat` instance. Use `messageformat.Compile(...)` when you already have a `datamodel.Message`.

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello {$name")
if err != nil {
	log.Fatal(err)
}
_ = mf
```

Use `Parse(...)` when you want explicit error handling for source text.

Construction failures are always returned as `error` values, including package initialization and test setup.

## Main Error Categories

The package exposes structured error types in `pkg/errors`:

- `*errors.MessageSyntaxError`
- `*errors.MessageResolutionError`
- `*errors.MessageSelectionError`

These types implement `error`, and selection/resolution errors support error unwrapping where there is an underlying cause.

## Syntax Errors

Syntax errors happen during construction.

Examples of syntax-related categories include:

- `missing-syntax`
- `bad-selector`
- `extra-content`
- `bad-input-expression`
- `duplicate-option-name`

Example:

```go
_, err := messageformat.Parse([]string{"en"}, ".match {$count}")
if err != nil {
	var syntaxErr *errors.MessageSyntaxError
	if stdErrors.As(err, &syntaxErr) {
		fmt.Println(syntaxErr.ErrorType())
		fmt.Println(syntaxErr.Start, syntaxErr.End)
	}
}
```

Important behavior:

- syntax errors now preserve their specific type
- malformed selectors are not flattened into a generic `parse-error`

## Runtime Resolution Errors

Resolution errors happen while formatting values.

Typical categories include:

- `unresolved-variable`
- `bad-operand`
- `bad-option`
- `bad-function-result`
- `unknown-function`

These do not usually cause `Format(...)` itself to fail. Instead, the formatter degrades to fallback output and can report the warning through an error handler.

Example:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello {$name} and {$missing}!")
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"name": "Alice"})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

The missing value is rendered through fallback behavior instead of aborting the entire format operation.

## Selection Errors

Selection errors occur while choosing variants for `.match`.

Common categories:

- `bad-selector`
- `no-match`

If a selector fails during evaluation, the error can be reported and the formatter can still degrade rather than panic the application.

## Format-Time Error Handlers

Use `messageformat.WithErrorHandler(...)` to observe recoverable runtime problems:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello {$name} and {$missing}!")
if err != nil {
	log.Fatal(err)
}

var warnings []error

out, err := mf.Format(
	map[string]any{"name": "Alice"},
	messageformat.WithErrorHandler(func(err error) {
		warnings = append(warnings, err)
	}),
)
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
fmt.Println(len(warnings))
```

## Error Type Inspection

Use `errors.As(...)` to inspect specific categories:

```go
func handle(err error) {
	var syntaxErr *errors.MessageSyntaxError
	var resolutionErr *errors.MessageResolutionError
	var selectionErr *errors.MessageSelectionError

	switch {
	case stdErrors.As(err, &syntaxErr):
		fmt.Println("syntax:", syntaxErr.ErrorType())
	case stdErrors.As(err, &resolutionErr):
		fmt.Println("resolution:", resolutionErr.ErrorType(), resolutionErr.Source)
	case stdErrors.As(err, &selectionErr):
		fmt.Println("selection:", selectionErr.ErrorType())
	default:
		fmt.Println("unknown:", err)
	}
}
```

## Error Chains

`MessageResolutionError` and `MessageSelectionError` support unwrapping when they wrap an underlying cause.

That means `errors.Is(...)` and `errors.As(...)` can work through the chain:

```go
var resolutionErr *errors.MessageResolutionError
if stdErrors.As(err, &resolutionErr) {
	if cause := resolutionErr.Unwrap(); cause != nil {
		fmt.Println("cause:", cause)
	}
}
```

This is especially useful when a higher-level formatting error wraps an underlying operand or option resolution problem.

## Practical Guidance

Use `Parse(...)` when:

- templates are static and checked during development
- templates come from configuration, user input, or external files
- you need explicit error propagation

Use `WithErrorHandler(...)` when:

- you want telemetry or logs for missing variables and formatting problems
- you want graceful runtime degradation without losing visibility

## Recommended Pattern

```go
func compileTemplate(locale, source string) (*messageformat.MessageFormat, error) {
	mf, err := messageformat.Parse([]string{locale}, source)
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}
	return mf, nil
}

func renderTemplate(mf *messageformat.MessageFormat, values map[string]any) string {
	out, err := mf.Format(
		values,
		messageformat.WithErrorHandler(func(err error) {
			log.Printf("messageformat warning: %v", err)
		}),
	)
	if err != nil {
		return "[format failed]"
	}
	return out
}
```

## Summary

MessageFormat Go is strict about invalid templates and forgiving about runtime data problems.

- construction is fail-fast
- formatting is resilient
- syntax errors preserve specific categories
- runtime errors can be observed through handlers
- wrapped causes can be inspected with `errors.Is` and `errors.As`
