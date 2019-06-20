package atlas

import "fmt"

var (
	ClusterStateIdle      = "IDLE"
	ClusterStateCreating  = "CREATING"
	ClusterStateUpdating  = "UPDATING"
	ClusterStateDeleting  = "DELETING"
	ClusterStateDeleted   = "DELETED"
	ClusterStateRepairing = "REPAIRING"
)

var (
	ClusterTypeReplicaSet = "REPLICASET"
	ClusterTypeSharded    = "SHARDED"
)

type ClustersResponse struct {
	Clusters []Cluster `json:"results"`
}

// Cluster represents a single cluster in Atlas
type Cluster struct {
	Name     string   `json:"name"`
	State    string   `json:"stateName,omitempty"`
	Type     string   `json:"clusterType,omitempty"`
	Provider Provider `json:"providerSettings"`
}

// Provider represents the provider setting for a cluster
type Provider struct {
	Name     string `json:"providerName"`
	Instance string `json:"instanceSizeName"`
	Region   string `json:"regionName"`
}

func (c *HTTPClient) CreateCluster(cluster Cluster) (*Cluster, error) {
	var resultingCluster Cluster
	err := c.request("POST", "clusters", cluster, &resultingCluster)
	return &resultingCluster, err
}

func (c *HTTPClient) TerminateCluster(name string) error {
	path := fmt.Sprintf("clusters/%s", name)
	return c.request("DELETE", path, nil, nil)
}
