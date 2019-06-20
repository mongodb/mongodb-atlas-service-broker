package broker

import (
	"context"
	"testing"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
)

func TestBind(t *testing.T) {
	broker, client := SetupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		ServiceID: "mongodb",
		PlanID:    "AWS-M10",
	}, true)

	bindingID := "binding"
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}

	user := client.Users[bindingID]
	if user == nil {
		t.Errorf("Expected user to exist with username %v", bindingID)
	}

	if user.Username != bindingID {
		t.Errorf("Expected created user to have username %v, got %v", bindingID, user.Username)
	}

	if user.Password == "" {
		t.Errorf("Expected password to have been generated")
	}
}

func TestBindAlreadyExisting(t *testing.T) {
	broker, _ := SetupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		ServiceID: "mongodb",
		PlanID:    "AWS-M10",
	}, true)

	bindingID := "binding"
	broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	if err != apiresponses.ErrBindingAlreadyExists {
		t.Errorf("Expected user already exists error, got %v", err)
	}
}

func TestBindMissingInstance(t *testing.T) {
	broker, _ := SetupTest()

	instanceID := "instance"
	bindingID := "binding"
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	if err != apiresponses.ErrInstanceDoesNotExist {
		t.Errorf("Expected instance does not exist error, got %v", err)
	}
}
