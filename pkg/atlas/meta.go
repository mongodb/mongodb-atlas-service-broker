package atlas

import (
	"fmt"
	"net/http"
)

// Provider represents a single cloud provider to which a cluster can be
// deployed.
type Provider struct {
	Name          string `json:"@provider"`
	InstanceSizes map[string]InstanceSize
}

// InstanceSize represents an available cluster size.
type InstanceSize struct {
	Name string `json:"name"`
}

// GetProvider will find a provider by name using the private API.
// GET /cloudProviders/{NAME}/options
func (c *HTTPClient) GetProvider(name string) (*Provider, error) {
	path := fmt.Sprintf("cloudProviders/%s/options", name)
	var provider Provider

	err := c.requestPrivate(http.MethodGet, path, nil, &provider)
	return &provider, err
}
