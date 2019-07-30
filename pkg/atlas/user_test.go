package atlas

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	expected := User{
		Username: "username",
		Database: "admin",
		LDAPType: "NONE",
		Roles: []Role{
			Role{
				Name:     "readWriteAnyDatabase",
				Database: "admin",
			},
		},
	}

	response := map[string]interface{}{
		"username":     "username",
		"databaseName": "admin",
		"ldapAuthType": "NONE",
		"roles": []map[string]interface{}{
			map[string]interface{}{
				"roleName":     "readWriteAnyDatabase",
				"databaseName": "admin",
			},
		},
	}

	atlas, server := setupTest(t, "/databaseUsers", http.MethodPost, 200, response)
	defer server.Close()

	user, err := atlas.CreateUser(expected)

	assert.NoError(t, err)
	assert.Equal(t, &expected, user)
}

func TestCreateUserExistingName(t *testing.T) {
	atlas, server := setupTest(t, "/databaseUsers", http.MethodPost, 400, errorResponse("USER_ALREADY_EXISTS"))
	defer server.Close()

	_, err := atlas.CreateUser(User{
		Username: "username",
	})

	assert.EqualError(t, err, ErrUserAlreadyExists.Error())
}

func TestGetUser(t *testing.T) {
	expected := &User{
		Username: "username",
	}

	response := map[string]interface{}{
		"username": "username",
	}

	atlas, server := setupTest(t, "/databaseUsers/admin/"+expected.Username, http.MethodGet, 200, response)
	defer server.Close()

	user, err := atlas.GetUser(expected.Username)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
}

func TestGetNonexistentUser(t *testing.T) {
	username := "username"
	atlas, server := setupTest(t, "/clusters/"+username, http.MethodGet, 404, errorResponse("USER_NOT_FOUND"))
	defer server.Close()

	_, err := atlas.GetCluster(username)
	assert.EqualError(t, err, ErrUserNotFound.Error())
}

func TestDeleteUser(t *testing.T) {
	username := "username"
	atlas, server := setupTest(t, "/databaseUsers/admin/"+username, http.MethodDelete, 200, nil)
	defer server.Close()

	err := atlas.DeleteUser(username)
	assert.NoError(t, err)
}

func TestDeleteNonexistentUser(t *testing.T) {
	username := "username"
	atlas, server := setupTest(t, "/databaseUsers/admin/"+username, http.MethodDelete, 404, errorResponse("USER_NOT_FOUND"))
	defer server.Close()

	err := atlas.DeleteUser(username)
	assert.EqualError(t, err, ErrUserNotFound.Error())
}
