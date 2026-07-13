# MF1 API Reference

Reference for the supported ICU MessageFormat v1 compatibility module.

## Import

```go
import mf "github.com/kaptinlin/messageformat-go/mf1"
```

## Construction

Construct from a locale with `New`:

```go
messageFormat, err := mf.New("en", nil)
```

The first argument is the locale. The second is an optional `*MessageFormatOptions`.

Construct from a custom plural function with `NewWithPlural`:

```go
messageFormat, err := mf.NewWithPlural(customPlural, nil)
```

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

- `New(locale string, options *MessageFormatOptions)`: construct from one locale
- `NewWithPlural(plural PluralFunction, options *MessageFormatOptions)`: construct from a custom plural function
- `(*MessageFormat).Compile(pattern)`: compile an ICU MessageFormat v1 pattern
- `(*MessageFormat).ResolvedOptions()`: inspect resolved configuration
- `SupportedLocalesOf(locales []string)`: report supported locales
- `GetPlural(locale string)`: resolve one locale's plural behavior

## Pattern Features

The package supports the usual ICU MessageFormat v1 features:

- variable substitution
- plural branches
- select branches
- locale-aware number handling
- custom formatters through options

## Module Installation

Install the independent module whose path is declared by `mf1/go.mod`:

```bash
go get github.com/kaptinlin/messageformat-go/mf1@latest
```

The module depends on `go-intl`, not on the root MessageFormat 2 module.
