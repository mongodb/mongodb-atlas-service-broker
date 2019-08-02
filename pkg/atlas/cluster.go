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
	Name string `json:"name"`

	AutoScaling              AutoScalingConfig `json:"autoScaling,omitempty"`
	BackupEnabled            bool              `json:"backupEnabled,omitempty"`
	BIConnector              BIConnectorConfig `json:"biConnector,omitempty"`
	Type                     string            `json:"clusterType,omitempty"`
	DiskSizeGB               float64           `json:"diskSizeGB,omitempty"`
	EncryptionAtRestProvider string            `json:"encryptionAtRestProvider,omitempty"`
	MongoDBMajorVersion      string            `json:"mongoDBMajorVersion,omitempty"`
	NumShards                uint              `json:"numShards,omitempty"`
	ProviderBackupEnabled    bool              `json:"providerBackupEnabled,omitempty"`
	ReplicationSpecs         []ReplicationSpec `json:"replicationSpecs,omitempty"`
	ProviderSettings         *ProviderSettings `json:"providerSettings"`

	// Read-only attributes
	State string `json:"stateName,omitempty"`
	URI   string `json:"srvAddress,omitempty"`
}

// AutoScalingConfig represents the autoscaling settings for a cluster.
type AutoScalingConfig struct {
	DiskGBEnabled bool `json:"diskGBEnabled,omitempty"`
}

// BIConnectorConfig represents the BI connector settings for a cluster.
type BIConnectorConfig struct {
	Enabled        bool   `json:"enabled,omitempty"`
	ReadPreference string `json:"readPreference,omitempty"`
}

// ProviderSettings represents the provider setting for a cluster.
type ProviderSettings struct {
	Name            string `json:"providerName"`
	Instance        string `json:"instanceSizeName"`
	Region          string `json:"regionName,omitempty"`
	BackingProvider string `json:"backingProviderName,omitempty"`

	DiskIOPS         uint   `json:"diskIOPS,omitempty"`
	DiskType         string `json:"diskTypeName,omitempty"`
	EncryptEBSVolume bool   `json:"encryptEBSVolume,omitempty"`
	VolumeType       string `json:"volumeType,omitempty"`
}

// ReplicationSpec represents the replication settings for a single region.
type ReplicationSpec struct {
	// Unique identifier for a zone's replication document. Required for existing
	// zones and optional if adding new zones to a Global Cluster.
	ID            string                   `json:"id,omitempty"`
	NumShards     uint                     `json:"numShards,omitempty"`
	RegionsConfig map[string]RegionsConfig `json:"regionsConfig,omitempty"`
	ZoneName      string                   `json:"zoneName,omitempty"`
}

// RegionsConfig represents a region's config in a replication spec.
type RegionsConfig struct {
	ElectableNodes int `json:"electableNodes"`
	ReadOnlyNodes  int `json:"readOnlyNodes"`
	AnalyticsNodes int `json:"analyticsNodes,omitempty"`
	Priority       int `json:"priority,omitempty"`
}

// CreateCluster will create a new cluster asynchronously.
// POST /clusters
func (c *HTTPClient) CreateCluster(cluster Cluster) (*Cluster, string, error) {
	var resultingCluster Cluster
	var dashboardURL string

	err := c.request(http.MethodPost, "clusters", cluster, &resultingCluster)
	if err == nil {
		dashboardURL = c.GetDashboardURL()
	}
	return &resultingCluster, dashboardURL, err
}

// UpdateCluster will update a cluster asynchronously.
// PATCH /clusters/{CLUSTER-NAME}
func (c *HTTPClient) UpdateCluster(cluster Cluster) (*Cluster, string, error) {
	path := fmt.Sprintf("clusters/%s", cluster.Name)
	var resultingCluster Cluster
	var dashboardURL string

	err := c.request(http.MethodPatch, path, cluster, &resultingCluster)
	if err == nil {
		dashboardURL = c.GetDashboardURL()
	}
	return &resultingCluster, dashboardURL, err
}

// DeleteCluster will terminate a cluster asynchronously.
// DELETE /clusters/{CLUSTER-NAME}
func (c *HTTPClient) DeleteCluster(name string) error {
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
