# Error Handling

Error handling guide for MessageFormat Go implementation based on [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html).

## 📖 Table of Contents

1. [Constructor Functions](#constructor-functions)
2. [Error Types](#error-types)
3. [Syntax Errors](#syntax-errors)
4. [Runtime Errors](#runtime-errors)
5. [Error Callbacks](#error-callbacks)
6. [Error Type Checking](#error-type-checking)
7. [Best Practices](#best-practices)

## Constructor Functions

MessageFormat Go provides two constructor functions:

- `New()` - Returns `(*MessageFormat, error)` for explicit error handling
- `MustNew()` - Returns `*MessageFormat` and panics on error (convenience function)

### Using New() for Explicit Error Handling

```go
mf, err := messageformat.New("en", "Hello {$name}!", nil)
if err != nil {
    log.Fatalf("Failed to create MessageFormat: %v", err)
}
```

### Using MustNew() for Convenience

```go
// Use when you're certain the input is valid
mf := messageformat.MustNew("en", "Hello {$name}!", nil)

// Or in initialization code where panics are acceptable
var templates = map[string]*messageformat.MessageFormat{
    "welcome": messageformat.MustNew("en", "Welcome, {$name}!", nil),
    "goodbye": messageformat.MustNew("en", "Goodbye, {$name}!", nil),
}
```

**⚠️ Warning**: `MustNew()` will panic if there's an error. Only use it when you're certain the input is valid or when panics are acceptable (e.g., during initialization).

## Error Types

MessageFormat Go provides structured error types that implement the standard Go `error` interface and support `errors.Is()` and `errors.As()`.

### Error Hierarchy

```go
// Base error interface
type error interface {
    Error() string
}

// MessageFormat errors also support:
type MessageError interface {
    GetType() string
    Is(target error) bool
}
```

### Error Categories

| Error Type | When | Behavior | Example |
|------------|------|----------|---------|
| **Syntax Errors** | During `New()` | Fails immediately | Invalid syntax, malformed messages |
| **Resolution Errors** | During `Format()` | Returns fallback | Missing variables, unknown functions |
| **Selection Errors** | During `Format()` | Returns fallback | Pattern matching failures |

## Syntax Errors

Syntax errors occur when creating MessageFormat instances due to invalid syntax.

```go
// Missing closing brace
_, err := messageformat.New("en", "Hello {$name")
if err != nil {
    fmt.Printf("Syntax error: %v\n", err)
    // Output: Syntax error: parse-error at 12
}

// Invalid function syntax
_, err = messageformat.New("en", "Count: {$num :}")
if err != nil {
    fmt.Printf("Function error: %v\n", err)
    // Output: Function error: parse-error at 14
}
```

### Handling Syntax Errors

```go
func createMessage(locale, source string) (*messageformat.MessageFormat, error) {
    mf, err := messageformat.New(locale, source)
    if err != nil {
        return nil, fmt.Errorf("template syntax error [%s]: %w", locale, err)
    }
    return mf, nil
}
```

## Runtime Errors

Runtime errors occur during message formatting but don't stop execution.

### Missing Variables

```go
mf, _ := messageformat.New("en", "Hello {$name}! You have {$count} items.")

// Missing variables - uses fallback representation
result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "count" variable missing
})

fmt.Printf("Result: %s\n", result) // Result: Hello ⁨Alice⁩! You have ⁨{$count}⁩ items.
fmt.Printf("Error: %v\n", err)     // Error: <nil> (no error returned)
```

### Type Mismatches

```go
mf, _ := messageformat.New("en", "Price: {$amount :number style=currency currency=USD}")

// String instead of number - graceful handling
result, err := mf.Format(map[string]interface{}{
    "amount": "not-a-number",
})

fmt.Printf("Result: %s\n", result) // Result: Price: ⁨not-a-number⁩
fmt.Printf("Error: %v\n", err)     // Error: <nil>
```

## Error Callbacks

Error callbacks capture warnings without stopping execution.

```go
var warnings []error
onError := func(err error) {
    warnings = append(warnings, err)
}

mf, _ := messageformat.New("en", "Hello {$name} and {$missing}!")

result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "missing" not provided
}, onError)

fmt.Printf("Result: %s\n", result) // Result: Hello ⁨Alice⁩ and ⁨{$missing}⁩!
fmt.Printf("Error: %v\n", err)     // Error: <nil>
fmt.Printf("Warnings: %d\n", len(warnings)) // Warnings: 1

for i, warning := range warnings {
    fmt.Printf("Warning %d: %v\n", i+1, warning)
    // Warning 1: unresolved-variable: Unresolved variable $missing
}
```

## Error Type Checking

Use `errors.As()` to check for specific error types:

```go
func handleError(err error) {
    var syntaxErr *errors.MessageSyntaxError
    var resolutionErr *errors.MessageResolutionError
    var selectionErr *errors.MessageSelectionError
    
    switch {
    case errors.As(err, &syntaxErr):
        fmt.Printf("Syntax error at position %d-%d: %v", syntaxErr.Start, syntaxErr.End, err)
    case errors.As(err, &resolutionErr):
        fmt.Printf("Resolution error (source: %s): %v", resolutionErr.Source, err)
    case errors.As(err, &selectionErr):
        fmt.Printf("Selection error: %v", err)
        if selectionErr.Cause != nil {
            fmt.Printf("Cause: %v", selectionErr.Cause)
        }
    default:
        fmt.Printf("Unknown error: %v", err)
    }
}
```

## Best Practices

### 1. Choose the Right Constructor

```go
// Use New() when error handling is important
func createUserMessage(locale, template string) (*messageformat.MessageFormat, error) {
    mf, err := messageformat.New(locale, template, nil)
    if err != nil {
        return nil, fmt.Errorf("invalid message template: %w", err)
    }
    return mf, nil
}

// Use MustNew() for static templates during initialization
var staticTemplates = map[string]*messageformat.MessageFormat{
    "welcome": messageformat.MustNew("en", "Welcome, {$name}!", nil),
    "error":   messageformat.MustNew("en", "Error: {$message}", nil),
}

// Use MustNew() in tests where panics are acceptable
func TestMessageFormatting(t *testing.T) {
    mf := messageformat.MustNew("en", "Hello {$name}!", nil)
    result, err := mf.Format(map[string]interface{}{"name": "World"})
    assert.NoError(t, err)
    assert.Equal(t, "Hello ⁨World⁩!", result)
}
```

### 2. Validate Templates at Startup

```go
func initializeTemplates() error {
    templates := map[string]string{
        "welcome": "Welcome, {$name}!",
        "goodbye": "Goodbye, {$name}!",
        "error":   "Error: {$message}",
    }
    
    for name, source := range templates {
        _, err := messageformat.New("en", source)
        if err != nil {
            return fmt.Errorf("invalid template %s: %w", name, err)
        }
    }
    
    return nil
}
```

### 3. Handle Errors Appropriately

```go
func formatMessage(mf *messageformat.MessageFormat, variables map[string]interface{}) string {
    result, err := mf.Format(variables)
    if err != nil {
        log.Printf("MessageFormat error: %v", err)
        return "[Formatting failed]"
    }
    return result
}
```

### 4. Test Error Conditions

```go
func TestErrorHandling(t *testing.T) {
    // Test syntax errors
    _, err := messageformat.New("en", "Invalid {$")
    assert.Error(t, err)
    
    var syntaxErr *errors.MessageSyntaxError
    assert.True(t, errors.As(err, &syntaxErr))
    
    // Test runtime fallbacks
    mf, _ := messageformat.New("en", "Hello {$name}!")
    result, err := mf.Format(map[string]interface{}{})
    
    assert.NoError(t, err) // Should not error
    assert.Contains(t, result, "{$name}") // Should contain fallback
    
    // Test error callbacks
    var warnings []error
    onError := func(err error) {
        warnings = append(warnings, err)
    }
    
    result, err = mf.Format(map[string]interface{}{}, onError)
    assert.NoError(t, err)
    assert.Len(t, warnings, 1) // Should capture warning
}
```

## Summary

MessageFormat Go implements **fail-fast construction, graceful runtime** error handling. Syntax errors are detected during template creation, while runtime issues (missing variables, type mismatches) use fallback values rather than failing. Use `New()` for production error handling, `MustNew()` for static templates, and error callbacks for monitoring warnings.

