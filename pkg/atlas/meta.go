package atlas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

	// First check to see if any of the providers have been set by the administrator of the cluster
	provider, isSet, err := setupProviders(name)
	if isSet {
		return &provider, err
	}

	// Otherwise we make a request to the private atlas API
	err = c.requestPrivate(http.MethodGet, path, nil, &provider)
	return &provider, err
}

// setupProvidors will setup all of the available providors together with their plans/sizes.
// Users may want to set their local providor, and combine that with a remote provider.
func setupProviders(name string) (provider Provider, isSet bool, err error) {
	file, err := ioutil.ReadFile("/Users/victor/atlas-service-broker/pkg/atlas/providers.json")
	if err != nil {
		return
	}

	providersFile := &ProvidersFile{}
	err = json.Unmarshal(file, providersFile)
	if err != nil {
		return
	}
	fmt.Println("INTERFACE: ", providersFile)

	for _, document := range providersFile.Providers {
		if val, ok := document["SET"]; ok {
			fmt.Println("VAL", val)
			return val, true, err
		}
	}

	return Provider{}, false, err
}

// fmt.Println("FILE: ", string(file))
// fmt.Println("INTERFACE: ", providersFile)
// if json file is SET
// just call err := c.requestPrivate(http.MethodGet, path, nil, &provider)
// else load in the json specified by the administrators
// _, err := setupProviders()
// if err != nil {
// 	return nil, err
// }

// b, err := json.Marshal(provider)
// if err != nil {
// 	fmt.Println(err)
// 	return nil, err
// }
// fmt.Println("PROVIDOR: ", string(b))
