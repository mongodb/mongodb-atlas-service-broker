package broker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/10gen/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/stretchr/testify/assert"
)

var (
	testServiceID = "mongodb-aws"
	testPlanID    = "AWS-M10"
)

// TestMissingAsync will make sure all async operations don't accept non-async
// clients.
func TestMissingAsync(t *testing.T) {
	broker, client := setupTest()

	// Try provisioning an instance without async support
	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
	assert.Len(t, client.Clusters, 0, "Expected no clusters to be created")

	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Try updating existing cluster without async support
	_, err = broker.Update(context.Background(), instanceID, brokerapi.UpdateDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())

	_, err = broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
}

func TestProvision(t *testing.T) {
	broker, client := setupTest()

	params, err := json.Marshal(map[string]string{
		"region": "EU_CENTRAL_1",
	})
	assert.NoError(t, err)

	instanceID := "instance"
	res, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:        testPlanID,
		ServiceID:     testServiceID,
		RawParameters: params,
	}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationProvision, res.OperationData)
	assert.Len(t, client.Clusters, 1)

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)
	assert.Equal(t, &atlas.ProviderSettings{
		Name:     "AWS",
		Instance: "M10",
		Region:   "EU_CENTRAL_1",
	}, cluster.Provider)
}

func TestProvisionDefaultRegion(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)
	assert.Len(t, client.Clusters, 1)

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)
	assert.Equal(t, &atlas.ProviderSettings{
		Name:     "AWS",
		Instance: "M10",
		Region:   "EU_WEST_1",
	}, cluster.Provider)
}

func TestProvisionInvalidRegion(t *testing.T) {
	broker, _ := setupTest()

	params, err := json.Marshal(map[string]string{
		"region": "NON_EXISTENT_REGION",
	})
	assert.NoError(t, err)

	instanceID := "instance"
	_, err = broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:        testPlanID,
		ServiceID:     testServiceID,
		RawParameters: params,
	}, true)

	assert.Error(t, err, "Invalid region \"NON_EXISTENT_REGION\"")
}

func TestProvisionAlreadyExisting(t *testing.T) {
	broker, _ := setupTest()

	// Provision a first instance
	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Try provisioning a second instance with the same ID
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceAlreadyExists.Error())
}

func TestUpdate(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	params, err := json.Marshal(map[string]string{
		"region": "EU_CENTRAL_1",
	})
	assert.NoError(t, err)

	res, err := broker.Update(context.Background(), instanceID, brokerapi.UpdateDetails{
		PlanID:        "AWS-M20",
		ServiceID:     testServiceID,
		RawParameters: params,
	}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationUpdate, res.OperationData)

	cluster := client.Clusters[instanceID]
	assert.NotEmptyf(t, cluster, "Expected cluster with name \"%s\" to exist", instanceID)

	// Ensure the instance size and region was updated and the provider
	// was not.
	assert.Equal(t, "M20", cluster.Provider.Instance)
	assert.Equal(t, "AWS", cluster.Provider.Name)
	assert.Equal(t, "EU_CENTRAL_1", cluster.Provider.Region)
}

func TestUpdateNonexistent(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	_, err := broker.Update(context.Background(), instanceID, brokerapi.UpdateDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.Error(t, err, brokerapi.ErrInstanceDoesNotExist.Error())
}

func TestDeprovision(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	res, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)

	assert.NoError(t, err)
	assert.True(t, res.IsAsync)
	assert.Equal(t, OperationDeprovision, res.OperationData)
	assert.Nil(t, client.Clusters[instanceID], "Expected cluster to have been removed")
}

func TestDeprovisionWithoutAsync(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())
	assert.NotEmpty(t, client.Clusters[instanceID], "Expected cluster to not have been removed")
}

func TestDeprovisionNonexistent(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}

func TestLastOperationProvision(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Set the cluster state to idle
	client.SetClusterState(instanceID, atlas.ClusterStateIdle)
	resp, err := broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
		OperationData: OperationProvision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)

	// Set the cluster state to creating
	client.SetClusterState(instanceID, atlas.ClusterStateCreating)
	resp, err = broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
		OperationData: OperationProvision,
	})

	// State of cluster should be "in progress"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.InProgress, resp.State)
}

func TestLastOperationDeprovision(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	// Set the cluster state to deleted
	client.SetClusterState(instanceID, atlas.ClusterStateDeleted)
	resp, err := broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)

	// Set the cluster state to deleting
	client.SetClusterState(instanceID, atlas.ClusterStateDeleting)
	resp, err = broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "in progress"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.InProgress, resp.State)

	// Fully remove cluster (causing a not found error)
	client.Clusters[instanceID] = nil
	resp, err = broker.LastOperation(context.Background(), instanceID, brokerapi.PollDetails{
		OperationData: OperationDeprovision,
	})

	// State of cluster should be "succeeded"
	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, resp.State)
}
