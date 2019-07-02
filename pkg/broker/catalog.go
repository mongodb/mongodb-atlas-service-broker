package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/pivotal-cf/brokerapi"
)

// Services generates the service catalog which will be presented to consumers of the API.
func (b Broker) Services(ctx context.Context) ([]brokerapi.Service, error) {
	b.logger.Info("Retrieving service catalog")

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
			Plans:                cloud.Plans(),
		}
	}

	return services, nil
}

func clouds() []cloud {
	return []cloud{
		cloud{
			Name: "AWS",
			Sizes: []size{
				size{Name: "M10"},
				size{Name: "M20"},
				size{Name: "M30"},
				size{Name: "M40"},
				size{Name: "R40"},
				size{Name: "M40_NVME"},
				size{Name: "M50"},
				size{Name: "R50"},
				size{Name: "M50_NVME"},
				size{Name: "M60"},
				size{Name: "R60"},
				size{Name: "M60_NVME"},
				size{Name: "R80"},
				size{Name: "M80_NVME"},
				size{Name: "M100"},
				size{Name: "M140"},
				size{Name: "M200"},
				size{Name: "R200"},
				size{Name: "M200_NVME"},
				size{Name: "M300"},
				size{Name: "R400"},
				size{Name: "M400_NVME"},
			},
		},
		cloud{
			Name: "GCP",
			Sizes: []size{
				size{Name: "M10"},
				size{Name: "M20"},
				size{Name: "M30"},
				size{Name: "M40"},
				size{Name: "M50"},
				size{Name: "M60"},
				size{Name: "M80"},
				size{Name: "M200"},
				size{Name: "M300"},
			},
		},
		cloud{
			Name: "AZURE",
			Sizes: []size{
				size{Name: "M10"},
				size{Name: "M20"},
				size{Name: "M30"},
				size{Name: "M40"},
				size{Name: "M50"},
				size{Name: "M60"},
				size{Name: "M80"},
				size{Name: "M200"},
			},
		},
	}
}

func cloudFromPlan(serviceID string, planID string) (*cloud, *size) {
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

// cloud represents a single available cloud in which a cluster can be deployed.
type cloud struct {
	Name  string
	Sizes []size
}

// Plans generates service broker plans based on the available instance sizes.
func (c cloud) Plans() []brokerapi.ServicePlan {
	plans := make([]brokerapi.ServicePlan, len(c.Sizes))

	for i, size := range c.Sizes {
		plans[i] = brokerapi.ServicePlan{
			ID:          size.ID(c),
			Name:        size.Name,
			Description: fmt.Sprintf("Instance size \"%s\"", size.Name),
		}
	}

	return plans
}

// ID generates a unique service ID for use in the catalog.
func (c cloud) ID() string {
	return fmt.Sprintf("mongodb-%s", strings.ToLower(c.Name))
}

// size represents a single instance size which clusters can use.
// TODO: Add memory and storage to generate better plan descriptions.
type size struct {
	Name string
}

// ID will generate a unique plan ID for use in the catalog.
func (s size) ID(cloud cloud) string {
	return fmt.Sprintf("%s-%s", cloud.Name, s.Name)
}
