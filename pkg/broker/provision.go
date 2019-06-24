package broker

import (
	"context"
	"fmt"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
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
		Name:     instanceID,
		Provider: plan.Provider(),
	})
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	spec = brokerapi.ProvisionedServiceSpec{
		IsAsync: true,
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

	err = b.atlas.TerminateCluster(instanceID)
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	spec = brokerapi.DeprovisionServiceSpec{
		IsAsync: true,
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
		IsAsync: true,
	}, nil
}

// LastOperation should fetch the state of the provision/deprovision
// of a cluster.
func (b *Broker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	b.logger.Infof("Fetching state of last operation for instance \"%s\" with details %+v", instanceID, details)
	return brokerapi.LastOperation{
		State: brokerapi.Succeeded,
	}, nil
}
