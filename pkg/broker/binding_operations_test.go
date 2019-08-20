package broker

import (
	"testing"

	"github.com/mongodb/mongodb-atlas-service-broker/pkg/atlas"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"

	_, err := broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)

	user := client.Users[bindingID]
	assert.NotEmptyf(t, user, "Expected user to exist with username %s", bindingID)
	assert.Equal(t, bindingID, user.Username)
	assert.NotEmpty(t, user.Password, "Expected password to have been genereated")

	expectedRoles := []atlas.Role{
		atlas.Role{
			Name:         "readWriteAnyDatabase",
			DatabaseName: "admin",
		},
	}
	assert.Equal(t, expectedRoles, user.Roles, "Expected default role to have been assigned")
}

func TestBindParams(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	params := `{
		"user": {
			"ldapAuthType": "NONE",
			"roles": [{
				"roleName": "role",
				"databaseName": "database",
				"collectionName": "collection"
			}]
		}}`

	bindingID := "binding"
	_, err := broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:        testPlanID,
		ServiceID:     testServiceID,
		RawParameters: []byte(params),
	}, true)

	assert.NoError(t, err)

	user := client.Users[bindingID]
	assert.NotEmptyf(t, user, "Expected user to exist with username %s", bindingID)

	assert.Equal(t, bindingID, user.Username)
	assert.NotEmpty(t, user.Password, "Expected password to have been genereated")
	assert.Equal(t, "NONE", user.LDAPAuthType)

	expectedRoles := []atlas.Role{
		atlas.Role{
			Name:           "role",
			DatabaseName:   "database",
			CollectionName: "collection",
		},
	}
	assert.Equal(t, expectedRoles, user.Roles)
}

func TestBindAlreadyExisting(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)
	_, err := broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.EqualError(t, err, apiresponses.ErrBindingAlreadyExists.Error())
}

func TestBindMissingInstance(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	bindingID := "binding"
	_, err := broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}

func TestUnbind(t *testing.T) {
	broker, client, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	broker.Bind(ctx, instanceID, bindingID, brokerapi.BindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	_, err := broker.Unbind(ctx, instanceID, bindingID, brokerapi.UnbindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.NoError(t, err)
	assert.Empty(t, client.Users[bindingID], "Expected to be removed")
}

func TestUnbindMissing(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	bindingID := "binding"
	_, err := broker.Unbind(ctx, instanceID, bindingID, brokerapi.UnbindDetails{}, true)

	assert.EqualError(t, err, apiresponses.ErrBindingDoesNotExist.Error())
}

func TestUnbindMissingInstance(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	bindingID := "binding"
	_, err := broker.Unbind(ctx, instanceID, bindingID, brokerapi.UnbindDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	assert.EqualError(t, err, apiresponses.ErrInstanceDoesNotExist.Error())
}
