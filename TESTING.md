# Testing Guide - MessageFormat 2.0 Go Implementation

This guide covers testing the MessageFormat 2.0 Go implementation for compliance with the official Unicode specification.

## 🏆 Specification Compliance

This implementation passes the official MessageFormat 2.0 test suite, ensuring:
- Compatibility with Unicode MessageFormat 2.0 specification
- Interoperability with other compliant implementations
- Consistent behavior across features and edge cases

## 🚀 Quick Start

### Prerequisites

```bash
# Initialize git submodules (required for official tests)
git submodule update --init --recursive

# Verify submodule initialization
ls tests/messageformat-wg/test/tests/
```

**Requirements**: Go 1.21+, Git

### Running Tests

```bash
# Run all tests
make test

# Unit tests only
make test-unit

# Official test suite only
make test-official

# With coverage
make test-coverage

# With verbose output
go test -v ./...
```

## 📁 Test Structure

### Test Categories

1. **Official Test Suite** (`./tests/`)
   - Unicode MessageFormat Working Group tests
   - Specification compliance verification
   - Covers syntax, formatting, errors, Unicode, bidi text

2. **API Tests** (`messageformat_test.go`)
   - Constructor and options testing
   - Format/FormatToParts methods
   - Error handling and custom functions

3. **Feature Tests** (`features_test.go`)
   - MessageFormat 2.0 feature compliance
   - Pattern matching, number formatting
   - Markup, internationalization

4. **Package Tests** (`./pkg/`, `./internal/`)
   - Component-specific functionality
   - Functions, data model, parser, resolver

### File Organization

```
messageformat-go/
├── messageformat_test.go              # API tests
├── features_test.go                   # Feature compliance
├── messageformat_bench_test.go        # Benchmarks
├── tests/                             # Official test suite
│   ├── messageformat-wg/             # Git submodule
│   └── spec_test.go                  # Test runner
├── pkg/                              # Package tests
│   ├── functions/
│   ├── messagevalue/
│   └── datamodel/
└── internal/                         # Internal tests
    ├── cst/
    └── resolve/
```

## 🔧 Development Commands

### Code Quality

```bash
# All CI checks
make ci

# Individual checks
make fmt          # Format code
make vet          # Static analysis
make lint         # Comprehensive linting
make verify       # Format + lint + test
```

### Benchmarks

```bash
# Run benchmarks
make bench

# With memory stats
go test -bench=. -benchmem ./...

# Specific benchmarks
go test -bench=BenchmarkSimpleMessage ./...
```

### Coverage

```bash
# Generate coverage
make test-coverage

# HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
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
make test           # Run all tests
make test-unit      # Unit tests only
make test-coverage  # With coverage
make ci             # All quality checks
make bench          # Benchmarks
make help           # Show all targets

# Debug & troubleshoot
git submodule update --init --recursive
go test -v -race ./...
go mod verify
```

---

**Ready to start testing?** Run `make test` to execute the complete test suite and verify MessageFormat 2.0 specification compliance.
