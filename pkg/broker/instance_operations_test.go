package broker

import (
	"testing"

	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/stretchr/testify/assert"
)

// TestMissingAsync will make sure all async operations don't accept non-async
// clients.
func TestMissingAsync(t *testing.T) {
	broker, client, ctx := setupTest()

	// Try provisioning an instance without async support
	instanceID := "instance"
	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
	assert.Len(t, client.Clusters, 0, "Expected no clusters to be created")

	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Try updating existing cluster without async support
	_, err = broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())

	_, err = broker.Deprovision(ctx, instanceID, brokerapi.DeprovisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
}

func TestProvision(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	res, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationProvision, res.OperationData)
	assert.Len(t, client.Clusters, 1)
	assert.NotEmpty(t, res.DashboardURL)

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)
	assert.Equal(t, &atlas.ProviderSettings{
		ProviderName:     "AWS",
		InstanceSizeName: "M10",
	}, cluster.ProviderSettings)
}

func TestProvisionParams(t *testing.T) {
	broker, client, ctx := setupTest()

	params := `{
	"cluster": {
		"autoScaling": {
			"diskGBEnabled": true
		},
		"backupEnabled": true,
		"biConnector": {
			"enabled": true,
			"readPreference": "primary"
		},
		"clusterType": "SHARDED",
		"diskSizeGB": 100.0,
		"encryptionAtRestProvider": "NONE",
		"mongoDBMajorVersion": "4.0",
		"numShards": 2,
		"providerBackupEnabled": true,
		"providerSettings": {
			"diskIOPS": 10,
			"diskTypeName": "P4",
			"encryptEBSVolume": true,
			"regionName": "EU_CENTRAL_1",
			"volumeType": "STANDARD"
		},
		"replicationSpecs": [
			{
				"id": "ID",
				"numShards": 2,
				"regionsConfig": {
					"REGION": {
						"electableNodes": 1,
						"readOnlyNodes": 1,
						"analyticsNodes": 1,
						"priority": 1
					}
				},
				"zoneName": "ZONE"
			}
		]
	}}`

	instanceID := "instance"
	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:        testPlanID,
		ServiceID:     testServiceID,
		RawParameters: []byte(params),
	}, true)

	assert.NoError(t, err)

	expected := &atlas.Cluster{
		StateName: "CREATING",

		Name:                     instanceID,
		AutoScaling:              atlas.AutoScalingConfig{DiskGBEnabled: true},
		BackupEnabled:            true,
		BIConnector:              atlas.BIConnectorConfig{Enabled: true, ReadPreference: "primary"},
		ClusterType:              "SHARDED",
		DiskSizeGB:               100.0,
		EncryptionAtRestProvider: "NONE",
		MongoDBMajorVersion:      "4.0",
		NumShards:                2,
		ProviderBackupEnabled:    true,
		ReplicationSpecs: []atlas.ReplicationSpec{
			atlas.ReplicationSpec{
				ID:        "ID",
				NumShards: 2,
				RegionsConfig: map[string]atlas.RegionsConfig{
					"REGION": atlas.RegionsConfig{
						ElectableNodes: 1,
						ReadOnlyNodes:  1,
						AnalyticsNodes: 1,
						Priority:       1,
					},
				},
				ZoneName: "ZONE",
			},
		},
		ProviderSettings: &atlas.ProviderSettings{
			ProviderName:     "AWS",
			InstanceSizeName: "M10",
			RegionName:       "EU_CENTRAL_1",
			DiskIOPS:         10,
			DiskTypeName:     "P4",
			EncryptEBSVolume: true,
			VolumeType:       "STANDARD",
		},
	}

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)
	assert.Equal(t, expected, cluster)
}

func TestProvisionAlreadyExisting(t *testing.T) {
	broker, _, ctx := setupTest()

	// Provision a first instance
	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Try provisioning a second instance with the same ID
	_, err := broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceAlreadyExists.Error())
}

func TestUpdate(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID: testServiceID,
		PlanID:    testPlanID,
	}, true)

	res, err := broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		PlanID:    "aosb-cluster-plan-aws-m20",
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationUpdate, res.OperationData)

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)

	// Ensure the instance size was updated and the provider
	// was not.
	assert.Equal(t, "M20", cluster.ProviderSettings.InstanceSizeName)
	assert.Equal(t, "AWS", cluster.ProviderSettings.ProviderName)
}

func TestUpdateWithoutPlan(t *testing.T) {
	broker, client, ctx := setupTest()

	params := `{
		"cluster": {
			"providerSettings": {
				"regionName": "EU_WEST_1"
			}
		}
	}`

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		ServiceID:     testServiceID,
		PlanID:        testPlanID,
		RawParameters: []byte(params),
	}, true)

	cluster := client.Clusters[instanceID]
	assert.Equal(t, "M10", cluster.ProviderSettings.InstanceSizeName)
	assert.Equal(t, "AWS", cluster.ProviderSettings.ProviderName)
	assert.Equal(t, "EU_WEST_1", cluster.ProviderSettings.RegionName)

	// Try updating the instance without specifying a plan ID. The expected
	// behaviour is for the existing plan (instance size) to remain the same.
	// We also pass params specifying a new region. The broker should fill in the
	// providerSettings with the existing plan.
	updateParams := `{
		"cluster": {
			"providerSettings": {
				"regionName": "EU_CENTRAL_1"
			}
		}
	}`

	res, err := broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		ServiceID:     testServiceID,
		RawParameters: []byte(updateParams),
	}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationUpdate, res.OperationData)

	updatedCluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, updatedCluster, "Expected cluster with name \"%s\" to exist", instanceID)

	// Ensure the service and plan were not changed, whilst the region should
	// have changed.
	assert.Equal(t, "M10", updatedCluster.ProviderSettings.InstanceSizeName)
	assert.Equal(t, "AWS", updatedCluster.ProviderSettings.ProviderName)
	assert.Equal(t, "EU_CENTRAL_1", updatedCluster.ProviderSettings.RegionName)
}

func TestUpdateNonexistent(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	_, err := broker.Update(ctx, instanceID, brokerapi.UpdateDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.Error(t, err, brokerapi.ErrInstanceDoesNotExist.Error())
}

func TestDeprovision(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	res, err := broker.Deprovision(ctx, instanceID, brokerapi.DeprovisionDetails{}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationDeprovision, res.OperationData)
	assert.Nil(t, client.Clusters[instanceID], "Expected cluster to have been removed")
}

func TestDeprovisionWithoutAsync(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	_, err := broker.Deprovision(ctx, instanceID, brokerapi.DeprovisionDetails{}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
	assert.NotEmpty(t, client.Clusters[instanceID], "Expected cluster to not have been removed")
}

func TestDeprovisionNonexistent(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	_, err := broker.Deprovision(ctx, instanceID, brokerapi.DeprovisionDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}

func TestLastOperationProvision(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Set the cluster state to idle
	client.SetClusterState(instanceID, atlas.ClusterStateIdle)
	resp, err := broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
		OperationData: OperationProvision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)

	// Set the cluster state to creating
	client.SetClusterState(instanceID, atlas.ClusterStateCreating)
	resp, err = broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
		OperationData: OperationProvision,
	})

	// State of cluster should be "in progress"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.InProgress, resp.State)
}

func TestLastOperationDeprovision(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Set the cluster state to deleted
	client.SetClusterState(instanceID, atlas.ClusterStateDeleted)
	resp, err := broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)

	// Set the cluster state to deleting
	client.SetClusterState(instanceID, atlas.ClusterStateDeleting)
	resp, err = broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "in progress"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.InProgress, resp.State)

	// Fully remove cluster (causing a not found error)
	client.Clusters[instanceID] = nil
	resp, err = broker.LastOperation(ctx, instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)
}
