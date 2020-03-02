package broker

import (
	"net/url"
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

func TestBindConnectionString(t *testing.T) {
	broker, _, ctx := setupTest()

	instanceID := "instance"
	broker.Provision(ctx, instanceID, brokerapi.ProvisionDetails{
		PlanID:    testPlanID,
		ServiceID: testServiceID,
	}, true)

	type inputs struct {
		details brokerapi.BindDetails
	}
	type outputs struct {
		scheme         string
		path           string
		queryString    string
		hasCredentials bool
	}

	tests := []struct {
		name    string
		inputs  inputs
		outputs outputs
	}{
		{
			name:   "no_connection_string",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID}},
			outputs: outputs{
				scheme:         "mongodb+srv",
				hasCredentials: true,
			},
		},
		{
			name: "empty_connection_string",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb+srv",
				hasCredentials: true,
			},
		},
		{
			name: "connection_string_without_credentials",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {
						"skipCredentials": true
					}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb+srv",
				hasCredentials: false,
			},
		},
		{
			name: "connection_string_with_database",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {
						"database": "atlas"
					}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb+srv",
				path:           "/atlas",
				hasCredentials: true,
			},
		},
		{
			name: "connection_string_with_options",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {
						"options": {
							"connectTimeoutMS": 1000,
							"tlsAllowInvalidCertificates": true,
							"w": "majority"
						}
					}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb+srv",
				path:           "/",
				queryString:    "connectTimeoutMS=1000&tlsAllowInvalidCertificates=true&w=majority",
				hasCredentials: true,
			},
		},
		{
			name: "connection_string_with_all_params",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {
						"skipCredentials": true,
						"database": "atlas",
						"options": {
							"connectTimeoutMS": 1000,
							"tlsAllowInvalidCertificates": true,
							"w": "majority"
						}
					}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb+srv",
				path:           "/atlas",
				queryString:    "connectTimeoutMS=1000&tlsAllowInvalidCertificates=true&w=majority",
				hasCredentials: false,
			},
		},
		{
			name: "connection_string_with_standard_format",
			inputs: inputs{brokerapi.BindDetails{PlanID: testPlanID, ServiceID: testServiceID, RawParameters: []byte(`{
					"connectionString": {
						"database": "atlas",
						"format": "standard",
						"options": {
							"connectTimeoutMS": 1000
						}
					}
				}`)},
			},
			outputs: outputs{
				scheme:         "mongodb",
				path:           "/atlas",
				queryString:    "authSource=admin&connectTimeoutMS=1000&replicaSet=shard&ssl=true",
				hasCredentials: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := broker.Bind(ctx, instanceID, tt.name, tt.inputs.details, true)
			assert.NoError(t, err)
			if c, ok := b.Credentials.(ConnectionDetails); ok {
				t.Log(c.ConnectionString)
				u, err := url.Parse(c.ConnectionString)
				assert.NoError(t, err)
				assert.Equal(t, tt.outputs.scheme, u.Scheme)
				assert.Equal(t, tt.outputs.path, u.Path)
				assert.Equal(t, tt.outputs.queryString, u.RawQuery)
				if tt.outputs.hasCredentials {
					assert.NotEmpty(t, u.User.String())
				} else {
					assert.Empty(t, u.User.String())
				}
			}
		})
	}
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
