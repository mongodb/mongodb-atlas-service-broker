package broker

import (
	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"go.uber.org/zap"
)

type MockAtlasClient struct {
	Clusters map[string]*atlas.Cluster
	Users    map[string]*atlas.User
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

func (m MockAtlasClient) GetCluster(name string) (*atlas.Cluster, error) {
	cluster := m.Clusters[name]
	if cluster == nil {
		return nil, atlas.ErrClusterNotFound
	}

	return cluster, nil
}

func (m MockAtlasClient) SetClusterState(name string, state string) {
	cluster := m.Clusters[name]
	if cluster == nil {
		return
	}

	cluster.State = state
}

func (m MockAtlasClient) CreateUser(user atlas.User) (*atlas.User, error) {
	if m.Users[user.Username] != nil {
		return nil, atlas.ErrUserAlreadyExists
	}

	m.Users[user.Username] = &user
	return &user, nil
}

func (m MockAtlasClient) DeleteUser(name string) error {
	if m.Users[name] == nil {
		return atlas.ErrUserNotFound
	}

	m.Users[name] = nil

	return nil
}

func setupTest() (*Broker, MockAtlasClient) {
	client := MockAtlasClient{
		Clusters: make(map[string]*atlas.Cluster),
		Users:    make(map[string]*atlas.User),
	}
	broker := NewBroker(client, zap.NewNop().Sugar())
	return broker, client
}
