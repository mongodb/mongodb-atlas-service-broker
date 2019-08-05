package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/10gen/atlas-service-broker/pkg/atlas"
	brokerlib "github.com/10gen/atlas-service-broker/pkg/broker"
	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var (
	broker *brokerlib.Broker
	client atlas.Client
)

func TestMain(m *testing.M) {
	baseURL := getEnvOrPanic("ATLAS_BASE_URL")
	groupID := getEnvOrPanic("ATLAS_GROUP_ID")
	publicKey := getEnvOrPanic("ATLAS_PUBLIC_KEY")
	privateKey := getEnvOrPanic("ATLAS_PRIVATE_KEY")
	client, _ = atlas.NewClient(baseURL, groupID, publicKey, privateKey)

	// Setup the broker which will be used
	broker = brokerlib.NewBroker(client, zap.NewNop().Sugar())

	// Run all tests in order. The tests will first provision a new instance,
	// create a new binding for the provsioned instance, delete the binding,
	// and finally deprovision the instance.
	result := m.Run()

	os.Exit(result)
}

func TestProvision(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	// Setting up our Expected cluster
	var expectedCluster = &atlas.Cluster{
		AutoScaling: atlas.AutoScalingConfig{
			DiskGBEnabled: true,
		},
		Name:          clusterName,
		BackupEnabled: true,
		BIConnector: atlas.BIConnectorConfig{
			Enabled:        true,
			ReadPreference: "primary",
		},
		Type:                     "REPLICASET",
		DiskSizeGB:               10,
		EncryptionAtRestProvider: "NONE",
		MongoDBMajorVersion:      "4.0",
		NumShards:                1,
		ProviderBackupEnabled:    false,
		ProviderSettings: &atlas.ProviderSettings{
			EncryptEBSVolume: true,
			Instance:         "M10",
			Name:             "AWS",
			Region:           "EU_WEST_1",
			VolumeType:       "STANDARD",
			DiskIOPS:         0,
		},
		ReplicationSpecs: []atlas.ReplicationSpec{
			atlas.ReplicationSpec{
				ID:        "5c87f79087d9d612a175f46c",
				NumShards: 1,
				RegionsConfig: map[string]atlas.RegionsConfig{
					"EU_WEST_1": atlas.RegionsConfig{
						ElectableNodes: 3,
						ReadOnlyNodes:  1,
						AnalyticsNodes: 1,
						Priority:       7,
					},
				},
				ZoneName: "Zone 1",
			},
		},
		State: "IDLE",
	}

	// Setting up the params for the body request
	params, error := json.Marshal(expectedCluster)
	assert.NoError(t, error)

	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "mongodb-aws",
		PlanID:        "AWS-M10",
		RawParameters: []byte(params),
	}, true)

	defer teardownInstance(instanceID)

	if !assert.NoError(t, err) {
		return
	}

	// Ensure the cluster is being created.
	cluster, err := client.GetCluster(clusterName)
	assert.NoError(t, err)
	assert.Equal(t, atlas.ClusterStateCreating, cluster.State)

	// Wait a maximum of 20 minutes for cluster to reach state idle.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationProvision, 20)
	if !assert.NoError(t, err) {
		return
	}

	// Request
	cluster, err = client.GetCluster(clusterName)
	assert.NoError(t, err)

	// Ensure response is equal to request cluster
	assert.Equal(t, expectedCluster, cluster)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()

	clusterName, err := setupInstance(instanceID)
	defer teardownInstance(instanceID)
	if !assert.NoError(t, err) {
		return
	}

	cluster, err := client.GetCluster(clusterName)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure cluster is in the correct starting state.
	// The instance size should be M10 and backups should be disabled.
	assert.Equal(t, "M10", cluster.ProviderSettings.Instance)
	assert.False(t, cluster.BackupEnabled)

	// Update the cluster plan (instance size) and enable backups.
	params := `{
		"cluster": {
			"backupEnabled": true
		}
	}`

	_, err = broker.Update(context.Background(), instanceID, brokerapi.UpdateDetails{
		ServiceID:     "mongodb-aws",
		PlanID:        "AWS-M20",
		RawParameters: []byte(params),
	}, true)

	if !assert.NoError(t, err) {
		return
	}

	// Wait a maximum of 25 minutes for cluster to finish updating.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationUpdate, 25)
	if !assert.NoError(t, err) {
		return
	}

	cluster, err = client.GetCluster(clusterName)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure instance size is now "M20" and backups are enabled.
	assert.Equal(t, atlas.ClusterStateIdle, cluster.State)
	assert.Equal(t, "M20", cluster.ProviderSettings.Instance)
	assert.True(t, cluster.BackupEnabled)
}

func TestBind(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	bindingID := uuid.New().String()

	clusterName, err := setupInstance(instanceID)
	defer teardownInstance(instanceID)
	if !assert.NoError(t, err) {
		return
	}

	spec, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)
	defer teardownBinding(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure user was created.
	_, err = client.GetUser(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	credentials, ok := spec.Credentials.(brokerlib.ConnectionDetails)
	if !assert.True(t, ok, "Expected credentials to have type broker.ConnectionDetails") {
		return
	}

	// Get the cluster to get its connection URI.
	cluster, err := client.GetCluster(clusterName)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure the MongoDB username is the binding ID, that the password is not
	// empty and that the connection URI matches the cluster's.
	assert.Equal(t, bindingID, credentials.Username)
	assert.NotEmpty(t, credentials.Password, "Expected non-empty password")
	assert.Equal(t, cluster.URI, credentials.URI)

	// Ensure the cluster can be connected to with the generated credentials.
	// We need to reset the auth source using a parameter otherwise the Go
	// MongoDB library will fail to parse the connection string.
	conn := options.Client().
		ApplyURI(credentials.URI + "/?authSource=").
		SetAuth(options.Credential{
			Username:    credentials.Username,
			Password:    credentials.Password,
			PasswordSet: true,
		})

	// Try connecting to the cluster to ensure that the credentials are
	// valid. There is sometimes a slight delay before the user is ready so this
	// will try to connect for up to a minute.
	err = poll(1, func() (bool, error) {
		client, err := mongo.NewClient(conn)
		if err != nil {
			return false, nil
		}

		err = client.Connect(context.Background())
		if err != nil {
			return false, nil
		}

		err = client.Ping(context.Background(), readpref.Primary())
		if err != nil {
			return false, nil
		}

		return true, nil
	})

	assert.NoError(t, err)
}

func TestUnbind(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	bindingID := uuid.New().String()

	_, err := setupInstance(instanceID)
	defer teardownInstance(instanceID)
	if !assert.NoError(t, err) {
		return
	}

	_, err = setupBinding(bindingID)
	defer teardownBinding(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	_, err = broker.Unbind(context.Background(), instanceID, bindingID, brokerapi.UnbindDetails{}, true)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure the user has been deleted and can't be found.
	_, err = client.GetUser(bindingID)
	assert.Error(t, err, "Expected user not found error")
}

func TestDeprovision(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()

	_, err := setupInstance(instanceID)
	defer teardownInstance(instanceID)
	if !assert.NoError(t, err) {
		return
	}

	// Deprovision the cluster.
	_, err = broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)
	if !assert.NoError(t, err) {
		return
	}

	err = waitForLastOperation(broker, instanceID, brokerlib.OperationDeprovision, 10)
	assert.NoError(t, err)

	_, err = client.GetCluster(brokerlib.NormalizeClusterName(instanceID))
	assert.Equal(t, atlas.ErrClusterNotFound, err)
}

// waitForLastOperation will poll the last operation function for a specified
// operation. The function returns once the operation was successful or the
// timeout has been reached.
func waitForLastOperation(broker *brokerlib.Broker, instanceID string, operation string, timeoutMinutes int) error {
	return poll(timeoutMinutes, func() (bool, error) {
		res, err := broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
			OperationData: operation,
		})

		if err != nil {
			return false, err
		}

		return res.State == brokerapi.Succeeded, nil
	})
}

// setupInstance will deploy a simple cluster to Atlas and wait for it to
// be created.
func setupInstance(instanceID string) (string, error) {
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	// Create a cluster running on AWS in eu-west-1. THe instance size should be
	// M10 and backup should be disabled.
	_, err := client.CreateCluster(atlas.Cluster{
		Name:          clusterName,
		BackupEnabled: false,
		ProviderSettings: &atlas.ProviderSettings{
			Name:     "AWS",
			Instance: "M10",
			Region:   "EU_WEST_1",
		},
	})
	if err != nil {
		return "", err
	}

	// Wait for cluster to reach state "idle".
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

func getEnvOrPanic(name string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		panic(fmt.Sprintf(`Could not find environment variable "%s"`, name))
	}

	return value
}
