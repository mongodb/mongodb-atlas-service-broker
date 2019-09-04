package atlas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TENANT is the provider that has plans sizes M2 and M5
const TENANT = "TENANT"

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
	var provider Provider

	// First check to see if any of the providers have been set by the administrator of the cluster locally
	provider, isSet, err := getLocalProvider(name, c.ProvidersConfig)
	if err != nil {
		return &provider, err
	}

	// Provider exists in local file
	if isSet {
		return &provider, err
	} else if name == TENANT { // Provider TENANT isn't set locally, but the user requires it. Return the harcoded version of it
		return &Provider{
			Name: TENANT,
			InstanceSizes: map[string]InstanceSize{
				"M2": {Name: "M2"},
				"M5": {Name: "M5"},
			},
		}, err
	}

	// Otherwise we make a request to the private atlas API when it's not set
	path := fmt.Sprintf("cloudProviders/%s/options", name)
	err = c.requestPrivate(http.MethodGet, path, nil, &provider)
	return &provider, err
}

// getLocalProvider will setup all of the available providors together with their plans/sizes.
// Users may want to set their local provider, and combine that with a remote provider that isn't set locally.
func getLocalProvider(name string, pathToFile string) (provider Provider, isSet bool, err error) {
	// Load in json file
	file, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return
	}

	var arr = []Provider{}
	err = json.Unmarshal([]byte(file), &arr)
	if err != nil {
		return
	}

	for _, document := range arr {
		// Check to see if document is present in config file
		if document.Name == name {
			return document, true, err
		}
	}

	// None SET
	return Provider{}, false, err
}
