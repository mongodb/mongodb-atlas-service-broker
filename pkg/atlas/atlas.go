package atlas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is an interface for interacting with the Atlas API.
type Client interface {
	CreateCluster(cluster Cluster) (*Cluster, error)
	UpdateCluster(cluster Cluster) (*Cluster, error)
	DeleteCluster(name string) error
	GetCluster(name string) (*Cluster, error)
	GetDashboardURL(clusterName string) string

	CreateUser(user User) (*User, error)
	GetUser(name string) (*User, error)
	DeleteUser(name string) error

	GetProvider(name string) (*Provider, error)
}

// HTTPClient is the main implementation of the Client interface which
// communicates with the Atlas API.
type HTTPClient struct {
	BaseURL    string
	GroupID    string
	PublicKey  string
	PrivateKey string

	HTTP *http.Client
}

// Different errors the api may return.
var (
	ErrPlanIDNotFound = errors.New("plan-id not in the catalog")

	ErrUnauthorized = errors.New("Invalid API key")

	ErrClusterNotFound      = errors.New("Cluster not found")
	ErrClusterAlreadyExists = errors.New("Cluster already exists")

	ErrUserNotFound      = errors.New("User not found")
	ErrUserAlreadyExists = errors.New("User already exists")
)

const (
	publicAPIPath  = "/api/atlas/v1.0"
	privateAPIPath = "/api/private/unauth"
)

// NewClient will create a new HTTPClient with the specified connection details.
func NewClient(baseURL string, groupID string, publicKey string, privateKey string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		GroupID:    groupID,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		HTTP:       &http.Client{},
	}
}

// requestPublic will make a request to an endpoint in the public API.
// The URL will be constructed by prepending the group to the specified endpoint.
func (c *HTTPClient) requestPublic(method string, endpoint string, body interface{}, response interface{}) error {
	url := fmt.Sprintf("%s%s/groups/%s/%s", c.BaseURL, publicAPIPath, c.GroupID, endpoint)
	return c.request(method, url, body, response)
}

// requestPrivate will make a request to an endpoint in the private API.
func (c *HTTPClient) requestPrivate(method string, endpoint string, body interface{}, response interface{}) error {
	url := fmt.Sprintf("%s%s/%s", c.BaseURL, privateAPIPath, endpoint)
	return c.request(method, url, body, response)
}

// request makes an HTTP request using the specified method.
// If body is passed it will be JSON encoded and included with the request.
// If the request was successful the response will be decoded into response.
func (c *HTTPClient) request(method string, url string, body interface{}, response interface{}) error {
	var data io.Reader

	// Construct the JSON payload if a body has been passed
	if body != nil {
		json, err := json.Marshal(body)
		if err != nil {
			return err
		}

		data = bytes.NewBuffer(json)
	}

	// Prepare API request.
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}

	// Perform digest authentication to retrieve single-use credentials.
	auth, err := c.digestAuth(method, url)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	req.Header.Set("Content-Type", "application/json")

	// Perform HTTP request.
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response if request was successful.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {

		if response != nil {
			err = json.NewDecoder(resp.Body).Decode(response)

			// EOF error means the response body was empty.
			if err != io.EOF {
				return err
			}
		}

		return nil
	}

	// Invalid credentials will cause a 401 Unauthorized response.
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	// Decode error if request was unsuccessful.
	var errorResponse struct {
		Code        string `json:"errorCode"`
		Description string `json:"detail"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return err
	}

	return errorFromErrorCode(errorResponse.Code, errorResponse.Description)
}

// digestAuth performs an unauthenticated request to retrieve a digest nonce.
// It returns the full authentication header constructed from the server response.
func (c *HTTPClient) digestAuth(method string, endpoint string) (string, error) {
	authReq, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTP.Do(authReq)
	if err != nil {
		return "", err
	}

	parts := digestParts(resp)
	parts["method"] = method

	// Retrieve URI from the full URL
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	parts["uri"] = endpointURL.RequestURI()

	// User public and private key as username and password
	parts["username"] = c.PublicKey
	parts["password"] = c.PrivateKey

	return getDigestAuthrization(parts), nil
}

// errorFromErrorCode converts an Atlas API error code into an error.
func errorFromErrorCode(code string, description string) error {
	errorsByCode := map[string]error{
		"CLUSTER_NOT_FOUND":                  ErrClusterNotFound,
		"CLUSTER_ALREADY_REQUESTED_DELETION": ErrClusterNotFound,

		"DUPLICATE_CLUSTER_NAME": ErrClusterAlreadyExists,

		"USER_ALREADY_EXISTS": ErrUserAlreadyExists,
		"USER_NOT_FOUND":      ErrUserNotFound,
	}

	// Default to an error wrapping the Atlas error description.
	err := errorsByCode[code]
	if err == nil {
		return fmt.Errorf("atlas error: [%s] %s", code, description)
	}

	return err
}
