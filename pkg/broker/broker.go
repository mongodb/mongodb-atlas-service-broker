package broker

import (
	"context"
	"fmt"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"go.uber.org/zap"
)

// Broker is responsible for translating OSB calls to Atlas API calls.
// Implements the Broker interface from brokerapi making it easy to spin up an
// API server.
type Broker struct {
	logger *zap.SugaredLogger
	atlas  atlas.Client
}

// NewBroker creates a new Broker with the specified Atlas client and logger.
func NewBroker(client atlas.Client, logger *zap.SugaredLogger) *Broker {
	return &Broker{
		logger: logger,
		atlas:  client,
	}
}

// Services generates the service catalog which will be presented to consumers of the API.
func (b *Broker) Services(ctx context.Context) ([]brokerapi.Service, error) {
	plans := Plans()
	servicePlans := make([]brokerapi.ServicePlan, len(plans))

	for i, plan := range plans {
		servicePlans[i] = brokerapi.ServicePlan{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
		}
	}

	return []brokerapi.Service{
		brokerapi.Service{
			ID:                   "mongodb",
			Name:                 "mongodb",
			Description:          "DESCRIPTION",
			Bindable:             true,
			InstancesRetrievable: true,
			BindingsRetrievable:  true,
			Metadata:             nil,
			Plans:                servicePlans,
		},
	}, nil
}

// Plan represents a single plan for the service with an associated instance
// size and broker.
type Plan struct {
	ID           string
	Name         string
	Description  string
	Instance     string
	ProviderName string
}

// Provider returns the Atlas provider settings corresponding to the plan.
func (p *Plan) Provider() atlas.Provider {
	return atlas.Provider{
		Name:     p.ProviderName,
		Instance: p.Instance,
		Region:   "EU_WEST_1",
	}
}

// Plans return all available plans across all providers
func Plans() []Plan {
	return append(providerPlans("AWS"), providerPlans("GCP")...)
}

func providerPlans(provider string) []Plan {
	instanceSizes := []string{"M10", "M20"}

	var plans []Plan

	// AWS Instances
	for _, instance := range instanceSizes {
		plans = append(plans, Plan{
			ID:           fmt.Sprintf("%s-%s", provider, instance),
			Name:         fmt.Sprintf("%s-%s", provider, instance),
			Description:  fmt.Sprintf("Instance size %s on %s", instance, provider),
			Instance:     instance,
			ProviderName: provider,
		})
	}

	return plans
}

// findPlan search all available plans by ID
func findPlan(id string) *Plan {
	for _, plan := range Plans() {
		if plan.ID == id {
			return &plan
		}
	}

	return nil
}

func (b *Broker) GetInstance(ctx context.Context, instanceID string) (spec brokerapi.GetInstanceDetailsSpec, err error) {
	b.logger.Infof("Fetching instance \"%s\"", instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown instance ID %s", instanceID), 404, "get-instance")
	return
}

func (b *Broker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	b.logger.Infof("Updating instance \"%s\" with details %+v", instanceID, details)
	return brokerapi.UpdateServiceSpec{
		IsAsync: true,
	}, nil
}

func (b *Broker) LastOperation(ctx context.Context, instanceID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	b.logger.Infof("Fetching state of last operation for instance \"%s\" with details %+v", instanceID, details)
	return brokerapi.LastOperation{
		State: brokerapi.Succeeded,
	}, nil
}

func (b *Broker) LastBindingOperation(ctx context.Context, instanceID string, bindingID string, details brokerapi.PollDetails) (brokerapi.LastOperation, error) {
	panic("not implemented")
}

// atlasToAPIError converts an Atlas error to a OSB response error
func atlasToAPIError(err error) error {
	switch err {
	case atlas.ErrClusterNotFound:
		return apiresponses.ErrInstanceDoesNotExist
	case atlas.ErrClusterAlreadyExists:
		return apiresponses.ErrInstanceAlreadyExists
	case atlas.ErrUserAlreadyExists:
		return apiresponses.ErrBindingAlreadyExists
	case atlas.ErrUserNotFound:
		return apiresponses.ErrBindingDoesNotExist
	}

	// Fall back on returning the error again if no others match
	// Will result in a 500 Internal Server Error
	return err
}
