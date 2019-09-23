package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
	brokerlib "github.com/mongodb/mongodb-atlas-service-broker/pkg/broker"
	testutil "github.com/mongodb/mongodb-atlas-service-broker/test/util"
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
	ctx    context.Context
)

func TestMain(m *testing.M) {
	baseURL := testutil.GetEnvOrPanic("ATLAS_BASE_URL")
	groupID := testutil.GetEnvOrPanic("ATLAS_GROUP_ID")
	publicKey := testutil.GetEnvOrPanic("ATLAS_PUBLIC_KEY")
	privateKey := testutil.GetEnvOrPanic("ATLAS_PRIVATE_KEY")

	client = atlas.NewClient(baseURL, groupID, publicKey, privateKey)
	ctx = context.WithValue(ctx, brokerlib.ContextKeyAtlasClient, client)

	whitelist := brokerlib.Whitelist{
		"AWS":    []string{"M10", "M20"},
		"GCP":    []string{"M10"},
		"TENANT": []string{"M2", "M5"},
	}

	// Setup the broker which will be used
	broker = brokerlib.NewBrokerWithWhitelist(zap.NewNop().Sugar(), whitelist)

	result := m.Run()

	os.Exit(result)
}

func TestCatalog(t *testing.T) {
	t.Parallel()

	services, err := broker.Services(ctx)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, services)
	for _, service := range services {
		assert.NotEmptyf(t, service.Plans, `Expected service "%s" to have plans`, service.Name)
	}
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
			Enabled: false,
		},
		ClusterType:              "REPLICASET",
		DiskSizeGB:               10,
		EncryptionAtRestProvider: "NONE",
		MongoDBMajorVersion:      "4.0",
		NumShards:                1,
		ProviderBackupEnabled:    false,
		ProviderSettings: &atlas.ProviderSettings{
			EncryptEBSVolume: true,
			InstanceSizeName: "M10",
			ProviderName:     "AWS",
			RegionName:       "EU_WEST_1",
			VolumeType:       "STANDARD",
			DiskIOPS:         100,
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
	}

	// Setting up the params for the body request
	paramsByte, marshalErr := json.Marshal(expectedCluster)
	assert.NoError(t, marshalErr)

	params := `{"cluster":` + string(paramsByte) + `}`

	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "aosb-cluster-service-aws",
		PlanID:        "aosb-cluster-plan-aws-m10",
		RawParameters: []byte(params),
	}, true)

	defer teardownInstance(instanceID)

	if !assert.NoError(t, err) {
		return
	}

	// Ensure the cluster is being created.
	cluster, err := client.GetCluster(clusterName)
	assert.NoError(t, err)
	assert.Equal(t, atlas.ClusterStateCreating, cluster.StateName)

	// Wait a maximum of 20 minutes for cluster to reach state idle.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationProvision, 20)
	if !assert.NoError(t, err) {
		return
	}

	cluster, err = client.GetCluster(clusterName)
	assert.NoError(t, err)

	// Altering these parameters due to the fact that, they can't be configured from up front
	cluster.SrvAddress = ""
	expectedCluster.StateName = "IDLE"
	expectedCluster.BIConnector.ReadPreference = "secondary"

	// Ensure response is equal to request cluster
	assert.Equal(t, expectedCluster, cluster)
}

func TestProvisionProvidersConfig(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	// Setting up our Expected cluster
	params := `{
		"cluster": {
			"providerSettings": {
				"regionName": "EU_WEST_1"
			}
		}
	}`

	// We try to provision something that the adminstrator didn't create
	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "aosb-cluster-service-azure",
		PlanID:        "aosb-cluster-plan-azure-m10",
		RawParameters: []byte(params),
	}, true)
	assert.Error(t, atlas.ErrPlanIDNotFound, err)

	// One the administrator did create
	// Setting up our Expected cluster
	params = `{
			"cluster": {
				"providerSettings": {
					"regionName": "EUROPE_WEST_2"
				}
			}
	}`
	_, err = broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "aosb-cluster-service-gcp",
		PlanID:        "aosb-cluster-plan-gcp-m10",
		RawParameters: []byte(params),
	}, true)

	defer teardownInstance(instanceID)

	if !assert.NoError(t, err) {
		return
	}

	// Ensure the cluster is being created.
	cluster, err := client.GetCluster(clusterName)
	assert.NoError(t, err)
	assert.Equal(t, atlas.ClusterStateCreating, cluster.StateName)

	// Wait a maximum of 20 minutes for cluster to reach state idle.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationProvision, 20)
	if !assert.NoError(t, err) {
		return
	}

	_, err = client.GetCluster(clusterName)
	assert.NoError(t, err)
}

func TestProvisionM2Size(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	// Setting up our Expected cluster
	var expectedCluster = &atlas.Cluster{
		Name: clusterName,
		ProviderSettings: &atlas.ProviderSettings{
			BackingProviderName: "AWS",
			InstanceSizeName:    "M2",
			ProviderName:        "TENANT",
			RegionName:          "US_EAST_1",
		},
	}

	// Setting up the params for the body request
	paramsByte, marshalErr := json.Marshal(expectedCluster)
	assert.NoError(t, marshalErr)

	params := `{"cluster":` + string(paramsByte) + `}`

	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "aosb-cluster-service-tenant",
		PlanID:        "aosb-cluster-plan-tenant-m2",
		RawParameters: []byte(params),
	}, true)

	defer teardownInstance(instanceID)

	if !assert.NoError(t, err) {
		return
	}

	// Ensure the cluster is being created.
	cluster, err := client.GetCluster(clusterName)
	assert.NoError(t, err)
	assert.Equal(t, atlas.ClusterStateCreating, cluster.StateName)

	// Wait a maximum of 20 minutes for cluster to reach state idle.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationProvision, 20)
	if !assert.NoError(t, err) {
		return
	}

	cluster, err = client.GetCluster(clusterName)
	assert.NoError(t, err)

	// Ensure response is equal to request cluster
	assert.Equal(t, expectedCluster.ProviderSettings.BackingProviderName, cluster.ProviderSettings.BackingProviderName)
	assert.Equal(t, expectedCluster.ProviderSettings.ProviderName, cluster.ProviderSettings.ProviderName)
	assert.Equal(t, expectedCluster.ProviderSettings.InstanceSizeName, cluster.ProviderSettings.InstanceSizeName)
	assert.Equal(t, expectedCluster.ProviderSettings.RegionName, cluster.ProviderSettings.RegionName)
}

func TestProvisionM5Size(t *testing.T) {
	t.Parallel()

	instanceID := uuid.New().String()
	clusterName := brokerlib.NormalizeClusterName(instanceID)

	// Setting up our Expected cluster
	var expectedCluster = &atlas.Cluster{
		Name: clusterName,
		ProviderSettings: &atlas.ProviderSettings{
			BackingProviderName: "AWS",
			InstanceSizeName:    "M5",
			ProviderName:        "TENANT",
			RegionName:          "US_EAST_1",
		},
	}

	// Setting up the params for the body request
	paramsByte, marshalErr := json.Marshal(expectedCluster)
	assert.NoError(t, marshalErr)

	params := `{"cluster":` + string(paramsByte) + `}`

	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     "aosb-cluster-service-tenant",
		PlanID:        "aosb-cluster-plan-tenant-m5",
		RawParameters: []byte(params),
	}, true)

	defer teardownInstance(instanceID)

	if !assert.NoError(t, err) {
		return
	}

	// Ensure the cluster is being created.
	cluster, err := client.GetCluster(clusterName)
	assert.NoError(t, err)
	assert.Equal(t, atlas.ClusterStateCreating, cluster.StateName)

	// Wait a maximum of 20 minutes for cluster to reach state idle.
	err = waitForLastOperation(broker, instanceID, brokerlib.OperationProvision, 20)
	if !assert.NoError(t, err) {
		return
	}

	cluster, err = client.GetCluster(clusterName)
	assert.NoError(t, err)

	// Ensure response is equal to request cluster
	assert.Equal(t, expectedCluster.ProviderSettings.BackingProviderName, cluster.ProviderSettings.BackingProviderName)
	assert.Equal(t, expectedCluster.ProviderSettings.ProviderName, cluster.ProviderSettings.ProviderName)
	assert.Equal(t, expectedCluster.ProviderSettings.InstanceSizeName, cluster.ProviderSettings.InstanceSizeName)
	assert.Equal(t, expectedCluster.ProviderSettings.RegionName, cluster.ProviderSettings.RegionName)
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
	assert.Equal(t, "M10", cluster.ProviderSettings.InstanceSizeName)
	assert.False(t, cluster.BackupEnabled)

	// Update the cluster plan (instance size) and enable backups.
	params := `{
		"cluster": {
			"backupEnabled": true
		}
	}`

	// Try to update to a plan that doesn't exist
	_, err = broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		ServiceID:     "aosb-cluster-service-aws",
		PlanID:        "aosb-cluster-plan-aws-m60",
		RawParameters: []byte(params),
	}, true)

	assert.Error(t, atlas.ErrPlanIDNotFound, err)

	_, err = broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		ServiceID:     "aosb-cluster-service-aws",
		PlanID:        "aosb-cluster-plan-aws-m20",
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
	assert.Equal(t, atlas.ClusterStateIdle, cluster.StateName)
	assert.Equal(t, "M20", cluster.ProviderSettings.InstanceSizeName)
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

	params := `{
		"user": {
			"ldapAuthType": "NONE",
			"roles": [{
				"roleName": "read",
				"databaseName": "database",
				"collectionName": "collection"
			}]
		}}`

	spec, err := broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		ServiceID:     "aosb-cluster-service-aws",
		PlanID:        "aosb-cluster-plan-aws-m10",
		RawParameters: []byte(params),
	}, true)
	defer teardownBinding(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	// Ensure user was created and all parameters made it through.
	user, err := client.GetUser(bindingID)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bindingID, user.Username)
	assert.Equal(t, "NONE", user.LDAPAuthType)

	expectedRoles := []atlas.Role{
		atlas.Role{
			Name:           "read",
			DatabaseName:   "database",
			CollectionName: "collection",
		},
	}
	assert.Equal(t, expectedRoles, user.Roles)

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
	assert.Equal(t, cluster.SrvAddress, credentials.URI)

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
	// will try to connect for up to 10 minutes.
	err = testutil.Poll(10, func() (bool, error) {
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

	_, err = broker.Unbind(ctx, instanceID, bindingID, brokerapi.UnbindDetails{}, true)
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
	_, err = broker.Deprovision(ctx, instanceID, brokerapi.DeprovisionDetails{}, true)
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
	return testutil.Poll(timeoutMinutes, func() (bool, error) {
		res, err := broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
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
			ProviderName:     "AWS",
			InstanceSizeName: "M10",
			RegionName:       "EU_WEST_1",
		},
	})
	if err != nil {
		return "", err
	}

	// Wait for cluster to reach state "idle".
	err = testutil.Poll(15, func() (bool, error) {
		cluster, err := client.GetCluster(clusterName)
		if err != nil {
			return false, err
		}

		if cluster.StateName == atlas.ClusterStateIdle {
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
		Roles: []atlas.Role{
			atlas.Role{
				Name:         "readWriteAnyDatabase",
				DatabaseName: "admin",
			},
		},
	})
}

func teardownInstance(instanceID string) {
	client.DeleteCluster(brokerlib.NormalizeClusterName(instanceID))
}

func teardownBinding(bindingID string) {
	client.DeleteUser(bindingID)
}
