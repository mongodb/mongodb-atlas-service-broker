package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/10gen/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
)

// The different async operations that can be performed.
// These constants are returned during provisioning, deprovisioning, and
// updates and are subsequently included in async polls from the platform.
const (
	OperationProvision   = "provision"
	OperationDeprovision = "deprovision"
	OperationUpdate      = "update"
)

// Provision will create a new Atlas cluster with the instance ID as its name.
// The process is always async.
func (b Broker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (spec brokerapi.ProvisionedServiceSpec, err error) {
	b.logger.Infof("Provisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	// Construct a cluster from the instance ID, service, plan, and params.
	cluster, err := clusterFromParams(instanceID, details.ServiceID, details.PlanID, details.RawParameters)
	if err != nil {
		return
	}

	// Create a new Atlas cluster with the instance ID as its name.
	_, err = b.atlas.CreateCluster(*cluster)
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

// Update will change the configuration of an existing Atlas cluster asynchronously.
func (b Broker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (spec brokerapi.UpdateServiceSpec, err error) {
	b.logger.Infof("Updating instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	return brokerapi.UpdateServiceSpec{
		IsAsync:       true,
		OperationData: OperationUpdate,
	}, nil
}

// Deprovision will destroy an Atlas cluster asynchronously.
func (b Broker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (spec brokerapi.DeprovisionServiceSpec, err error) {
	b.logger.Infof("Deprovisioning instance \"%s\" with details %+v", instanceID, details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	err = b.atlas.DeleteCluster(normalizeClusterName(instanceID))
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
func (b Broker) GetInstance(ctx context.Context, instanceID string) (spec brokerapi.GetInstanceDetailsSpec, err error) {
	b.logger.Infof("Fetching instance \"%s\"", instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown instance ID %s", instanceID), 404, "get-instance")
	return
}

// LastOperation should fetch the state of the provision/deprovision
// of a cluster.
func (b Broker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (resp brokerapi.LastOperation, err error) {
	b.logger.Infof("Fetching state of last operation for instance \"%s\" with details %+v", instanceID, details)

	cluster, err := b.atlas.GetCluster(normalizeClusterName(instanceID))
	if err != nil && err != atlas.ErrClusterNotFound {
		b.logger.Error(err)
		err = atlasToAPIError(err)
		return
	}

	state := brokerapi.LastOperationState(brokerapi.Failed)

	switch details.OperationData {
	case OperationProvision:
		switch cluster.State {
		// Provision has succeeded if the cluster is in state "idle".
		case atlas.ClusterStateIdle:
			state = brokerapi.Succeeded
		case atlas.ClusterStateCreating:
			state = brokerapi.InProgress
		}
	case OperationDeprovision:
		// The Atlas API may return a 404 response if a cluster is deleted or it
		// will return the cluster with a state of "DELETED". Both of these
		// scenarios indicate that a cluster has been successfully deleted.
		if err == atlas.ErrClusterNotFound || cluster.State == atlas.ClusterStateDeleted {
			state = brokerapi.Succeeded
		} else if cluster.State == atlas.ClusterStateDeleting {
			state = brokerapi.InProgress
		}
	case OperationUpdate:
		// We assume that the cluster transitions to the "UPDATING" state
		// in a synchronous manner during the update request.
		switch cluster.State {
		case atlas.ClusterStateIdle:
			state = brokerapi.Succeeded
		case atlas.ClusterStateUpdating:
			state = brokerapi.InProgress
		}
	}

	return brokerapi.LastOperation{
		State: state,
	}, nil
}

// normalizeClusterName will sanitize a name to make sure it will be accepted
// by the Atlas API. Atlas requires cluster names to be 30 characters or less.
func normalizeClusterName(name string) string {
	if len(name) > 30 {
		return string(name[0:30])
	}

	return name
}

// clusterFromParams will construct a cluster object from an instance ID,
// service, plan, and raw parameters. This way users can pass all the
// configuration available for clusters in the Atlas API as "cluster" in the params.
func clusterFromParams(instanceID string, serviceID string, planID string, rawParams []byte) (*atlas.Cluster, error) {
	// Set up a params object with default values.
	params := struct {
		Cluster *atlas.Cluster `json:"cluster"`
	}{
		&atlas.Cluster{},
	}

	// If params were passed we unmarshal them into the params object.
	if len(rawParams) > 0 {
		err := json.Unmarshal(rawParams, &params)
		if err != nil {
			return nil, err
		}
	}

	// If the plan ID is specified we construct the provider object from the service and plan.
	// The plan ID is optional during updates but not during creation.
	if planID != "" {
		cloud, size := cloudFromPlan(serviceID, planID)
		if cloud == nil || size == nil {
			return nil, errors.New("Invalid service ID or plan ID")
		}

		if params.Cluster.ProviderSettings == nil {
			params.Cluster.ProviderSettings = &atlas.ProviderSettings{}
		}

		// Configure provider based on service and plan.
		params.Cluster.ProviderSettings.Name = cloud.Name
		params.Cluster.ProviderSettings.Instance = size.Name
	}

	// Add the instance ID as the name of the cluster.
	params.Cluster.Name = normalizeClusterName(instanceID)

	return params.Cluster, nil
}