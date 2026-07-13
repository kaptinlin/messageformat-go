# Testing Guide - MessageFormat Go

This guide covers testing the MessageFormat Go library.

## 📋 Testing Structure Overview

This repository validates the current Unicode MessageFormat 2.0 implementation through API, package, and specification tests.

## 🏆 Specification Compliance

- **MessageFormat 2.0**: `task test-official` checks the implementation against the corpus pinned by `tests/messageformat-wg`
- **Module Coverage**: Root MF2 and `mf1/` have independent `go.mod` files and are both covered by aggregate gates

## 🚀 Quick Start

### Prerequisites

```bash
# Initialize the fixture required by official tests
task submodules

# Verify submodule initialization
ls tests/messageformat-wg/test/tests/
```

**Requirements**: The Go toolchain declared by `go.mod` and `mf1/go.mod`, plus Git and Task

### Running Tests

#### All Tests

```bash
# Run all tests with race detection
task test

# Run root tests with a coverage report
task test-coverage

# Run root tests with verbose output
task test-verbose
```

#### Focused Testing

```bash
# Run package and official MessageFormat 2.0 tests
task test-v2

# Official MessageFormat 2.0 test suite only
task test-official
```

#### Examples and Benchmarks

```bash
# Run all examples
task examples

# Run benchmarks
task bench
```

## 📁 Test Structure

### Test Categories

1. **Official Test Suite** (`./tests/`): Unicode MessageFormat Working Group tests
2. **API Tests** (`messageformat_test.go`): Constructor and formatting methods
3. **Package Tests** (`./pkg/`, `./internal/`): Component-specific functionality
4. **MF1 Module Tests** (`mf1/*_test.go`): ICU MessageFormat v1 behavior and examples

### File Organization

```text
messageformat-go/
├── messageformat_test.go             # Root public API tests
├── tests/
│   ├── official_test.go              # Pinned official corpus runner
│   ├── messageformat-wg/             # Official corpus git submodule
│   └── utils/mfwg_test.go            # Corpus adapter tests
├── pkg/*/*_test.go                   # Public package tests
├── internal/*/*_test.go              # Internal package tests
├── examples/*/main_test.go           # Root example tests
└── mf1/                              # Independent MF1 module
    ├── go.mod
    ├── *_test.go                     # MF1 API and behavior tests
    └── examples/*/main_test.go        # MF1 example tests
```

## 🔧 Development Commands

### Code Quality

```bash
# Read-only checks for root and MF1
task verify

# Explicit write and individual checks
task fmt          # Rewrite Go formatting
task vet          # Static analysis for both modules
task lint         # Tidy and golangci-lint checks for both modules
```

### Coverage and Benchmarks

```bash
# Coverage report
task test-coverage

# Benchmarks
task bench
```

## 🛠️ Troubleshooting

### Common Issues

**Submodule not initialized:**

```bash
task submodules
```

**Test files missing:**

```bash
git submodule status
ls tests/messageformat-wg/test/tests/
```

**Go module issues:**

```bash
task deps       # Explicitly download and tidy root dependencies
task tidy-lint  # Check both module graphs without rewriting them
```

**Linting tool missing:**

```bash
task install-golangci-lint
```

### Debug Commands

```bash
# Verbose with no cache
go test -v -count=1 ./...

# Race detection
go test -race -count=1 ./...

# MF1 module only
(cd mf1 && go test -race -count=1 ./...)

# Specific test
go test -v -run TestSpecificFunction ./pkg/functions/
```

## 📝 Contributing Tests

### Test Guidelines

1. **Follow patterns**: Use table-driven tests
2. **Comprehensive coverage**: Positive, negative, edge cases
3. **Clear naming**: `TestFunctionName`, `TestFunctionName_ErrorCase`
4. **Include benchmarks**: For performance-critical code
5. **Maintain compliance**: Don't break official test suite

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        // Test cases
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 🎯 Quick Reference

```bash
# Essential commands
task test           # Run all tests
task test-v2        # Package and official tests
task test-coverage  # With coverage
task verify         # All quality checks
task bench          # Benchmarks
task help           # Show all targets

# Debug & troubleshoot
task submodules
go test -v -race -count=1 ./...
task tidy-lint
```

---

**Ready to start testing?** Run `task test` to execute fresh race-enabled tests for both modules.
