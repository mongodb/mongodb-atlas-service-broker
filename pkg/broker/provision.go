package broker

import (
	"context"
	"fmt"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
)

const (
	OperationProvision   = "provision"
	OperationDeprovision = "deprovision"
	OperationUpdate      = "update"
)

// Provision will create a new Atlas cluster with the instance ID as its name.
// The process is always async.
func (b *Broker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (spec brokerapi.ProvisionedServiceSpec, err error) {
	b.logger.Infof("Provisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported to provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	// Find the plan corresponding to the passed plan ID.
	plan := findPlan(details.PlanID)

	// Create a new Atlas cluster with the instance ID as its name.
	_, err = b.atlas.CreateCluster(atlas.Cluster{
		Name:     sanitizeClusterName(instanceID),
		Provider: plan.Provider(),
	})
	if err != nil {
		b.logger.Error(err)
		err = atlasToAPIError(err)
		return
	}

	spec = brokerapi.ProvisionedServiceSpec{
		IsAsync:       true,
		OperationData: OperationProvision,
	}
	return
}

// Deprovision will destroy an Atlas cluster asynchronously.
func (b *Broker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (spec brokerapi.DeprovisionServiceSpec, err error) {
	b.logger.Infof("Deprovisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	err = b.atlas.TerminateCluster(sanitizeClusterName(instanceID))
	if err != nil {
		b.logger.Error(err)
		err = atlasToAPIError(err)
		return
	}

	spec = brokerapi.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: OperationDeprovision,
	}
	return
}

// GetInstance is currently not supported as specified by the
// InstancesRetrievable setting in the service catalog.
func (b *Broker) GetInstance(ctx context.Context, instanceID string) (spec brokerapi.GetInstanceDetailsSpec, err error) {
	b.logger.Infof("Fetching instance \"%s\"", instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown instance ID %s", instanceID), 404, "get-instance")
	return
}

// Update is not yet impemented.
func (b *Broker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	b.logger.Infof("Updating instance \"%s\" with details %+v", instanceID, details)
	return brokerapi.UpdateServiceSpec{
		IsAsync:       true,
		OperationData: OperationUpdate,
	}, nil
}

// LastOperation should fetch the state of the provision/deprovision
// of a cluster.
func (b *Broker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (resp brokerapi.LastOperation, err error) {
	b.logger.Infof("Fetching state of last operation for instance \"%s\" with details %+v", instanceID, details)

	cluster, err := b.atlas.GetCluster(sanitizeClusterName(instanceID))
	if err != nil && err != atlas.ErrClusterNotFound {
		b.logger.Error(err)
		err = atlasToAPIError(err)
		return
	}

	state := brokerapi.LastOperationState(brokerapi.Failed)

	switch details.OperationData {
	case OperationProvision:
		switch cluster.State {
		case atlas.ClusterStateIdle:
			state = brokerapi.Succeeded
		case atlas.ClusterStateCreating:
			state = brokerapi.InProgress
		}
	case OperationDeprovision:
		if err == atlas.ErrClusterNotFound || cluster.State == atlas.ClusterStateDeleted {
			state = brokerapi.Succeeded
		} else if cluster.State == atlas.ClusterStateDeleting {
			state = brokerapi.InProgress
		}
	case OperationUpdate:
		// TODO: Implement once update has been implemented
	}

	return brokerapi.LastOperation{
		State: state,
	}, nil
}

func sanitizeClusterName(name string) string {
	trimmed := name[0:30]
	return string(trimmed)
}
