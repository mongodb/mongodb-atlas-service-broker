package broker

import (
	"context"
	"fmt"
	"math/rand"

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
	password := generatePassword()
	_, err = b.atlas.CreateUser(atlas.User{Username: bindingID, Password: password})
	if err != nil {
		err = atlasToAPIError(err)
		return
	}

	// TODO: Place credentials in some sort of secrets manager.

	spec = brokerapi.Binding{
		Credentials: brokerapi.BrokerCredentials{
			Username: bindingID,
			Password: password,
		},
	}
	return
}

// Disconnect/unbind an application from an Atlas cluster
func (b *Broker) Unbind(ctx context.Context, instanceID string, bindingID string, details brokerapi.UnbindDetails, asyncAllowed bool) (brokerapi.UnbindSpec, error) {
	b.logger.Infof("Releasing binding \"%s\" for instance \"%s\" with details %+v", bindingID, instanceID, details)
	return brokerapi.UnbindSpec{}, nil
}

func (b *Broker) GetBinding(ctx context.Context, instanceID string, bindingID string) (spec brokerapi.GetBindingSpec, err error) {
	b.logger.Infof("Retrieving binding \"%s\" for instance \"%s\"", bindingID, instanceID)
	err = brokerapi.NewFailureResponse(fmt.Errorf("Unknown binding ID %s", bindingID), 404, "get-binding")
	return
}

// generatePassword will generate a fixed length password consisting of
// alphanumerical characters.
func generatePassword() string {
	const randomPasswordLength = 50
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

	password := make([]rune, randomPasswordLength)
	for i := range password {
		password[i] = chars[rand.Intn(len(chars))]
	}

	return string(password)
}
