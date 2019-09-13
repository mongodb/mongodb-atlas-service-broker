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
	logger    *zap.SugaredLogger
	whitelist Whitelist
}

// NewBroker creates a new Broker with a logger.
func NewBroker(logger *zap.SugaredLogger) *Broker {
	return &Broker{
		logger: logger,
	}
}

// NewBrokerWithWhitelist creates a new Broker with a given logger and a
// whitelist for allowed providers and their plans.
func NewBrokerWithWhitelist(logger *zap.SugaredLogger, whitelist Whitelist) *Broker {
	return &Broker{
		logger:    logger,
		whitelist: whitelist,
	}
}

// ContextKey represents the key for a value saved in a context. Linter
// requires keys to have their own type.
type ContextKey string

// ContextKeyAtlasClient is the key used to store the Atlas client in the
// request context.
var ContextKeyAtlasClient = ContextKey("atlas-client")

// AuthMiddleware is used to validate and parse Atlas API credentials passed
// using basic auth. The credentials parsed into an Atlas client which is
// attached to the request context. This client can later be retrieved by the
// broker from the context.
func AuthMiddleware(baseURL string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			// The username contains both the group ID and public key
			// formatted as "<PUBLIC_KEY>@<GROUP_ID>".
			splitUsername := strings.Split(username, "@")

			// If the credentials are invalid we respond with 401 Unauthorized.
			// The username needs have the correct format and the password must
			// not be empty.
			validUsername := len(splitUsername) == 2
			validPassword := password != ""
			if !(ok && validUsername && validPassword) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Create a new client with the extracted API credentials and
			// attach it to the request context.
			client := atlas.NewClient(baseURL, splitUsername[1], splitUsername[0], password)
			ctx := context.WithValue(r.Context(), ContextKeyAtlasClient, client)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// atlasClientFromContext will retrieve an Atlas client stored inside the
// provided context.
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
