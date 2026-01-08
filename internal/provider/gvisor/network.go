//go:build linux

package gvisor

import (
	"context"
	"fmt"

	"github.com/happyhackingspace/sindoq/internal/provider"
)

// gvisorNetwork implements provider.Network for gVisor containers.
type gvisorNetwork struct {
	instance *Instance
	ports    map[int]*provider.PublishedPort
}

// PublishPort exposes a port publicly.
func (n *gvisorNetwork) PublishPort(ctx context.Context, port int) (*provider.PublishedPort, error) {
	// Like Docker, gVisor requires port mappings at container creation time
	return nil, fmt.Errorf("gVisor provider requires ports to be configured at sandbox creation time. " +
		"Use CreateOptions.Metadata[\"ports\"] to specify port mappings")
}

// GetPublicURL returns the public URL for an exposed port.
func (n *gvisorNetwork) GetPublicURL(port int) (string, error) {
	if n.ports == nil {
		return "", fmt.Errorf("no ports published")
	}

	p, ok := n.ports[port]
	if !ok {
		return "", fmt.Errorf("port %d not published", port)
	}

	return p.PublicURL, nil
}

// ListPorts returns all published ports.
func (n *gvisorNetwork) ListPorts(ctx context.Context) ([]*provider.PublishedPort, error) {
	ports := make([]*provider.PublishedPort, 0)

	if n.ports != nil {
		for _, p := range n.ports {
			ports = append(ports, p)
		}
	}

	return ports, nil
}

// UnpublishPort removes port exposure.
func (n *gvisorNetwork) UnpublishPort(ctx context.Context, port int) error {
	return fmt.Errorf("gVisor provider does not support dynamic port unpublishing")
}

var _ provider.Network = (*gvisorNetwork)(nil)
