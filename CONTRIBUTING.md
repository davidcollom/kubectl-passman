# Contributing to kubectl-passman

## Project Structure

This project follows Go best practices for project layout:

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

## Development

### Building

Use the provided Makefile for common development tasks:

```bash
# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Run linting
make lint

# Full development build (lint + test + build)
make dev-build

# Build for multiple platforms
make build-all
```

### Testing

Run tests for all packages:

```bash
go test ./...
```

Or use the Makefile:

```bash
make test
```

### Adding New Providers

To add a new password manager provider:

1. Create a new file in `internal/providers/` (e.g., `newprovider.go`)
2. Implement the `Provider` interface:
   ```go
   type Provider interface {
       Get(itemName string) (string, error)
       Set(itemName, secret string) error
       Name() string
   }
   ```
3. Add tests in `internal/providers/newprovider_test.go`
4. Register the provider in `internal/cli/app.go` in the `setupProviders()` method
5. Add a new CLI command in the `setupCLI()` method

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Add tests for new functionality
- Document public APIs

### Package Guidelines

- `cmd/`: Contains main application entry points
- `internal/`: Contains private application code that should not be imported by other projects
- `pkg/`: Contains public library code that can be imported by other projects
- Keep interfaces small and focused
- Use dependency injection for testing

### Build Constraints

The project uses build constraints for platform-specific code:

- `keychain_default.go`: For non-macOS systems using `go-keyring`
- `keychain_mac.go`: For macOS systems using native keychain
- `keychain_factory_*.go`: Factory functions that choose the correct implementation

Make sure to test on different platforms when modifying keychain code.
