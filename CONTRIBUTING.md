# Contributing to MessageFormat 2.0 Go Implementation

Thank you for your interest in contributing to the MessageFormat 2.0 Go implementation! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.19 or later
- Git
- Basic understanding of MessageFormat 2.0 specification
- Familiarity with Go conventions and best practices

### Development Setup

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub, then clone your fork
   git clone --recurse-submodules https://github.com/YOUR_USERNAME/messageformat-go.git
   cd messageformat-go
   
   # Add upstream remote
   git remote add upstream https://github.com/kaptinlin/messageformat-go.git
   ```

2. **Initialize Submodules**
   ```bash
   # Required for official test suite
   git submodule update --init --recursive
   ```

3. **Verify Setup**
   ```bash
   # Run all tests to ensure everything works
   go test ./...
   
   # Check code formatting
   go fmt ./...
   
   # Run linter
   go vet ./...
   ```

## 📋 Development Workflow

### 1. Create a Feature Branch

```bash
# Update your fork
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feat/your-feature-name
```

### 2. Make Changes

- Follow Go conventions and best practices
- Maintain API compatibility with TypeScript implementation
- Write comprehensive tests for new features
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run official test suite specifically
go test ./tests/

# Check formatting
go fmt ./...

# Run linter
go vet ./...
```

### 4. Commit Your Changes

Follow conventional commit format:

```bash
git add .
git commit -m "feat: add new datetime formatting option"
```

### 5. Submit Pull Request

```bash
git push origin feat/your-feature-name
```

Then create a pull request on GitHub.

## 📝 Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code (formatting, etc)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools
- `ci`: Changes to CI configuration files and scripts

### Examples

```bash
feat(functions): add custom datetime formatting options
fix(parser): handle edge case in selector parsing
docs(readme): update installation instructions
test(functions): add comprehensive number formatting tests
```

## 🏗️ Code Guidelines

### Go Conventions

1. **Follow Go Standards**
   - Use `gofmt` for formatting
   - Follow effective Go practices
   - Use meaningful variable and function names
   - Write clear, concise comments

2. **Package Structure**
   ```
   messageformat/
   ├── messageformat.go          # Main API
   ├── options.go               # Configuration options
   ├── parts.go                 # Format parts implementation
   ├── pkg/                     # Public packages
   │   ├── datamodel/          # Data model types
   │   ├── functions/          # Built-in functions
   │   └── messagevalue/       # Value types
   └── internal/               # Internal packages
       ├── cst/               # Concrete syntax tree
       └── resolve/           # Resolution logic
   ```

3. **Error Handling**
   - Use descriptive error messages
   - Include context information
   - Follow Go error handling patterns
   - Provide position information for parsing errors

### API Compatibility

1. **TypeScript Compatibility**
   - Maintain similar method signatures
   - Use similar option names and structures
   - Provide equivalent functionality

2. **Backward Compatibility**
   - Don't break existing APIs
   - Use functional options for new features
   - Deprecate features gracefully

### Testing Requirements

1. **Test Coverage**
   - Write tests for all new features
   - Include both positive and negative test cases
   - Test error conditions thoroughly
   - Maintain high test coverage

2. **Test Types**
   - Unit tests for individual functions
   - Integration tests for complete workflows
   - Compatibility tests with official test suite
   - Performance tests for critical paths

3. **Test Naming**
   ```go
   func TestFunctionName(t *testing.T) {
       t.Run("specific case description", func(t *testing.T) {
           // Test implementation
       })
   }
   ```

## 🧪 Testing Guidelines

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/functions

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Official test suite only
go test ./tests/
```

### Writing Tests

1. **Test Structure**
   ```go
   func TestNewFeature(t *testing.T) {
       t.Run("should handle valid input", func(t *testing.T) {
           // Arrange
           input := "test input"
           expected := "expected output"
           
           // Act
           result, err := NewFeature(input)
           
           // Assert
           require.NoError(t, err)
           assert.Equal(t, expected, result)
       })
       
       t.Run("should return error for invalid input", func(t *testing.T) {
           // Test error cases
       })
   }
   ```

2. **Test Data**
   - Use table-driven tests for multiple cases
   - Include edge cases and boundary conditions
   - Test with various locales and inputs

3. **Assertions**
   - Use `testify/require` for critical assertions
   - Use `testify/assert` for non-critical assertions
   - Provide descriptive failure messages

### Official Test Suite

The official MessageFormat 2.0 test suite is included as a git submodule. When contributing:

1. Ensure all official tests continue to pass
2. Don't modify official test files
3. If tests fail, fix the implementation, not the tests
4. Add new tests to supplement official coverage

## 📚 Documentation

### Code Documentation

1. **Package Documentation**
   ```go
   // Package messageformat provides a complete implementation of MessageFormat 2.0
   // for internationalization (i18n) support in Go applications.
   package messageformat
   ```

2. **Function Documentation**
   ```go
   // New creates a new MessageFormat instance with the specified locale and pattern.
   // It returns an error if the pattern is invalid or the locale is not supported.
   func New(locale, pattern string, options ...Option) (*MessageFormat, error) {
   ```

3. **Type Documentation**
   ```go
   // MessageFormat represents a compiled MessageFormat 2.0 pattern that can be
   // used to format messages with variable substitution and localization.
   type MessageFormat struct {
   ```

### README Updates

When adding new features:

1. Update feature list
2. Add usage examples
3. Update API documentation
4. Include migration notes if needed

## 🐛 Bug Reports

When reporting bugs:

1. **Use GitHub Issues**
2. **Provide Clear Title**
3. **Include Reproduction Steps**
4. **Provide Expected vs Actual Behavior**
5. **Include Environment Information**
   - Go version
   - Operating system
   - Library version

### Bug Report Template

```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- Go version: 
- OS: 
- Library version: 

## Additional Context
Any other relevant information
```

## 💡 Feature Requests

When requesting features:

1. **Check Existing Issues** first
2. **Describe the Use Case** clearly
3. **Provide Examples** of desired behavior
4. **Consider API Impact** and compatibility
5. **Reference MessageFormat 2.0 Spec** if applicable

## 🔍 Code Review Process

### For Contributors

1. **Self-Review** your code before submitting
2. **Write Clear PR Description** explaining changes
3. **Include Tests** for new functionality
4. **Update Documentation** as needed
5. **Respond to Feedback** promptly and constructively

### For Reviewers

1. **Be Constructive** and helpful
2. **Focus on Code Quality** and maintainability
3. **Check Test Coverage** and quality
4. **Verify API Compatibility**
5. **Test Changes Locally** when possible

## 📦 Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Checklist

1. Update version numbers
2. Update CHANGELOG.md
3. Run full test suite
4. Update documentation
5. Create release tag
6. Publish release notes

## 🤝 Community

### Communication

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code contributions and reviews

### Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please be respectful and professional in all interactions.

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.

## 🙏 Recognition

Contributors will be recognized in:

- README.md contributors section
- Release notes
- Git commit history

Thank you for contributing to MessageFormat 2.0 Go implementation! 