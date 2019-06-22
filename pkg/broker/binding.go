package broker

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/fabianlindfors/atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
)

type ConnectionDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URI      string `json:"uri"`
}

// Bind will create a new database user with a username matching the binding ID
// and a randomly generated password. The user credentials will be returned back.
func (b *Broker) Bind(ctx context.Context, instanceID string, bindingID string, details brokerapi.BindDetails, asyncAllowed bool) (spec brokerapi.Binding, err error) {
	b.logger.Infof("Creating binding \"%s\" for instance \"%s\" with details %+v", bindingID, instanceID, details)

	cluster, err := b.atlas.GetCluster(instanceID)
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	// Create a new user with the binding ID as its username and a randomly
	// generated password.
	password, err := generatePassword()
	if err != nil {
		b.logger.Error("Failed to generate password", err)
		err = errors.New("Failed to generate binding password")
		return
	}

	_, err = b.atlas.CreateUser(atlas.User{Username: bindingID, Password: password})
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	// TODO: Place credentials in some sort of secrets manager.

	spec = brokerapi.Binding{
		Credentials: ConnectionDetails{
			Username: bindingID,
			Password: password,
			URI:      cluster.URI,
		},
	}
	return
}

// Disconnect/unbind an application from an Atlas cluster
func (b *Broker) Unbind(ctx context.Context, instanceID string, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (spec brokerapi.UnbindSpec, err error) {
	b.logger.Infof("Releasing binding \"%s\" for instance \"%s\" with details %+v", bindingID, instanceID, details)

	_, err = b.atlas.GetCluster(instanceID)
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	err = b.atlas.DeleteUser(bindingID)
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	spec = brokerapi.UnbindSpec{}
	return
}

func (b *Broker) GetBinding(ctx context.Context, instanceID string, bindingID string) (spec brokerapi.GetBindingSpec, err error) {
	b.logger.Infof("Retrieving binding \"%s\" for instance \"%s\"", bindingID, instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown binding ID %s", bindingID), 404, "get-binding")
	return
}

// generatePassword will generate a cryptographically secure password.
// The password will be base64 encoded for easy usage.
func generatePassword() (string, error) {
	const numberOfBytes = 32
	b := make([]byte, numberOfBytes)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
