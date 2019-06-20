package broker

import (
	"context"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
)

// Provision will create a new Atlas cluster with the instance ID as its name.
// The process is always async.
func (b *Broker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (spec brokerapi.ProvisionedServiceSpec, err error) {
	b.logger.Infof("Provisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported to provisioning to work
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	// Find the plan corresponding to the passed plan ID
	plan := findPlan(details.PlanID)

	// Create a new Atlas cluster with the instance ID as its name
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

// Deprovision will destroy an Atlas cluster.
func (b *Broker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (spec brokerapi.DeprovisionServiceSpec, err error) {
	b.logger.Infof("Deprovisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported to provisioning to work
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
