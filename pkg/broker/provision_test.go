package broker

import (
	"context"
	"testing"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
)

func TestProvision(t *testing.T) {
	broker, client := SetupTest()

	// Provision a valid instance
	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M10",
		ServiceID: "mongodb",
	}, true)

	if err != nil {
		t.Fatalf("Expected error to be nil, got %v", err)
	}

	if len(client.Clusters) != 1 {
		t.Fatalf("Expected number of clusters to be 1, got %d", len(client.Clusters))
	}

	cluster := client.Clusters[instanceID]
	if cluster == nil {
		t.Fatalf("Expected cluster with name %s to have been created", instanceID)
	}
}

func TestProvisionAlreadyExisting(t *testing.T) {
	broker, _ := SetupTest()

	// Provision a first instance
	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M10",
		ServiceID: "mongodb",
	}, true)

	// Try provisioning a second instance with the same ID
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M20",
		ServiceID: "mongodb",
	}, true)

	if err != apiresponses.ErrInstanceAlreadyExists {
		t.Fatalf("Expected instance already exists error, got %v", err)
	}
}

func TestProvisionWithoutAsync(t *testing.T) {
	broker, client := SetupTest()

	// Try provisioning an instance without async support
	instanceID := "instance"
	_, err := broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M10",
		ServiceID: "mongodb",
	}, false)

	if err != apiresponses.ErrAsyncRequired {
		t.Fatalf("Expected error to be \"async required\", got %v", err)
	}

	// Ensure no clusters were deployed
	if len(client.Clusters) > 0 {
		t.Fatal("Expected no clusters to be created")
	}
}

func TestDeprovision(t *testing.T) {
	broker, client := SetupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M10",
		ServiceID: "mongodb",
	}, true)

	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)

	if err != nil {
		t.Fatalf("Expected error to be nil, got %v", err)
	}

	if client.Clusters[instanceID] != nil {
		t.Fatal("Expected cluster to have been removed")
	}
}

func TestDeprovisionWithoutAsync(t *testing.T) {
	broker, client := SetupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    "AWS-M10",
		ServiceID: "mongodb",
	}, true)

	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, false)

	if err != apiresponses.ErrAsyncRequired {
		t.Fatalf("Expected error to be \"async required\", got %v", err)
	}

	// Ensure the cluster was not terminated
	if client.Clusters[instanceID] == nil {
		t.Fatal("Expected cluster to not be terminated")
	}
}

func TestDeprovisionNonexistent(t *testing.T) {
	broker, _ := SetupTest()

	instanceID := "instance"
	_, err := broker.Deprovision(context.Background(), instanceID, brokerapi.DeprovisionDetails{}, true)

	if err != apiresponses.ErrInstanceDoesNotExist {
		t.Fatalf("Expected instance does not exist error, got %v", err)
	}
}
