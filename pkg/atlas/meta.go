package atlas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// PathToProviderJSON is the path to the json file
var PathToProviderJSON = "providers.json"

// ProvidersFile represents all of the providers that have been set in the provider.json file
type ProvidersFile struct {
	Providers []map[string]Provider `json:"providers,omitempty"`
}

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

	// First check to see if any of the providers have been set by the administrator of the cluster locally
	provider, isSet, err := setupProviders(name)
	if err != nil {
		return &provider, err
	}
	if isSet {
		return &provider, err
	}

	// Otherwise we make a request to the private atlas API when it's not set
	err = c.requestPrivate(http.MethodGet, path, nil, &provider)
	return &provider, err
}

// setupProviders will setup all of the available providors together with their plans/sizes.
// Users may want to set their local providor, and combine that with a remote provider that isn't set locally.
func setupProviders(name string) (provider Provider, isSet bool, err error) {
	// Load in json file
	file, err := ioutil.ReadFile(PathToProviderJSON)
	if err != nil {
		return
	}

	providersFile := &ProvidersFile{}
	err = json.Unmarshal(file, providersFile)
	if err != nil {
		return
	}

	for _, document := range providersFile.Providers {
		// each item (provider) has a name and instancesizes
		if item, ok := document["SET"]; ok && item.Name == name {
			return item, true, err
		}
	}

	// None SET
	return Provider{}, false, err
}
