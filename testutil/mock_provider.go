// Package testutil provides test utilities and mock implementations for sindoq.
package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/happyhackingspace/sindoq/internal/provider"
)

// MockProvider is a configurable mock implementation of provider.Provider.
type MockProvider struct {
	name         string
	capabilities provider.Capabilities
	instances    map[string]*MockInstance
	mu           sync.RWMutex

	// Hooks for testing
	OnCreate   func(ctx context.Context, opts *provider.CreateOptions) (provider.Instance, error)
	OnValidate func(ctx context.Context) error
	OnClose    func() error
}

// NewMockProvider creates a new mock provider with default settings.
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:      name,
		instances: make(map[string]*MockInstance),
		capabilities: provider.Capabilities{
			SupportsStreaming:  true,
			SupportsAsync:      true,
			SupportsFileSystem: true,
			SupportsNetwork:    true,
			SupportedLanguages: []string{"Python", "JavaScript", "Go"},
			MaxExecutionTime:   10 * time.Minute,
			MaxMemoryMB:        512,
			MaxCPUs:            2,
		},
	}
}

// Name returns the provider name.
func (p *MockProvider) Name() string {
	return p.name
}

// Create creates a new mock instance.
func (p *MockProvider) Create(ctx context.Context, opts *provider.CreateOptions) (provider.Instance, error) {
	if p.OnCreate != nil {
		return p.OnCreate(ctx, opts)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	instance := NewMockInstance(p)
	p.instances[instance.ID()] = instance
	return instance, nil
}

// Capabilities returns the mock capabilities.
func (p *MockProvider) Capabilities() provider.Capabilities {
	return p.capabilities
}

// SetCapabilities allows tests to configure capabilities.
func (p *MockProvider) SetCapabilities(caps provider.Capabilities) {
	p.capabilities = caps
}

// Validate validates the mock provider.
func (p *MockProvider) Validate(ctx context.Context) error {
	if p.OnValidate != nil {
		return p.OnValidate(ctx)
	}
	return nil
}

// Close closes the mock provider.
func (p *MockProvider) Close() error {
	if p.OnClose != nil {
		return p.OnClose()
	}
	return nil
}

// Instances returns all created instances (for testing).
func (p *MockProvider) Instances() []*MockInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*MockInstance, 0, len(p.instances))
	for _, inst := range p.instances {
		result = append(result, inst)
	}
	return result
}

var _ provider.Provider = (*MockProvider)(nil)
