# Testing Guide - MessageFormat Go

This guide covers testing the MessageFormat Go library with its unified architecture.

## 📋 Testing Structure Overview

This repository contains a unified MessageFormat implementation with both V1 and V2 functionality under single version management. Tests are organized to validate both implementations while maintaining unified versioning.

## 🏆 Specification Compliance

- **MessageFormat 2.0**: Passes the official MessageFormat 2.0 test suite
- **ICU MessageFormat (V1)**: Maintains compatibility with ICU specification and TypeScript messageformat.js library
- **Unified Management**: Single go.mod and versioning for both implementations

## 🚀 Quick Start

### Prerequisites

```bash
# Initialize git submodules (required for official tests)
git submodule update --init --recursive

# Verify submodule initialization
ls tests/messageformat-wg/test/tests/
```

**Requirements**: Go 1.25+, Git

### Running Tests

#### All Tests

```bash
# Run all tests (V1 + V2) with race detection
task test

# Run with coverage report
task test-coverage

# Run with verbose output
task test-verbose
```

#### Version-Specific Testing

```bash
# V1 Tests (ICU MessageFormat)
task test-v1

# V2 Tests (MessageFormat 2.0, includes official test suite)
task test-v2

# Official MessageFormat 2.0 test suite only
task test-official
```

#### Examples and Benchmarks

```bash
# Run all examples (V1 + V2)
task examples

# Run benchmarks
task bench
```

## 📁 Test Structure

### Test Categories

#### MessageFormat 2.0 Tests

1. **Official Test Suite** (`./tests/`): Unicode MessageFormat Working Group tests
2. **API Tests** (`messageformat_test.go`): Constructor and formatting methods
3. **Feature Tests** (`features_test.go`): MessageFormat 2.0 feature compliance
4. **Package Tests** (`./pkg/`, `./internal/`): Component-specific functionality

#### ICU MessageFormat V1 Tests

1. **Core API Tests** (`v1/messageformat_test.go`): Constructor and compilation
2. **Parser Tests** (`v1/parse_test.go`): Message parsing and validation
3. **Compatibility Tests** (`v1/typescript_compatibility_test.go`): TypeScript API compatibility
4. **Performance Tests** (`v1/benchmarks_test.go`): Performance and memory optimization

### File Organization

```text
messageformat-go/
├── messageformat_test.go              # V2 API tests
├── features_test.go                   # V2 feature compliance
├── messageformat_bench_test.go        # V2 benchmarks
├── tests/                             # V2 official test suite
│   ├── messageformat-wg/             # Git submodule
│   └── spec_test.go                  # V2 test runner
├── pkg/                              # V2 package tests
│   └── */*_test.go                   # Component tests
├── internal/                         # V2 internal tests
│   └── */*_test.go                   # Internal tests
└── v1/                               # V1 tests
    ├── messageformat_test.go         # V1 API tests
    ├── parse_test.go                 # V1 parser tests
    ├── typescript_compatibility_test.go # V1 compatibility
    └── benchmarks_test.go            # V1 benchmark tests
```

## 🔧 Development Commands

### Code Quality

```bash
# Format, vet, lint, and test
task verify

# Individual checks
task fmt          # Format code
task vet          # Static analysis
task lint         # Comprehensive linting
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
git submodule update --init --recursive
```

**Test files missing:**

```bash
git submodule status
ls tests/messageformat-wg/test/tests/
```

**Go module issues:**

```bash
go mod download
go mod verify
go mod tidy
```

**Linting tool missing:**

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Debug Commands

```bash
# Verbose with no cache
go test -v -count=1 ./...

# Race detection
go test -race ./...

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
task test-v2        # V2 tests
task test-coverage  # With coverage
task verify         # All quality checks
task bench          # Benchmarks
task help           # Show all targets

# Debug & troubleshoot
git submodule update --init --recursive
go test -v -race ./...
go mod verify
```

---

**Ready to start testing?** Run `task test` to execute the complete test suite and verify MessageFormat 2.0 specification compliance.
