# Contributing to Notelink

Thank you for your interest in contributing to Notelink! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

1. **Fork the Repository**
   ```bash
   # Fork the repository on GitHub
   # Then clone your fork
   git clone https://github.com/YOUR_USERNAME/notelink.git
   cd notelink
   ```

2. **Add Upstream Remote**
   ```bash
   git remote add upstream https://github.com/canvas-tech-horizon/notelink.git
   ```

3. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or for bug fixes
   git checkout -b fix/issue-description
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- A text editor or IDE

### Installation

1. Clone the repository (see Getting Started)

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Verify the setup:
   ```bash
   go test ./...
   go build ./...
   ```

4. Run the example application:
   ```bash
   cd examples
   go run main.go
   ```
   The API documentation should be available at http://localhost:8080/api-docs

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

- **Bug Fixes**: Fix issues reported in GitHub Issues
- **Features**: Add new functionality (discuss in an issue first for major features)
- **Documentation**: Improve README, code comments, or add examples
- **Tests**: Add or improve test coverage
- **Performance**: Optimize existing code
- **Refactoring**: Improve code quality without changing functionality

### Before Starting

1. **Check Existing Issues**: Look for existing issues or create a new one
2. **Discuss Major Changes**: For significant features, create an issue to discuss first
3. **One Feature Per PR**: Keep pull requests focused on a single change

## Coding Standards

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Use `golangci-lint` for linting
- Write clear, self-documenting code with minimal comments
- Add comments for exported functions and complex logic

### Code Organization

```go
// Package comment describing the package
package notelink

// Imports grouped and sorted
import (
    "fmt"
    "strings"

    "github.com/gofiber/fiber/v3"
)

// Exported constants and variables
const (
    DefaultPort = 8080
)

// Exported types with documentation
// User represents a user in the system
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Exported functions with documentation
// NewUser creates a new user with the given parameters
func NewUser(id int, name string) *User {
    return &User{
        ID:   id,
        Name: name,
    }
}
```

### Best Practices

1. **Error Handling**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to process data: %w", err)
   }

   // Avoid
   if err != nil {
       panic(err) // Don't panic in library code
   }
   ```

2 **Testing**
   - Write tests for new code
   - Maintain or improve code coverage
   - Use table-driven tests where appropriate

3. **Security**
   - Sanitize all user inputs
   - Escape HTML and JavaScript properly
   - Never hardcode secrets
   - Validate all request data

4. **Performance**
   - Avoid unnecessary allocations
   - Use appropriate data structures
   - Profile before optimizing

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific tests
go test -run TestValidate ./...

# Run with race detector
go test -race ./...
```

### Writing Tests

1. **Unit Tests**
   ```go
   func TestValidateParameters(t *testing.T) {
       tests := []struct {
           name        string
           input       Parameter
           expectError bool
       }{
           {
               name: "Valid parameter",
               input: Parameter{Name: "id", Type: "integer"},
               expectError: false,
           },
           // Add more test cases
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test implementation
           })
       }
   }
   ```

2. **Benchmark Tests**
   ```go
   func BenchmarkValidation(b *testing.B) {
       for i := 0; i < b.N; i++ {
           // Code to benchmark
       }
   }
   ```

### Test Coverage

- Aim for at least 70% code coverage
- Focus on critical paths and edge cases
- Don't test vendor code or generated code

## Pull Request Process

### Before Submitting

1. **Update Your Branch**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run Tests**
   ```bash
   go test ./...
   go test -race ./...
   ```

3. **Run Linters**
   ```bash
   golangci-lint run
   ```

4. **Format Code**
   ```bash
   gofmt -w .
   go mod tidy
   ```

### Commit Messages

Use clear, descriptive commit messages following this format:

```
<type>: <short summary>

<optional detailed description>

<optional footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat: add recursive validation for nested structs

Implement recursive validation for deeply nested structures
and arrays to ensure all fields are properly validated.

Closes #123
```

```
fix: escape HTML in endpoint descriptions

Add HTML escaping to prevent XSS vulnerabilities in
user-provided endpoint descriptions.
```

### Creating a Pull Request

1. **Push Your Changes**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR on GitHub**
   - Provide a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Add "Closes #XXX" to auto-close issues

3. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Testing
   - [ ] Tests pass locally
   - [ ] Added new tests
   - [ ] Updated documentation

   ## Related Issues
   Closes #XXX
   ```

### Review Process

1. **Automated Checks**: CI/CD will run tests and linters
2. **Code Review**: Maintainers will review your code
3. **Address Feedback**: Make requested changes
4. **Approval**: Once approved, maintainers will merge

## Reporting Issues

### Bug Reports

Include:
- **Description**: Clear description of the bug
- **Steps to Reproduce**: Detailed steps to reproduce the issue
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: Go version, OS, etc.
- **Code Sample**: Minimal reproducible example

### Feature Requests

Include:
- **Use Case**: Why is this feature needed?
- **Proposed Solution**: How should it work?
- **Alternatives**: Other solutions you've considered
- **Additional Context**: Any other relevant information

## Development Workflow

1. Create an issue or pick an existing one
2. Fork and create a feature branch
3. Write code and tests
4. Run tests and linters
5. Commit with clear messages
6. Push and create a pull request
7. Address review feedback
8. Merge (maintainers only)

## Questions?

- **General Questions**: Open a GitHub Discussion
- **Bug Reports**: Create a GitHub Issue
- **Security Issues**: See [SECURITY.md](SECURITY.md)

## License

By contributing to Notelink, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Notelink!
