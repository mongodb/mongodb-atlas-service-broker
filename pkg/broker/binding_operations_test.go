package broker

import (
	"context"
	"testing"

	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	assert.NoError(t, err)

	user := client.Users[bindingID]
	assert.NotEmptyf(t, user, "Expected user to exist with username %s", bindingID)
	assert.Equal(t, bindingID, user.Username)
	assert.NotEmpty(t, user.Password, "Expected password to have been genereated")
}

func TestBindAlreadyExisting(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrBindingAlreadyExists.Error())
}

func TestBindMissingInstance(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	bindingID := "binding"
	_, err := broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}

func TestUnbind(t *testing.T) {
	broker, client := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	broker.Bind(context.Background(), instanceID, bindingID, brokerapi.BindDetails{}, true)

	_, err := broker.Unbind(context.Background(), instanceID, bindingID, brokerapi.UnbindDetails{}, true)

	assert.NoError(t, err)
	assert.Empty(t, client.Users[bindingID], "Expected to be removed")
}

func TestUnbindMissing(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	broker.Provision(context.Background(), instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	_, err := broker.Unbind(context.Background(), instanceID, bindingID, brokerapi.UnbindDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrBindingDoesNotExist.Error())
}

func TestUnbindMissingInstance(t *testing.T) {
	broker, _ := setupTest()

	instanceID := "instance"
	bindingID := "binding"
	_, err := broker.Unbind(context.Background(), instanceID, bindingID, brokerapi.UnbindDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}
