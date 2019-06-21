package atlas

import "net/http"

// User represents a single Atlas database user.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type role struct {
	DatabaseName string `json:"databaseName"`
	RoleName     string `json:"roleName"`
}

type createUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
	Roles        []role `json:"roles"`
}

const (
	// The default database in Atlas is "admin"
	defaultDatabaseName = "admin"
	defaultRole         = "readWriteAnyDatabase"
)

// CreateUser will create a new database user with read/write access to all
// databases.
// Endpoint: POST /databaseUsers
func (c *HTTPClient) CreateUser(user User) (*User, error) {
	req := createUserRequest{
		Username:     user.Username,
		Password:     user.Password,
		DatabaseName: defaultDatabaseName,
		Roles: []role{
			role{
				DatabaseName: defaultDatabaseName,
				RoleName:     defaultRole,
			},
		},
	}

	err := c.request(http.MethodPost, "databaseUsers", req, nil)
	return &user, err
}
