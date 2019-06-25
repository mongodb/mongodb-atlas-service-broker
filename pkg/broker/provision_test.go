package broker

import (
	"context"
	"testing"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/stretchr/testify/assert"
)

var (
	testServiceID = "mongodb-aws"
	testPlanID    = "AWS-M10"
)

func TestProvision(t *testing.T) {
	broker, client := setupTest()

	// Provision a valid instance
	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)
	assert.Len(t, client.Clusters, 1)
	assert.NotEmptyf(t, client.Clusters[instanceID], "Expected cluster with name \"%s\" to exist", instanceID)
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

func TestProvisionWithoutAsync(t *testing.T) {
	broker, client := setupTest()

	// Try provisioning an instance without async support
	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, false)

	assert.EqualError(t, err, apiresponses.ErrAsyncRequired.Error())

	// Ensure no clusters were deployed
	assert.Len(t, client.Clusters, 0, "Expected no clusters to be created")
}

func TestDeprovision(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)

	assert.NoError(t, err)
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
