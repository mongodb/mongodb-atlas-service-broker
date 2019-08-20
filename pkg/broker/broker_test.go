package broker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testServiceID = "aosb-cluster-service-aws"
	testPlanID    = "aosb-cluster-plan-aws-m10"
)

type MockAtlasClient struct {
	Clusters map[string]*atlas.Cluster
	Users    map[string]*atlas.User
}

func (m MockAtlasClient) CreateCluster(cluster atlas.Cluster) (*atlas.Cluster, error) {
	if m.Clusters[cluster.Name] != nil {
		return nil, atlas.ErrClusterAlreadyExists
	}

	cluster.StateName = atlas.ClusterStateCreating

	m.Clusters[cluster.Name] = &cluster

	return &cluster, nil
}

func (m MockAtlasClient) UpdateCluster(cluster atlas.Cluster) (*atlas.Cluster, error) {
	if m.Clusters[cluster.Name] == nil {
		return nil, atlas.ErrClusterNotFound
	}

	m.Clusters[cluster.Name] = &cluster

	return &cluster, nil
}

func (m MockAtlasClient) DeleteCluster(name string) error {
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

	cluster.StateName = state
}

func (m MockAtlasClient) CreateUser(user atlas.User) (*atlas.User, error) {
	if m.Users[user.Username] != nil {
		return nil, atlas.ErrUserAlreadyExists
	}

	m.Users[user.Username] = &user
	return &user, nil
}

func (m MockAtlasClient) GetUser(name string) (*atlas.User, error) {
	user := m.Users[name]
	if user == nil {
		return nil, atlas.ErrUserNotFound
	}

	return user, nil
}

func (m MockAtlasClient) DeleteUser(name string) error {
	if m.Users[name] == nil {
		return atlas.ErrUserNotFound
	}

	m.Users[name] = nil

	return nil
}

func (m MockAtlasClient) GetProvider(name string) (*atlas.Provider, error) {
	return &atlas.Provider{
		Name: "AWS",
		InstanceSizes: map[string]atlas.InstanceSize{
			"M10": atlas.InstanceSize{
				Name: "M10",
			},
			"M20": atlas.InstanceSize{
				Name: "M20",
			},
		},
	}, nil
}

func (m MockAtlasClient) GetDashboardURL(clusterName string) string {
	return "http://dashboard"
}

func setupTest() (*Broker, MockAtlasClient, context.Context) {
	client := MockAtlasClient{
		Clusters: make(map[string]*atlas.Cluster),
		Users:    make(map[string]*atlas.User),
	}
	ctx := context.WithValue(context.Background(), ContextKeyAtlasClient, client)

	broker := NewBroker(zap.NewNop().Sugar())
	return broker, client, ctx
}

func TestAuthMiddleware(t *testing.T) {
	baseURL := "http://baseURL"
	groupID := "group-id"
	publicKey := "public-key"
	privateKey := "private-key"

	middleware := AuthMiddleware(baseURL)

	// On successful auth the middleware will run testHandler which ensures
	// the context was set up correctly.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client, ok := r.Context().Value(ContextKeyAtlasClient).(*atlas.HTTPClient)
		if !assert.True(t, ok, "expected context to have client") {
			return
		}

		assert.Equal(t, baseURL, client.BaseURL)
		assert.Equal(t, groupID, client.GroupID)
		assert.Equal(t, publicKey, client.PublicKey)
		assert.Equal(t, privateKey, client.PrivateKey)
	})

	// Fake HTTP request which will be sent to middleware. Response is captured
	// by a recorder.
	req, err := http.NewRequest("GET", "http://test", nil)
	if !assert.NoError(t, err) {
		return
	}
	w := httptest.NewRecorder()

	// Missing basic auth credentials.
	middleware(testHandler).ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

	// Empty basic auth credentials.
	req.SetBasicAuth("", "")
	middleware(testHandler).ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

	// Incorrect username format.
	req.SetBasicAuth("incorrect-username", "password")
	middleware(testHandler).ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

	// Valid credentials. testHandler will run and validate the context.
	req.SetBasicAuth(publicKey+"@"+groupID, privateKey)
	middleware(testHandler).ServeHTTP(w, req)
}
