# V1 API Reference

Reference for the supported ICU MessageFormat v1 compatibility package.

## Import

```go
import mf "github.com/kaptinlin/messageformat-go/v1"
```

## Construction

Create a formatter with `New`:

```go
messageFormat, err := mf.New("en", nil)
```

The first argument is the locale. The second is an optional `*MessageFormatOptions`.

## Core Flow

Typical usage is compile once, then execute:

```go
messageFormat, err := mf.New("en", nil)
if err != nil {
	return
}

compiled, err := messageFormat.Compile("Hello, {name}!")
if err != nil {
	return
}

result, err := compiled(map[string]any{"name": "World"})
if err != nil {
	return
}
```

## Main Types

- `MessageFormat`: locale-aware compiler and formatter
- `MessageFormatOptions`: construction-time options
- `ResolvedOptions`: resolved runtime options

## Main Entry Points

- `New(locale, options)`: construct a formatter
- `(*MessageFormat).Compile(pattern)`: compile an ICU MessageFormat v1 pattern
- `(*MessageFormat).ResolvedOptions()`: inspect resolved configuration
- `SupportedLocalesOf(locales)`: report supported locales

## Pattern Features

The package supports the usual ICU MessageFormat v1 features:

- variable substitution
- plural branches
- select branches
- locale-aware number handling
- custom formatters through options

## Compatibility Note

This package is shipped from the root module. Use:

```bash
go get github.com/kaptinlin/messageformat-go@latest
```

Do not treat `github.com/kaptinlin/messageformat-go/v1` as a standalone Go module.
