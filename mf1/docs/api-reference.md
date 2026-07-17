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

Construct from complete custom plural facts with `NewWithPlural`:

```go
messageFormat, err := mf.NewWithPlural(mf.PluralProfile{
    Locale:    "fr",
    Select:    selectPlural,
    Cardinals: []mf.PluralCategory{mf.PluralOne, mf.PluralOther},
    Ordinals:  []mf.PluralCategory{mf.PluralOther},
}, nil)
```

Construction validates the locale, selector, and category sets and snapshots
both category slices.

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

result, err := compiled.Format(map[string]any{"name": "World"})
if err != nil {
	return
}
```

## Main Types

- `MessageFormat`: locale-aware compiler and formatter
- `CompiledMessage`: immutable parsed message with typed text and values projections
- `Formatter`: typed custom formatter receiving value, effective locale, and style
- `MessageFormatOptions`: construction-time options
- `PluralProfile`: locale and category facts for custom plural selection
- `ResolvedMessageFormatOptions`: resolved runtime options

## Main Entry Points

- `New(locale string, options *MessageFormatOptions)`: construct from one locale
- `NewWithPlural(profile PluralProfile, options *MessageFormatOptions)`: construct from custom plural facts
- `(*MessageFormat).Compile(pattern) (*CompiledMessage, error)`: compile an ICU MessageFormat v1 pattern
- `(*CompiledMessage).Format(values) (string, error)`: render text
- `(*CompiledMessage).FormatValues(values) ([]any, error)`: render ordered values
- `(*MessageFormat).ResolvedOptions()`: inspect resolved configuration
- `SupportedLocalesOf(locales []string)`: report supported locales
- `GetPlural(locale string)`: resolve one locale's plural behavior

Both projection methods accept `map[string]any`; nil means an empty map.
Without `RequireAllArguments`, a missing plain argument contributes an empty
string. With it enabled, either projection returns `ErrMissingArgument`.

## Pattern Features

The package supports the usual ICU MessageFormat v1 features:

- variable substitution
- plural branches
- select branches
- locale-aware number, date, and time formatting through `go-intl`
- number styles: `integer`, `percent`, and `currency[:CODE]`
- date/time styles: `default`, `short`, `long`, and `full`
- custom formatters through options

Unknown built-in styles return `ErrInvalidFormatterStyle`. The constructor's
`Currency` and `TimeZone` values apply to compiled built-in arguments.

Custom formatters use one typed contract:

```go
formatter := func(value any, locale, style string) (string, error) {
    return locale + ":" + style + ":" + value.(string), nil
}

messageFormat, err := mf.New("en", &mf.MessageFormatOptions{
    CustomFormatters: map[string]mf.Formatter{"label": formatter},
})
```

Construction rejects invalid names and nil handlers with
`ErrInvalidFormatter` and snapshots the map. Formatter errors retain their
identity through `Format`; handlers must be safe for concurrent calls when a
compiled message is formatted concurrently.

## Module Installation

Install the independent module whose path is declared by `mf1/go.mod`:

```bash
go get github.com/kaptinlin/messageformat-go/mf1@latest
```

The module depends on `go-intl`, not on the root MessageFormat 2 module.
