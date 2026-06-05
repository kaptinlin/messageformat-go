# V1 Performance Guide

Performance notes for the supported ICU MessageFormat v1 compatibility package.

## General Guidance

- compile message patterns once and reuse them
- avoid recompiling the same template on hot paths
- reuse formatter instances after construction
- benchmark your own message shapes if latency matters

## Recommended Pattern

```go
messageFormat, err := mf.New("en", nil)
if err != nil {
	return
}

compiled, err := messageFormat.Compile("Hello, {name}!")
if err != nil {
	return
}

for _, name := range names {
	_, err := compiled(map[string]any{"name": name})
	if err != nil {
		return
	}
}
```

## Benchmarks

Run repository benchmarks from the root:

```bash
go test -bench=. -benchmem ./v1
```

## Concurrency

`MessageFormat` instances are intended to be safe for concurrent use after construction. Prefer sharing compiled formatters instead of rebuilding them per request.

## Scope

This document is intentionally short. The source of truth for current behavior is the package implementation and tests under `./v1`.
