package broker

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
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
}

// NewBroker creates a new Broker with the specified Atlas client and logger.
func NewBroker(logger *zap.SugaredLogger) *Broker {
	return &Broker{
		logger: logger,
	}
}

type ContextKey string

var ContextKeyAtlasClient = ContextKey("atlas-client")

func AuthMiddleware(baseURL string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			splitUsername := strings.Split(username, "@")
			groupID := splitUsername[1]
			publicKey := splitUsername[0]

			validUsername := len(splitUsername) == 2
			validCredentials := validUsername && username != "" && password != ""
			if !ok || !validCredentials {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			client := atlas.NewClient(baseURL, groupID, publicKey, password)

			ctx := context.WithValue(r.Context(), ContextKeyAtlasClient, client)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func atlasClientFromContext(ctx context.Context) (atlas.Client, error) {
	client, ok := ctx.Value(ContextKeyAtlasClient).(atlas.Client)
	if !ok {
		return nil, errors.New("no Atlas client in context")
	}

	return client, nil
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
	case atlas.ErrUnauthorized:
		return apiresponses.NewFailureResponse(err, http.StatusUnauthorized, "")
	}

	// Fall back on returning the error again if no others match.
	// Will result in a 500 Internal Server Error.
	return err
}
