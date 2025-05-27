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

### All Tests

```bash
# Run all tests including official test suite
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Specific Test Suites

```bash
# Run only the official MessageFormat 2.0 test suite
go test ./tests/

# Run main package tests (includes advanced features)
go test .

# Run specific package tests
go test ./pkg/datamodel
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

#### 3. Package Tests (`./pkg/`, `./internal/`)
- **Coverage**: Individual package functionality
- **Tests Include**:
  - Data model validation
  - Function implementations
  - Parser functionality
  - Value types

## Test Structure

```
messageformat/
├── messageformat_test.go          # Main API tests
├── tests/                         # Official test suite
│   ├── messageformat-wg/         # Git submodule
│   │   └── test/tests/           # Official test files
│   ├── spec_test.go              # Official test runner
│   └── basic_test.go             # Basic functionality tests
├── pkg/                          # Package-specific tests
│   ├── functions/
│   │   ├── number_test.go
│   │   ├── datetime_test.go
│   │   └── ...
│   └── messagevalue/
│       ├── value_test.go
│       └── ...
└── internal/                     # Internal package tests
    ├── cst/
    └── resolve/
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
```

## Continuous Integration

The project includes comprehensive CI testing that:

1. Initializes submodules automatically
2. Runs all test suites
3. Checks code coverage
4. Validates against multiple Go versions
5. Ensures cross-platform compatibility

## Test Coverage

Current test coverage includes:

- ✅ **100% Official Test Suite Compliance**
- ✅ **Complete API Coverage**
- ✅ **Error Handling**
- ✅ **Unicode Support**
- ✅ **Bidirectional Text**
- ✅ **Custom Functions**
- ✅ **Advanced Features**

## Contributing

When adding new tests:

1. Follow existing test patterns
2. Include both positive and negative test cases
3. Test error conditions
4. Ensure compatibility with official test suite
5. Update documentation as needed

For more information, see [CONTRIBUTING.md](CONTRIBUTING.md). 