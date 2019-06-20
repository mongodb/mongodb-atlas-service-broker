package broker

import (
	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"go.uber.org/zap"
)

type MockAtlasClient struct {
	Clusters map[string]*atlas.Cluster
}

func (m MockAtlasClient) CreateCluster(cluster atlas.Cluster) (*atlas.Cluster, error) {
	if m.Clusters[cluster.Name] != nil {
		return nil, atlas.ErrClusterAlreadyExists
	}

	cluster.State = atlas.ClusterStateCreating
	cluster.Type = atlas.ClusterTypeReplicaSet

	m.Clusters[cluster.Name] = &cluster

	return &cluster, nil
}

func (m MockAtlasClient) TerminateCluster(name string) error {
	if m.Clusters[name] == nil {
		return atlas.ErrClusterNotFound
	}

	m.Clusters[name] = nil

	return nil
}

func SetupTest() (*Broker, MockAtlasClient) {
	client := MockAtlasClient{
		Clusters: make(map[string]*atlas.Cluster),
	}
	broker := NewBroker(client, zap.NewNop().Sugar())
	return broker, client
}
