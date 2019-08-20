package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
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
	b.logger.Infow("Provisioning instance", "instance_id", instanceID, "details", details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	// Construct a cluster definition from the instance ID, service, plan, and params.
	cluster, err := b.clusterFromParams(instanceID, details.ServiceID, details.PlanID, details.RawParameters)
	if err != nil {
		b.logger.Errorw("Couldn't create cluster from the passed parameters", "error", err, "instance_id", instanceID, "details", details)
		return
	}

	// If cluster doesn't exist create it, otherwise compare them
	resultingCluster, err := b.atlas.GetCluster(cluster.Name)
	if err != nil {
		// Create a new Atlas cluster from the generated definition
		resultingCluster, err = b.atlas.CreateCluster(*cluster)
		if err != nil {
			b.logger.Errorw("Failed to create Atlas cluster", "error", err, "cluster", cluster)
			err = atlasToAPIError(err)
			return
		}

		b.logger.Infow("Successfully started Atlas creation process", "instance_id", instanceID, "cluster", resultingCluster)

		return brokerapi.ProvisionedServiceSpec{
			IsAsync:       true,
			OperationData: OperationProvision,
			DashboardURL:  b.atlas.GetDashboardURL(resultingCluster.Name),
		}, nil
	}

	err = CompareAndReturnAppropiateResponseCode(resultingCluster, cluster)
	return
}

// CompareAndReturnAppropiateResponseCode converts structs to maps and afterwards returns the appropiate response code
func CompareAndReturnAppropiateResponseCode(resultingCluster *atlas.Cluster, cluster *atlas.Cluster) error {
	//Convert structs to maps
	var remoteClusterInterface map[string]interface{}
	var localClusterInterface map[string]interface{}
	inrec, _ := json.Marshal(resultingCluster)
	json.Unmarshal(inrec, &remoteClusterInterface)

	inrec, _ = json.Marshal(cluster)
	json.Unmarshal(inrec, &localClusterInterface)

	if compareHelper(remoteClusterInterface, localClusterInterface) {
		return nil
	}
	return apiresponses.NewFailureResponse(errors.New("There IDs are equal but differ in their attributes"), http.StatusConflict, "")
}

func compareHelper(remoteClusterInterface map[string]interface{}, localClusterInterface map[string]interface{}) bool {
	for k, v := range localClusterInterface {
		if reflect.ValueOf(v).Kind() == reflect.Map && len(v.(map[string]interface{})) != 0 {
			if val, ok := remoteClusterInterface[k]; ok { // Check to see if key is present
				if reflect.ValueOf(val).Kind() == reflect.Map && len(val.(map[string]interface{})) != 0 { // Must be of type map too
					equal := compareHelper(val.(map[string]interface{}), v.(map[string]interface{}))
					if equal {
						continue
					}
				}
			}
			return false
		} else if reflect.ValueOf(v).Kind() == reflect.Slice && len(v.([]interface{})) != 0 {
			for index, document := range v.([]interface{}) {
				if val, ok := remoteClusterInterface[k]; ok {
					if reflect.ValueOf(val).Kind() == reflect.Slice && len(val.([]interface{})) != 0 {
						if reflect.ValueOf(val.([]interface{})[index]).Kind() == reflect.Map && len(val.([]interface{})[index].(map[string]interface{})) != 0 {
							equal := compareHelper(val.([]interface{})[index].(map[string]interface{}), document.(map[string]interface{}))
							if equal {
								continue
							}
						}
					}
				}
				return false
			}
		} else if reflect.ValueOf(v).Kind() != reflect.Map && reflect.ValueOf(v).Kind() != reflect.Slice { // Others are not maps nor slices, but rather other types
			if val, ok := remoteClusterInterface[k]; ok {
				if reflect.TypeOf(v) != reflect.TypeOf(val) { // Must be of same type
					return false
				}
				if v != val {
					return false
				}
				continue
			}
			return false // Not present in the remote cluster, so they differ in attributes
		} else { // Empty map or slice found
			continue
		}
	}

	return true
}

// Update will change the configuration of an existing Atlas cluster asynchronously.
func (b Broker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (spec brokerapi.UpdateServiceSpec, err error) {
	b.logger.Infow("Updating instance", "instance_id", instanceID, "details", details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	// Fetch the cluster from Atlas. The Atlas API requires an instance size to
	// be passed during updates (if there are other update to the provider, such
	// as region). The plan is not included in the OSB call unless it has changed
	// hence we need to fetch the current value from Atlas.
	existingCluster, err := b.atlas.GetCluster(NormalizeClusterName(instanceID))
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	// Construct a cluster from the instance ID, service, plan, and params.
	cluster, err := b.clusterFromParams(instanceID, details.ServiceID, details.PlanID, details.RawParameters)
	if err != nil {
		return
	}

	// Make sure the cluster provider has all the neccessary params for the
	// Atlas API. The Atlas API requires both the provider name and instance
	// size if the provider object is set. If they are missing we use the
	// existing values.
	if cluster.ProviderSettings != nil {
		if cluster.ProviderSettings.ProviderName == "" {
			cluster.ProviderSettings.ProviderName = existingCluster.ProviderSettings.ProviderName
		}

		if cluster.ProviderSettings.InstanceSizeName == "" {
			cluster.ProviderSettings.InstanceSizeName = existingCluster.ProviderSettings.InstanceSizeName
		}
	}

	resultingCluster, err := b.atlas.UpdateCluster(*cluster)
	if err != nil {
		b.logger.Errorw("Failed to update Atlas cluster", "error", err, "cluster", cluster)
		err = atlasToAPIError(err)
		return
	}

	b.logger.Infow("Successfully started Atlas cluster update process", "instance_id", instanceID, "cluster", resultingCluster)

	return brokerapi.UpdateServiceSpec{
		IsAsync:       true,
		OperationData: OperationUpdate,
		DashboardURL:  b.atlas.GetDashboardURL(resultingCluster.Name),
	}, nil
}

// Deprovision will destroy an Atlas cluster asynchronously.
func (b Broker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (spec brokerapi.DeprovisionServiceSpec, err error) {
	b.logger.Infow("Deprovisioning instance", "instance_id", instanceID, "details", details)

	// Async needs to be supported for provisioning to work.
	if !asyncAllowed {
		err = apiresponses.ErrAsyncRequired
		return
	}

	err = b.atlas.DeleteCluster(NormalizeClusterName(instanceID))
	if err != nil {
		b.logger.Errorw("Failed to delete Atlas cluster", "error", err, "instance_id", instanceID)
		err = atlasToAPIError(err)
		return
	}

	b.logger.Infow("Successfully started Atlas cluster deletion process", "instance_id", instanceID)

	return brokerapi.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: OperationDeprovision,
	}, nil
}

// GetInstance is currently not supported as specified by the
// InstancesRetrievable setting in the service catalog.
func (b Broker) GetInstance(ctx context.Context, instanceID string) (spec brokerapi.GetInstanceDetailsSpec, err error) {
	b.logger.Infow("Fetching instance", "instance_id", instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown instance ID %s", instanceID), 404, "get-instance")
	return
}

// LastOperation should fetch the state of the provision/deprovision
// of a cluster.
func (b Broker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (resp brokerapi.LastOperation, err error) {
	b.logger.Infow("Fetching state of last operation", "instance_id", instanceID, "details", details)

	cluster, err := b.atlas.GetCluster(NormalizeClusterName(instanceID))
	if err != nil && err != atlas.ErrClusterNotFound {
		b.logger.Errorw("Failed to get existing cluster", "error", err, "instance_id", instanceID)
		err = atlasToAPIError(err)
		return
	}

	b.logger.Infow("Found existing cluster", "cluster", cluster)

	state := brokerapi.LastOperationState(brokerapi.Failed)

	switch details.OperationData {
	case OperationProvision:
		switch cluster.StateName {
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
		if err == atlas.ErrClusterNotFound || cluster.StateName == atlas.ClusterStateDeleted {
			state = brokerapi.Succeeded
		} else if cluster.StateName == atlas.ClusterStateDeleting {
			state = brokerapi.InProgress
		}
	case OperationUpdate:
		// We assume that the cluster transitions to the "UPDATING" state
		// in a synchronous manner during the update request.
		switch cluster.StateName {
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

// NormalizeClusterName will sanitize a name to make sure it will be accepted
// by the Atlas API. Atlas has different name length requirements depending on
// which environment it's running in. A length of 23 is a safe choice and
// truncates UUIDs nicely.
func NormalizeClusterName(name string) string {
	const maximumNameLength = 23

	if len(name) > maximumNameLength {
		return string(name[0:maximumNameLength])
	}

	return name
}

// clusterFromParams will construct a cluster object from an instance ID,
// service, plan, and raw parameters. This way users can pass all the
// configuration available for clusters in the Atlas API as "cluster" in the params.
func (b Broker) clusterFromParams(instanceID string, serviceID string, planID string, rawParams []byte) (*atlas.Cluster, error) {
	// Set up a params object which will be used for deserialiation.
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
		provider, err := b.findProviderByServiceID(serviceID)
		if err != nil {
			return nil, err
		}

		instanceSize, err := findInstanceSizeByPlanID(provider, planID)
		if err != nil {
			return nil, err
		}

		if params.Cluster.ProviderSettings == nil {
			params.Cluster.ProviderSettings = &atlas.ProviderSettings{}
		}

		// Configure provider based on service and plan.
		params.Cluster.ProviderSettings.ProviderName = provider.Name
		params.Cluster.ProviderSettings.InstanceSizeName = instanceSize.Name
	}

	// Add the instance ID as the name of the cluster.
	params.Cluster.Name = NormalizeClusterName(instanceID)

	return params.Cluster, nil
}
