package atlas

import (
	"fmt"
	"net/http"
)

// All states a cluster can be in.
var (
	ClusterStateIdle      = "IDLE"
	ClusterStateCreating  = "CREATING"
	ClusterStateUpdating  = "UPDATING"
	ClusterStateDeleting  = "DELETING"
	ClusterStateDeleted   = "DELETED"
	ClusterStateRepairing = "REPAIRING"
)

// The different types of clusters available in Atlas.
var (
	ClusterTypeReplicaSet = "REPLICASET"
	ClusterTypeSharded    = "SHARDED"
)

// Cluster represents a single cluster in Atlas.
type Cluster struct {
	Name     string   `json:"name"`
	State    string   `json:"stateName,omitempty"`
	Type     string   `json:"clusterType,omitempty"`
	URI      string   `json:"srvAddress,omitempty"`
	Provider Provider `json:"providerSettings"`
}

// Provider represents the provider setting for a cluster.
type Provider struct {
	Name     string `json:"providerName"`
	Instance string `json:"instanceSizeName"`
	Region   string `json:"regionName"`
}

// CreateCluster will create a new cluster asynchronously.
// POST /clusters
func (c *HTTPClient) CreateCluster(cluster Cluster) (*Cluster, error) {
	var resultingCluster Cluster
	err := c.request(http.MethodPost, "clusters", cluster, &resultingCluster)
	return &resultingCluster, err
}

// TerminateCluster will terminate a cluster asynchronously.
// DELETE /clusters/{CLUSTER-NAME}
func (c *HTTPClient) TerminateCluster(name string) error {
	path := fmt.Sprintf("clusters/%s", name)
	return c.request(http.MethodDelete, path, nil, nil)
}

// GetCluster will find a cluster by name.
// GET /clusters/{CLUSTER-NAME}
func (c *HTTPClient) GetCluster(name string) (*Cluster, error) {
	path := fmt.Sprintf("clusters/%s", name)

	var cluster Cluster
	err := c.request(http.MethodGet, path, nil, &cluster)
	return &cluster, err
}
