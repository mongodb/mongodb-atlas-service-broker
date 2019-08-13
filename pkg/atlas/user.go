package atlas

import (
	"fmt"
	"net/http"
)

// User represents a single Atlas database user.
type User struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
	LDAPAuthType string `json:"ldapAuthType,omitempty"`
	Roles        []Role `json:"roles,omitempty"`
}

// Role represents the role of a database user.
type Role struct {
	Name           string `json:"roleName"`
	DatabaseName   string `json:"databaseName,omitempty"`
	CollectionName string `json:"collectionName,omitempty"`
}

// CreateUser will create a new database user with read/write access to all
// databases.
// Endpoint: POST /databaseUsers
func (c *HTTPClient) CreateUser(user User) (*User, error) {
	// Atlas always uses "admin" for the authentication database.
	user.DatabaseName = "admin"

	var resultingUser User
	err := c.requestPublic(http.MethodPost, "databaseUsers", user, &resultingUser)
	return &resultingUser, err
}

// GetUser will find a database user by its username.
// GET /databaseUsers/admin/{USERNAME}
func (c *HTTPClient) GetUser(name string) (*User, error) {
	path := fmt.Sprintf("databaseUsers/admin/%s", name)

	var user User
	err := c.requestPublic(http.MethodGet, path, nil, &user)
	return &user, err
}

// DeleteUser will delete an existing database user.
// Endpoint: DELETE /databaseUsers/{USERNAME}
func (c *HTTPClient) DeleteUser(name string) error {
	path := fmt.Sprintf("databaseUsers/admin/%s", name)
	return c.requestPublic(http.MethodDelete, path, nil, nil)
}
