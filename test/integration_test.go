package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/10gen/atlas-service-broker/pkg/atlas"
	brokerlib "github.com/10gen/atlas-service-broker/pkg/broker"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	broker *brokerlib.Broker
	client atlas.Client
)

func TestMain(m *testing.M) {
	baseURL := os.Getenv("ATLAS_BASE_URL")
	groupID := os.Getenv("ATLAS_GROUP_ID")
	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")

	client, _ = atlas.NewClient(baseURL, groupID, publicKey, privateKey)

	// Setup the broker which will be used
	broker = brokerlib.NewBroker(client, zap.NewNop().Sugar())

	// Run all tests in order. The tests will first provision a new instance,
	// create a new binding for the provsioned instance, delete the binding,
	// and finally deprovision the instance.
	result := m.Run()

	os.Exit(result)
}

// TestHelpers will test the setupInstance function to ensure a cluster is created
// as well as the setupBinding function to ensure a database user is created.
// This test gives an examples of what the other integration test will roughly
// look like. They'll use the helper functions to set up a fresh environment
// and use the Atlas client to verify the results.
func TestHelpers(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	bindingID := uuid.New().String()
	defer teardownInstance(instanceID)
	defer teardownBinding(bindingID)

	// Set up an instance and wait for it to be ready.
	clusterName, err := setupInstance(instanceID)
	if !assert.NoError(t, err) {
		return
	}

	// Fetch the newly created cluster.
	cluster, err := client.GetCluster(clusterName)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure cluster is in idle state.
	assert.Equal(t, atlas.ClusterStateIdle, cluster.State)

	// Set up a new database user.
	_, err = setupBinding(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure database user was created.
	_, err = client.GetUser(bindingID)
	assert.NoError(t, err)
}

// setupInstance will deploy a simple cluster to Atlas and wait for it to
// be created.
func setupInstance(instanceID string) (string, error) {
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	_, err := client.CreateCluster(atlas.Cluster{
		Name: clusterName,
		ProviderSettings: &atlas.ProviderSettings{
			Name:     "AWS",
			Instance: "M10",
			Region:   "EU_WEST_1",
		},
	})
	if err != nil {
		return "", err
	}

	err = poll(15, func() (bool, error) {
		cluster, err := client.GetCluster(clusterName)
		if err != nil {
			return false, err
		}

		if cluster.State == atlas.ClusterStateIdle {
			return true, nil
		}

		return false, nil
	})

	return clusterName, err
}

// setupBinding will create a new user with the binding ID as its username and
// a random password.
func setupBinding(bindingID string) (*atlas.User, error) {
	return client.CreateUser(atlas.User{
		Username: bindingID,
		Password: uuid.New().String(),
	})
}

func teardownInstance(instanceID string) {
	client.DeleteCluster(brokerlib.NormalizeClusterName(instanceID))
}

func teardownBinding(bindingID string) {
	client.DeleteUser(bindingID)
}

// poll will run f every 10 seconds until it returns true or the timout is
// reached.
func poll(timeoutMinutes int, f func() (bool, error)) error {
	pollInterval := 10

	for i := 0; i < timeoutMinutes*60; i++ {
		res, err := f()
		if err != nil {
			return err
		}

		if res {
			return nil
		}

		i += pollInterval
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

	return fmt.Errorf("timeout while polling (waited %d minutes)", timeoutMinutes)
}
