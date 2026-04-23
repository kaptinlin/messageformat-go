# Testing Guide - MessageFormat Go

This guide covers testing the MessageFormat Go library.

## 📋 Testing Structure Overview

This repository validates the current Unicode MessageFormat 2.0 implementation through API, package, and specification tests.

## 🏆 Specification Compliance

- **MessageFormat 2.0**: Passes the official MessageFormat 2.0 test suite
- **Unified Management**: Single go.mod and versioning for the current implementation

## 🚀 Quick Start

### Prerequisites

```bash
# Initialize git submodules (required for official tests)
git submodule update --init --recursive

# Verify submodule initialization
ls tests/messageformat-wg/test/tests/
```

**Requirements**: Go 1.26.2+, Git

### Running Tests

#### All Tests

```bash
# Run all tests with race detection
task test

# Run with coverage report
task test-coverage

# Run with verbose output
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
4. **Repository Regression Tests** (`legacy_pruning_test.go`): Ensure legacy surfaces stay removed

### File Organization

```text
messageformat-go/
├── legacy_pruning_test.go            # Legacy-surface regression coverage
├── messageformat_test.go             # API tests
├── tests/                            # Official test suite
│   ├── basic_test.go                 # Core behavior coverage
│   ├── bench_test.go                 # Benchmark helpers
│   ├── features_test.go              # Feature compliance coverage
│   ├── messageformat-wg/             # Git submodule
│   ├── spec_test.go                  # Official suite runner
│   └── utils/                        # Test helpers
├── pkg/                              # Package tests
│   └── */*_test.go                   # Component tests
├── internal/                         # Internal tests
│   └── */*_test.go                   # Internal tests
└── examples/                         # Example programs
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
task test-v2        # Package and official tests
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
