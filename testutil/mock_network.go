package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/happyhackingspace/sindoq/internal/provider"
)

// MockNetwork is a configurable mock implementation of provider.Network.
type MockNetwork struct {
	ports map[int]*provider.PublishedPort
	mu    sync.RWMutex

	// Hooks for testing
	OnPublishPort   func(ctx context.Context, port int) (*provider.PublishedPort, error)
	OnGetPublicURL  func(port int) (string, error)
	OnListPorts     func(ctx context.Context) ([]*provider.PublishedPort, error)
	OnUnpublishPort func(ctx context.Context, port int) error
}

// NewMockNetwork creates a new mock network.
func NewMockNetwork() *MockNetwork {
	return &MockNetwork{
		ports: make(map[int]*provider.PublishedPort),
	}
}

// PublishPort exposes a port publicly.
func (n *MockNetwork) PublishPort(ctx context.Context, port int) (*provider.PublishedPort, error) {
	if n.OnPublishPort != nil {
		return n.OnPublishPort(ctx, port)
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	published := &provider.PublishedPort{
		LocalPort:  port,
		PublicPort: port,
		Protocol:   "tcp",
		PublicURL:  fmt.Sprintf("http://localhost:%d", port),
	}
	n.ports[port] = published
	return published, nil
}

// GetPublicURL returns the public URL for an exposed port.
func (n *MockNetwork) GetPublicURL(port int) (string, error) {
	if n.OnGetPublicURL != nil {
		return n.OnGetPublicURL(port)
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	published, ok := n.ports[port]
	if !ok {
		return "", fmt.Errorf("port %d not published", port)
	}
	return published.PublicURL, nil
}

// ListPorts returns all published ports.
func (n *MockNetwork) ListPorts(ctx context.Context) ([]*provider.PublishedPort, error) {
	if n.OnListPorts != nil {
		return n.OnListPorts(ctx)
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	result := make([]*provider.PublishedPort, 0, len(n.ports))
	for _, p := range n.ports {
		result = append(result, p)
	}
	return result, nil
}

// UnpublishPort removes port exposure.
func (n *MockNetwork) UnpublishPort(ctx context.Context, port int) error {
	if n.OnUnpublishPort != nil {
		return n.OnUnpublishPort(ctx, port)
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.ports, port)
	return nil
}

// SetPort adds a port mapping (for test setup).
func (n *MockNetwork) SetPort(localPort int, publicURL string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ports[localPort] = &provider.PublishedPort{
		LocalPort:  localPort,
		PublicPort: localPort,
		Protocol:   "tcp",
		PublicURL:  publicURL,
	}
}

// Clear removes all port mappings.
func (n *MockNetwork) Clear() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ports = make(map[int]*provider.PublishedPort)
}

var _ provider.Network = (*MockNetwork)(nil)
