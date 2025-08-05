# Migration Guide: Restructuring kubectl-passman

## Overview

This document outlines the restructuring of kubectl-passman to follow Go best practices.

## Changes Made

### 1. Project Structure

**Before:**
```
kubectl-passman/
├── app.go
├── app_test.go
├── 1password.go
├── 1password_test.go
├── gopass.go
├── gopass_test.go
├── keychain_default.go
├── keychain_mac.go
├── keychain_mac_test.go
└── go.mod
```

**After:**
```
kubectl-passman/
├── cmd/kubectl-passman/     # Main application entry point
│   └── main.go
├── internal/                # Private application code
│   ├── cli/                 # CLI application logic
│   │   └── app.go
│   └── providers/           # Password manager providers
│       ├── provider.go      # Provider interface
│       ├── keychain_*.go    # Keychain implementations
│       ├── onepassword.go   # 1Password implementation
│       └── gopass.go        # Gopass implementation
├── pkg/passman/             # Public library code
│   ├── response.go          # Kubernetes credential response types
│   └── response_test.go
└── Makefile                 # Build automation
```

### 2. Package Organization

- **`cmd/kubectl-passman/`**: Contains only the main entry point
- **`internal/cli/`**: Contains CLI application logic
- **`internal/providers/`**: Contains password manager provider implementations
- **`pkg/passman/`**: Contains public API for credential formatting

### 3. Interface Design

Created a clean `Provider` interface:

```go
type Provider interface {
    Get(itemName string) (string, error)
    Set(itemName, secret string) error
    Name() string
}
```

### 4. Build Constraints

Improved build constraint handling:
- Platform-specific keychain implementations
- Factory pattern for provider selection

### 5. Testing

- Separated tests by package
- Added comprehensive test coverage
- Maintained backward compatibility

### 6. Build System

Added Makefile with common development tasks:
- `make build`: Build the binary
- `make test`: Run tests
- `make lint`: Format and vet code
- `make build-all`: Cross-platform builds

## Benefits

1. **Maintainability**: Clear separation of concerns
2. **Testability**: Better isolation for unit testing
3. **Extensibility**: Easy to add new providers
4. **Go Standards**: Follows Go project layout standards
5. **Documentation**: Better project documentation

## Backward Compatibility

The restructured code maintains full backward compatibility:
- Same CLI interface
- Same functionality
- Same dependencies
- Same build output

## Development Workflow

1. Clone the repository
2. Use `make dev-build` for development
3. See `CONTRIBUTING.md` for detailed guidelines
4. Follow the established patterns for new features

## Breaking Changes

None - this is purely a structural improvement that maintains API compatibility.
