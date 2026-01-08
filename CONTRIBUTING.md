# Contributing to sindoq

Thanks for your interest in contributing to sindoq!

## Getting Started

```bash
# Clone the repo
git clone https://github.com/happyhackingspace/sindoq.git
cd sindoq

# Install dependencies
go mod download

# Run tests
go test ./...
```

## Development Setup

**Requirements:**
- Go 1.25+
- Docker (for local testing)

**Optional:**
- Podman
- Kubernetes cluster (minikube/kind)

## Project Structure

```
sindoq/
├── sindoq.go           # Main SDK entry point
├── config.go           # Configuration and options
├── errors.go           # Error types
├── cmd/sindoq/         # CLI tool
├── pkg/
│   ├── langdetect/     # Language detection
│   ├── executor/       # Execution types
│   ├── event/          # Event system
│   └── fs/             # Filesystem interface
├── internal/
│   ├── provider/       # Provider implementations
│   │   ├── docker/
│   │   ├── vercel/
│   │   ├── e2b/
│   │   ├── kubernetes/
│   │   ├── podman/
│   │   └── firecracker/
│   └── factory/        # Provider factory
├── examples/           # Example programs
└── testutil/           # Test utilities
```

## Making Changes

### 1. Fork and Branch

```bash
git checkout -b feature/your-feature
```

### 2. Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Keep functions small and focused
- Add tests for new functionality

### 3. Testing

```bash
# Unit tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test ./... -cover

# Integration tests (requires Docker)
go test -tags=integration ./internal/provider/docker/...
```

### 4. Commit Messages

Use clear, descriptive commit messages:

```
Add streaming support for Kubernetes provider

- Implement ExecuteStream for k8s pods
- Add log streaming via pod logs API
- Update tests
```

### 5. Pull Request

- Describe what changes you made
- Reference any related issues
- Ensure all tests pass
- Update documentation if needed

## Adding a New Provider

1. Create directory: `internal/provider/yourprovider/`

2. Implement the provider interface:

```go
package yourprovider

import (
    "github.com/happyhackingspace/sindoq/internal/factory"
    "github.com/happyhackingspace/sindoq/internal/provider"
)

func init() {
    factory.Register("yourprovider", New)
}

func New(config any) (provider.Provider, error) {
    // Initialize provider
}

type Provider struct {
    // Provider state
}

func (p *Provider) Name() string { return "yourprovider" }
func (p *Provider) Create(ctx context.Context, opts *provider.CreateOptions) (provider.Instance, error) { ... }
func (p *Provider) Capabilities() provider.Capabilities { ... }
func (p *Provider) Validate(ctx context.Context) error { ... }
func (p *Provider) Close() error { ... }
```

3. Implement the instance interface:

```go
type Instance struct {
    // Instance state
}

func (i *Instance) ID() string { ... }
func (i *Instance) Provider() string { ... }
func (i *Instance) Execute(ctx context.Context, code string, opts *executor.ExecutionOptions) (*executor.ExecutionResult, error) { ... }
func (i *Instance) ExecuteStream(ctx context.Context, code string, opts *executor.ExecutionOptions, handler executor.StreamHandler) error { ... }
func (i *Instance) RunCommand(ctx context.Context, cmd string, args []string) (*executor.CommandResult, error) { ... }
func (i *Instance) FileSystem() fs.FileSystem { ... }
func (i *Instance) Network() provider.Network { ... }
func (i *Instance) Stop(ctx context.Context) error { ... }
func (i *Instance) Status(ctx context.Context) (provider.InstanceStatus, error) { ... }
```

4. Add blank import to `cmd/sindoq/main.go`:

```go
_ "github.com/happyhackingspace/sindoq/internal/provider/yourprovider"
```

5. Add configuration to `config.go`:

```go
type YourProviderConfig struct {
    // Config fields
}

func WithYourProviderConfig(cfg YourProviderConfig) Option {
    return func(c *Config) {
        c.Provider = "yourprovider"
        c.ProviderConfig = cfg
    }
}
```

6. Add tests and documentation

## Adding Language Support

Edit `pkg/langdetect/runtime.go`:

```go
var defaultRuntimes = map[string]*RuntimeInfo{
    // Add your language
    "YourLang": {
        Language:    "YourLang",
        Runtime:     "yourlang",
        FileExt:     ".yl",
        DockerImage: "yourlang:latest",
        RunCmd:      []string{"yourlang", "run"},
    },
}
```

Add detection patterns in `pkg/langdetect/detect.go`:

```go
"YourLang": {
    `(?m)^yourlang_keyword`,
    `unique_pattern`,
},
```

## Reporting Issues

- Check existing issues first
- Include Go version, OS, and provider
- Provide minimal reproduction steps
- Include relevant error messages

## Questions?

Open an issue or discussion on GitHub.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
