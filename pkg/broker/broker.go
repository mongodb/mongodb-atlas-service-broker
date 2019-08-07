package broker

import (
	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"go.uber.org/zap"
)

// Ensure broker adheres to the ServiceBroker interface.
var _ brokerapi.ServiceBroker = Broker{}

// Broker is responsible for translating OSB calls to Atlas API calls.
// Implements the brokerapi.ServiceBroker interface making it easy to spin up
// an API server.
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
