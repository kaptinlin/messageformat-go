# Custom Functions

MessageFormat Go allows you to create custom formatting functions that extend the built-in functionality. This guide covers the essentials of developing and using custom functions.

## 📖 Table of Contents

1. [Function Interface](#function-interface)
2. [Context and Values](#context-and-values)
3. [Basic Examples](#basic-examples)
4. [Registration](#registration)
5. [Advanced Examples](#advanced-examples)

## 🔧 Function Interface

Custom functions implement the `MessageFunction` type:

```go
type MessageFunction func(
    ctx MessageFunctionContext,
    options map[string]interface{},
    operand interface{},
) messagevalue.MessageValue
```

### Parameters

- **`ctx`**: Provides locale, direction, and error handling context
- **`options`**: Function options from the message template (e.g., `{$value :func opt=val}`)
- **`operand`**: The input value to format
- **Returns**: A `MessageValue` containing the formatted result

## 🌐 Context and Values

### MessageFunctionContext

Access formatting environment through the context:

```go
func MyFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    locale := ctx.Locales()[0]        // Current locale
    direction := ctx.Dir()            // Text direction: "ltr", "rtl", "auto"
    source := ctx.Source()            // Source text for error reporting
    
    // Handle errors
    if err != nil {
        ctx.OnError(err)
        return messagevalue.NewFallbackValue(source, locale)
    }
    
    return messagevalue.NewStringValue(result, locale, source)
}
```

### MessageValue Types

Return appropriate value types:

```go
// String values
messagevalue.NewStringValue(text, locale, source)

// Number values  
messagevalue.NewNumberValue(number, locale, source, options)

// Fallback for errors
messagevalue.NewFallbackValue(source, locale)
```

## 🎨 Basic Examples

### 1. Simple Text Transform

```go
func UppercaseFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    text := fmt.Sprintf("%v", operand)
    result := strings.ToUpper(text)
    
    locale := ctx.Locales()[0]
    return messagevalue.NewStringValue(result, locale, ctx.Source())
}

// Usage: {$name :uppercase}
```

### 2. Function with Options

```go
func TruncateFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    text := fmt.Sprintf("%v", operand)
    
    // Default values
    maxLength := 50
    suffix := "..."
    
    // Parse options
    if val, ok := options["length"]; ok {
        if length, ok := val.(float64); ok {
            maxLength = int(length)
        }
    }
    
    if val, ok := options["suffix"]; ok {
        suffix = fmt.Sprintf("%v", val)
    }
    
    // Apply truncation
    if len(text) > maxLength {
        if maxLength <= len(suffix) {
            text = suffix[:maxLength]
        } else {
            text = text[:maxLength-len(suffix)] + suffix
        }
    }
    
    locale := ctx.Locales()[0]
    return messagevalue.NewStringValue(text, locale, ctx.Source())
}

// Usage: {$text :truncate length=20 suffix=…}
```

### 3. Locale-Aware Function

```go
func RelativeTimeFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    // Parse time from operand
    var t time.Time
    switch v := operand.(type) {
    case time.Time:
        t = v
    case string:
        parsed, err := time.Parse(time.RFC3339, v)
        if err != nil {
            ctx.OnError(err)
            return messagevalue.NewFallbackValue(ctx.Source(), ctx.Locales()[0])
        }
        t = parsed
    default:
        ctx.OnError(fmt.Errorf("expected time value"))
        return messagevalue.NewFallbackValue(ctx.Source(), ctx.Locales()[0])
    }
    
    diff := time.Since(t)
    locale := ctx.Locales()[0]
    
    var result string
    switch {
    case diff < time.Minute:
        result = localizeJustNow(locale)
    case diff < time.Hour:
        minutes := int(diff.Minutes())
        result = localizeMinutesAgo(locale, minutes)
    default:
        hours := int(diff.Hours())
        result = localizeHoursAgo(locale, hours)
    }
    
    return messagevalue.NewStringValue(result, locale, ctx.Source())
}

func localizeJustNow(locale string) string {
    switch locale {
    case "es": return "ahora mismo"
    case "fr": return "à l'instant"
    default: return "just now"
    }
}

// Usage: {$timestamp :relativeTime}
```

## 📝 Registration

### Single Function

```go
mf := messageformat.MustNew("en", "Hello {$name :uppercase}!",
    messageformat.WithFunction("uppercase", UppercaseFunction))
```

### Multiple Functions

```go
mf := messageformat.MustNew("en", template,
    messageformat.WithFunction("uppercase", UppercaseFunction),
    messageformat.WithFunction("truncate", TruncateFunction),
    messageformat.WithFunction("relativeTime", RelativeTimeFunction))
```

### Function Registry

```go
// Create reusable registry
registry := functions.NewFunctionRegistry()
registry.Register("uppercase", UppercaseFunction)
registry.Register("truncate", TruncateFunction)

// Convert registry to function map for use with MessageFormat
funcs := make(map[string]functions.MessageFunction)
for _, name := range registry.List() {
    if fn, ok := registry.Get(name); ok {
        funcs[name] = fn
    }
}

mf := messageformat.MustNew("en", template,
    messageformat.WithFunctions(funcs))
```

## 🎨 Advanced Examples

### 1. Markdown Processor

```go
func MarkdownFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    markdown := fmt.Sprintf("%v", operand)
    
    // Simple markdown to HTML
    html := markdown
    html = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(html, `<strong>$1</strong>`)
    html = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(html, `<em>$1</em>`)
    
    // Check output format
    format := "text" // default
    if val, ok := options["format"]; ok {
        format = fmt.Sprintf("%v", val)
    }
    
    var result string
    if format == "html" {
        result = html
    } else {
        // Strip HTML tags for plain text
        result = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, "")
    }
    
    locale := ctx.Locales()[0]
    return messagevalue.NewStringValue(result, locale, ctx.Source())
}

// Usage: {$content :markdown format=html}
```

### 2. Number Formatter with Custom Rules

```go
func CustomNumberFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    // Convert to number
    var num float64
    switch v := operand.(type) {
    case float64:
        num = v
    case int:
        num = float64(v)
    case string:
        parsed, err := strconv.ParseFloat(v, 64)
        if err != nil {
            ctx.OnError(err)
            return messagevalue.NewFallbackValue(ctx.Source(), ctx.Locales()[0])
        }
        num = parsed
    default:
        ctx.OnError(fmt.Errorf("expected number"))
        return messagevalue.NewFallbackValue(ctx.Source(), ctx.Locales()[0])
    }
    
    // Apply custom formatting rules
    style := "decimal"
    if val, ok := options["style"]; ok {
        style = fmt.Sprintf("%v", val)
    }
    
    var result string
    switch style {
    case "compact":
        result = formatCompactNumber(num, ctx.Locales()[0])
    case "ordinal":
        result = formatOrdinalNumber(int(num), ctx.Locales()[0])
    default:
        result = fmt.Sprintf("%.2f", num)
    }
    
    locale := ctx.Locales()[0]
    return messagevalue.NewStringValue(result, locale, ctx.Source())
}

func formatCompactNumber(num float64, locale string) string {
    switch {
    case num >= 1e9:
        return fmt.Sprintf("%.1fB", num/1e9)
    case num >= 1e6:
        return fmt.Sprintf("%.1fM", num/1e6)
    case num >= 1e3:
        return fmt.Sprintf("%.1fK", num/1e3)
    default:
        return fmt.Sprintf("%.0f", num)
    }
}

// Usage: {$count :customNumber style=compact}
```

### 3. Error Handling Best Practices

```go
func SafeFunction(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
    locale := ctx.Locales()[0]
    source := ctx.Source()
    
    // Validate input
    if operand == nil {
        ctx.OnError(fmt.Errorf("operand is required"))
        return messagevalue.NewFallbackValue(source, locale)
    }
    
    // Try processing
    result, err := processValue(operand, options)
    if err != nil {
        // Log error but don't fail
        ctx.OnError(fmt.Errorf("processing failed: %w", err))
        
        // Return graceful fallback
        fallback := fmt.Sprintf("[Error: %v]", operand)
        return messagevalue.NewStringValue(fallback, locale, source)
    }
    
    return messagevalue.NewStringValue(result, locale, source)
}

func processValue(operand interface{}, options map[string]interface{}) (string, error) {
    // Your processing logic here
    return fmt.Sprintf("processed: %v", operand), nil
}
```

## Summary

Custom functions extend MessageFormat with domain-specific formatting. Implement the `MessageFunction` type, handle errors gracefully, and use the context for locale-aware processing. Register functions during MessageFormat creation or use a registry for reusability.

Key points:
- Functions receive context, options, and operand
- Return appropriate `MessageValue` types
- Handle errors with `ctx.OnError()` and fallback values
- Use locale from context for internationalization
- Register functions via `WithFunction()` or `WithFunctions()`
