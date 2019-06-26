package broker

import (
	"context"
	"fmt"
	"strings"

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
	clouds := clouds()
	services := make([]brokerapi.Service, len(clouds))

	for i, cloud := range clouds {
		services[i] = brokerapi.Service{
			ID:                   cloud.ID(),
			Name:                 cloud.ID(),
			Description:          fmt.Sprintf("Cluster hosted on \"%s\"", cloud.Name),
			Bindable:             true,
			InstancesRetrievable: false,
			BindingsRetrievable:  false,
			Metadata:             nil,
			Plans:                plansForCloud(cloud),
		}
	}

	return services, nil
}

func plansForCloud(cloud Cloud) []brokerapi.ServicePlan {
	plans := make([]brokerapi.ServicePlan, len(cloud.Sizes))

	for i, size := range cloud.Sizes {
		plans[i] = brokerapi.ServicePlan{
			ID:          size.ID(cloud),
			Name:        size.Name,
			Description: fmt.Sprintf("Instance size \"%s\"", size.Name),
		}
	}

	return plans
}

type Cloud struct {
	Name    string
	Regions []string
	Sizes   []Size
}

func (c Cloud) ID() string {
	return fmt.Sprintf("mongodb-%s", strings.ToLower(c.Name))
}

func (c Cloud) DefaultRegion() string {
	return c.Regions[0]
}

func (c Cloud) ValidateRegion(region string) error {
	for _, validRegion := range c.Regions {
		if validRegion == region {
			return nil
		}
	}

	return fmt.Errorf("Invalid region %s", region)
}

type Size struct {
	Name string
}

func (s Size) ID(cloud Cloud) string {
	return fmt.Sprintf("%s-%s", cloud.Name, s.Name)
}

func clouds() []Cloud {
	return []Cloud{
		Cloud{
			Name:    "AWS",
			Regions: []string{"EU_WEST_1", "EU_CENTRAL_1"},
			Sizes: []Size{
				Size{Name: "M10"},
				Size{Name: "M20"},
				Size{Name: "M30"},
				Size{Name: "M40"},
				Size{Name: "R40"},
				Size{Name: "M40_NVME"},
				Size{Name: "M50"},
				Size{Name: "R50"},
				Size{Name: "M50_NVME"},
				Size{Name: "M60"},
				Size{Name: "R60"},
				Size{Name: "M60_NVME"},
				Size{Name: "R80"},
				Size{Name: "M80_NVME"},
				Size{Name: "M100"},
				Size{Name: "M140"},
				Size{Name: "M200"},
				Size{Name: "R200"},
				Size{Name: "M200_NVME"},
				Size{Name: "M300"},
				Size{Name: "R400"},
				Size{Name: "M400_NVME"},
			},
		},
		Cloud{
			Name: "GCP",
			Sizes: []Size{
				Size{Name: "M10"},
			},
		},
		Cloud{
			Name: "AZURE",
			Sizes: []Size{
				Size{Name: "M10"},
			},
		},
	}
}

func cloudFromPlan(serviceID string, planID string) (*Cloud, *Size) {
	for _, cloud := range clouds() {
		if cloud.ID() == serviceID {
			for _, size := range cloud.Sizes {
				if size.ID(cloud) == planID {
					return &cloud, &size
				}
			}
		}
	}

	return nil, nil
}

// atlasToAPIError converts an Atlas error to a OSB response error.
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

	// Fall back on returning the error again if no others match.
	// Will result in a 500 Internal Server Error.
	return err
}
