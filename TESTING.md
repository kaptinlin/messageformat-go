# Testing Guide

This document provides comprehensive instructions for running tests in the MessageFormat 2.0 Go implementation.

## Prerequisites

### Git Submodules

Before running tests, you **must** initialize the git submodules to fetch the official MessageFormat 2.0 test suite:

```bash
# Initialize and update git submodules (required for official tests)
git submodule init
git submodule update

# Or in one command
git submodule update --init --recursive
```

### For New Contributors

If you're cloning the repository for the first time:

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/kaptinlin/messageformat-go.git
cd messageformat-go

# Or if already cloned without submodules
git submodule update --init --recursive
```

## Running Tests

### Using Make Commands (Recommended)

The project includes a comprehensive Makefile with various testing targets:

```bash
# Show all available make targets
make help

# Run all tests (unit + official test suite)
make test

# Run only unit tests (faster, no submodule required)
make test-unit

# Run only official MessageFormat 2.0 test suite
make test-official

# Run tests with coverage report
make test-coverage

# Run tests with verbose output
make test-verbose

# Run benchmarks
make bench

# Run examples to verify they work
make examples
```

### Using Go Commands Directly

```bash
# Run all tests including official test suite
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test suites
go test ./tests/          # Official test suite only
go test .                 # Main package tests
go test ./pkg/datamodel   # Specific package tests
go test ./pkg/functions
go test ./internal/cst
```

### Test Categories

#### 1. Official MessageFormat 2.0 Test Suite (`./tests/`)
- **Location**: `tests/messageformat-wg/test/tests/`
- **Source**: Unicode MessageFormat Working Group
- **Coverage**: Complete specification compliance
- **Tests Include**:
  - Syntax validation
  - Formatting behavior
  - Error handling
  - Unicode normalization
  - Bidirectional text
  - Function implementations
  - Multi-selector messages

#### 2. Implementation Tests (`./`)
- **Location**: `messageformat_test.go`
- **Coverage**: Go-specific API and advanced features
- **Tests Include**:
  - Constructor variations
  - Options handling
  - Custom functions
  - Multi-selector messages
  - Format to parts
  - Error handling
  - Performance edge cases

#### 3. Package Tests (`./pkg/`, `./internal/`)
- **Coverage**: Individual package functionality
- **Tests Include**:
  - Data model validation
  - Function implementations (number, datetime, string, etc.)
  - Parser functionality
  - Value types and conversions
  - Unicode handling
  - Bidi text support

## Test Structure

```
messageformat/
├── messageformat_test.go          # Main API tests
├── tests/                         # Official test suite
│   ├── messageformat-wg/         # Git submodule
│   │   └── test/tests/           # Official test files
│   ├── spec_test.go              # Official test runner
│   ├── basic_test.go             # Basic functionality tests
│   └── utils/                    # Test utilities
├── pkg/                          # Package-specific tests
│   ├── functions/
│   │   ├── number_test.go
│   │   ├── datetime_test.go
│   │   ├── math_test.go
│   │   └── ...
│   ├── messagevalue/
│   │   ├── value_test.go
│   │   ├── number_test.go
│   │   └── ...
│   └── datamodel/
│       ├── validate_test.go
│       └── ...
└── internal/                     # Internal package tests
    ├── cst/
    │   └── parser_test.go
    └── resolve/
        └── resolve_test.go
```

## Code Quality and Linting

### Running Code Quality Checks

```bash
# Run all CI checks (formatting, linting, tests)
make ci

# Run individual checks
make fmt          # Format code
make vet          # Run go vet
make lint         # Run golangci-lint
make staticcheck  # Run staticcheck
```

### Fixing Common Issues

```bash
# Auto-fix formatting issues
make fmt

# Check for potential issues
make vet

# Run comprehensive linting
make lint
```

## Benchmarking

The project includes comprehensive benchmarks for performance testing:

```bash
# Run all benchmarks
make bench

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmarks
go test -bench=BenchmarkSimpleMessage ./...
go test -bench=BenchmarkNumberFormatting ./...
```

### Benchmark Categories

- **Simple Messages**: Basic string formatting
- **Number Formatting**: Numeric value formatting with various options
- **Select Messages**: Conditional message selection
- **Complex Messages**: Multi-selector and nested expressions
- **FormatToParts**: Detailed part-by-part formatting
- **Message Creation**: Constructor performance

## Coverage Reporting

```bash
# Generate coverage report
make test-coverage

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting

### Common Issues

#### 1. Submodule Not Initialized
```
Error: no such file or directory: tests/messageformat-wg/test/tests/
```

**Solution**: Initialize submodules
```bash
git submodule update --init --recursive
```

#### 2. Test Files Missing
```
Error: cannot find package "./tests/messageformat-wg"
```

**Solution**: Ensure submodule is properly cloned
```bash
ls tests/messageformat-wg/test/tests/
# Should show: bidi.json, functions/, syntax.json, etc.
```

#### 3. Permission Issues
```
Error: permission denied
```

**Solution**: Check file permissions and git configuration
```bash
git config --global --add safe.directory /path/to/messageformat-go
```

#### 4. Go Module Issues
```
Error: module not found
```

**Solution**: Ensure Go modules are properly initialized
```bash
go mod download
go mod verify
```

#### 5. Linting Errors
```
Error: golangci-lint not found
```

**Solution**: Install golangci-lint
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or use make target that handles installation
make lint
```

### Verification Commands

```bash
# Verify submodule status
git submodule status

# Check test files exist
ls tests/messageformat-wg/test/tests/

# Verify Go modules
go mod verify

# Check for any issues
go vet ./...

# Verify all dependencies
go mod tidy
```

### Debug Mode

For debugging test failures:

```bash
# Run tests with verbose output and no caching
go test -v -count=1 ./...

# Run specific test with debugging
go test -v -run TestSpecificFunction ./pkg/functions/

# Run with race detection
go test -race ./...
```

## Continuous Integration

The project includes comprehensive CI testing that:

1. **Multi-Platform Testing**: Ubuntu, macOS, Windows
2. **Multi-Version Testing**: Go 1.21.x, 1.22.x, 1.23.x
3. **Automatic Submodule Initialization**
4. **Comprehensive Test Coverage**:
   - Unit tests
   - Official test suite compliance
   - Code quality checks (linting, formatting)
   - Examples verification
   - Benchmark execution
5. **Performance Monitoring**: Benchmark results tracking
6. **Coverage Reporting**: Automated coverage analysis

### CI Workflow Structure

```yaml
# Simplified CI workflow overview
jobs:
  test:          # Run tests on multiple platforms/versions
  coverage:      # Generate and upload coverage reports  
  lint:          # Code quality and formatting checks
  examples:      # Verify examples work correctly
  benchmarks:    # Performance regression testing
```

## Test Coverage

Current test coverage includes:

- ✅ **100% Official Test Suite Compliance** (all 1000+ tests passing)
- ✅ **Complete API Coverage** (all public methods tested)
- ✅ **Error Handling** (comprehensive error scenarios)
- ✅ **Unicode Support** (normalization, bidi text)
- ✅ **Function Implementations** (number, datetime, string, math, etc.)
- ✅ **Custom Functions** (user-defined function support)
- ✅ **Advanced Features** (multi-selectors, format-to-parts)
- ✅ **Performance Testing** (benchmarks for all major operations)

### Coverage Statistics

- **Total Lines**: ~19,000
- **Source Code**: ~13,000 lines (49 files)
- **Test Code**: ~6,000 lines (25 files)
- **Test Coverage**: ~47% (focused on critical paths)
- **Official Tests**: 100% passing

## Contributing

When adding new tests:

1. **Follow Existing Patterns**: Use established test structures
2. **Include Both Positive and Negative Cases**: Test success and failure scenarios
3. **Test Error Conditions**: Ensure proper error handling
4. **Maintain Official Test Compatibility**: Don't break existing compliance
5. **Add Benchmarks for New Features**: Performance testing for new functionality
6. **Update Documentation**: Keep this guide current

### Test Naming Conventions

```go
// Unit tests
func TestFunctionName(t *testing.T) { ... }
func TestFunctionName_ErrorCase(t *testing.T) { ... }

// Benchmarks  
func BenchmarkFunctionName(b *testing.B) { ... }

// Examples (in _test.go files)
func ExampleFunctionName() { ... }
```

For more information about contributing, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Performance Considerations

### Test Performance Tips

1. **Use `make test-unit`** for faster iteration (skips official tests)
2. **Run specific packages** when working on isolated features
3. **Use benchmarks** to verify performance improvements
4. **Enable race detection** when debugging concurrency issues

### CI Performance

- **Parallel Execution**: Tests run in parallel where possible
- **Intelligent Caching**: Dependencies and build artifacts cached
- **Selective Testing**: Only relevant tests run for specific changes
- **Fast Feedback**: Critical tests run first for quick failure detection 
