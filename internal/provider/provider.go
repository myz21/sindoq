// Package provider defines the interfaces for sandbox providers.
package provider

import (
	"context"
	"time"

	"github.com/happyhackingspace/sindoq/pkg/event"
	"github.com/happyhackingspace/sindoq/pkg/executor"
	"github.com/happyhackingspace/sindoq/pkg/fs"
)

// Provider defines the interface that all sandbox providers must implement.
// This is the Strategy pattern - each provider has its own implementation.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// Create initializes a new sandbox instance.
	Create(ctx context.Context, opts *CreateOptions) (Instance, error)

	// Capabilities returns what this provider supports.
	Capabilities() Capabilities

	// Validate checks if the provider is available and configured.
	Validate(ctx context.Context) error

	// Close releases provider resources.
	Close() error
}

// Instance represents a sandbox instance from a specific provider.
type Instance interface {
	// ID returns the unique instance identifier.
	ID() string

	// Execute runs code in the sandbox.
	Execute(ctx context.Context, code string, opts *executor.ExecutionOptions) (*executor.ExecutionResult, error)

	// ExecuteStream runs code with streaming output.
	ExecuteStream(ctx context.Context, code string, opts *executor.ExecutionOptions, handler executor.StreamHandler) error

	// RunCommand executes a shell command.
	RunCommand(ctx context.Context, cmd string, args []string) (*executor.CommandResult, error)

	// FileSystem returns the file system handler.
	FileSystem() fs.FileSystem

	// Network returns the network handler (may be nil if not supported).
	Network() Network

	// Stop terminates the instance.
	Stop(ctx context.Context) error

	// Status returns the current status.
	Status(ctx context.Context) (InstanceStatus, error)
}

// InstanceStatus represents the current state of an instance.
type InstanceStatus string

const (
	StatusCreating  InstanceStatus = "creating"
	StatusRunning   InstanceStatus = "running"
	StatusExecuting InstanceStatus = "executing"
	StatusPaused    InstanceStatus = "paused"
	StatusStopped   InstanceStatus = "stopped"
	StatusError     InstanceStatus = "error"
)

// Capabilities describes what a provider supports.
type Capabilities struct {
	// SupportsStreaming indicates if real-time output streaming is supported.
	SupportsStreaming bool

	// SupportsAsync indicates if async execution is supported.
	SupportsAsync bool

	// SupportsFileSystem indicates if file operations are supported.
	SupportsFileSystem bool

	// SupportsNetwork indicates if network/port publishing is supported.
	SupportsNetwork bool

	// SupportedLanguages lists supported programming languages.
	SupportedLanguages []string

	// MaxExecutionTime is the maximum execution duration.
	MaxExecutionTime time.Duration

	// MaxMemoryMB is the maximum memory in megabytes.
	MaxMemoryMB int

	// MaxCPUs is the maximum CPU count.
	MaxCPUs int

	// SupportsGPU indicates if GPU acceleration is available.
	SupportsGPU bool

	// SupportsPersistence indicates if sandbox state can be persisted.
	SupportsPersistence bool
}

// CreateOptions configures sandbox creation.
type CreateOptions struct {
	// Image specifies the container/VM image.
	Image string

	// Runtime specifies the language runtime (e.g., "python3.12", "node22").
	Runtime string

	// Resources defines resource limits.
	Resources ResourceConfig

	// Environment variables.
	Environment map[string]string

	// Timeout for the sandbox lifetime.
	Timeout time.Duration

	// Labels for tagging/identification.
	Labels map[string]string

	// WorkDir is the initial working directory.
	WorkDir string

	// InternetAccess controls network access.
	InternetAccess bool

	// Metadata is provider-specific configuration.
	Metadata map[string]any
}

// DefaultCreateOptions returns sensible defaults.
func DefaultCreateOptions() *CreateOptions {
	return &CreateOptions{
		Resources: ResourceConfig{
			MemoryMB: 512,
			CPUs:     1,
			DiskMB:   1024,
		},
		Environment:    make(map[string]string),
		Timeout:        5 * time.Minute,
		Labels:         make(map[string]string),
		WorkDir:        "/workspace",
		InternetAccess: false,
		Metadata:       make(map[string]any),
	}
}

// ResourceConfig specifies resource limits.
type ResourceConfig struct {
	// MemoryMB is the memory limit in megabytes.
	MemoryMB int

	// CPUs is the CPU limit (can be fractional).
	CPUs float64

	// DiskMB is the disk space limit in megabytes.
	DiskMB int
}

// Network provides network operations for a sandbox.
type Network interface {
	// PublishPort exposes a port publicly.
	PublishPort(ctx context.Context, port int) (*PublishedPort, error)

	// GetPublicURL returns the public URL for an exposed port.
	GetPublicURL(port int) (string, error)

	// ListPorts returns all published ports.
	ListPorts(ctx context.Context) ([]*PublishedPort, error)

	// UnpublishPort removes port exposure.
	UnpublishPort(ctx context.Context, port int) error
}

// PublishedPort represents an exposed network port.
type PublishedPort struct {
	// LocalPort is the port inside the sandbox.
	LocalPort int

	// PublicPort is the externally accessible port.
	PublicPort int

	// Protocol is the network protocol (tcp, udp).
	Protocol string

	// PublicURL is the full URL to access this port.
	PublicURL string
}

// EventAwareInstance extends Instance with event capabilities.
type EventAwareInstance interface {
	Instance

	// EventBus returns the event emitter for this instance.
	EventBus() event.Emitter
}

// PoolableInstance extends Instance with pool management capabilities.
type PoolableInstance interface {
	Instance

	// Reset cleans up the instance for reuse.
	Reset(ctx context.Context) error

	// LastUsed returns when the instance was last used.
	LastUsed() time.Time

	// UseCount returns how many times the instance has been used.
	UseCount() int
}
