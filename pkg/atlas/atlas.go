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

type Client interface {
	CreateCluster(cluster Cluster) (*Cluster, error)
	TerminateCluster(name string) error
	GetCluster(name string) (*Cluster, error)

	CreateUser(user User) (*User, error)
}

// HTTPClient is the main implementation of the Client interface which
// communicates with the Atlas API.
type HTTPClient struct {
	baseUrl    string
	groupID    string
	publicKey  string
	privateKey string
}

var (
	ErrClusterNotFound      = errors.New("Cluster not found")
	ErrClusterAlreadyExists = errors.New("Cluster already exists")

	ErrUserNotFound      = errors.New("User not found")
	ErrUserAlreadyExists = errors.New("User already exists")
)

func NewClient(baseUrl string, groupID string, publicKey string, privateKey string) (*HTTPClient, error) {
	return &HTTPClient{
		baseUrl:    baseUrl,
		groupID:    groupID,
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

func (c *HTTPClient) url(path string) string {
	return fmt.Sprintf("%s/groups/%s/%s", c.baseUrl, c.groupID, path)
}

func (c *HTTPClient) request(method string, path string, body interface{}, response interface{}) error {
	var data io.Reader

	// Construct the JSON payload if a body has been passed
	if body != nil {
		json, err := json.Marshal(body)
		if err != nil {
			return err
		}

		data = bytes.NewBuffer(json)
	}

	url := c.url(path)

	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}

	// Perform digest authentication to retrieve single-use credentials
	auth, err := c.digestAuth(method, url)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	req.Header.Set("Content-Type", "application/json")

	// Perform HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response if request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.NewDecoder(resp.Body).Decode(response)
		return nil
	}

	// Decode error if request was unsuccessful
	var errorResponse struct {
		Code        string `json:"errorCode"`
		Description string `json:"detail"`
	}
	json.NewDecoder(resp.Body).Decode(&errorResponse)

	return errorFromErrorCode(errorResponse.Code, errorResponse.Description)
}

// digestAuth performs an unauthenticated request to retrieve a digest nonce.
// It returns the full authentication header constructed from the server response.
func (c *HTTPClient) digestAuth(method string, endpoint string) (string, error) {
	client := &http.Client{}

	authReq, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(authReq)
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
	parts["username"] = c.publicKey
	parts["password"] = c.privateKey

	return getDigestAuthrization(parts), nil
}

// errorFromErrorCode converts an Atlas API error code into an error
func errorFromErrorCode(code string, description string) error {
	errorsByCode := map[string]error{
		"CLUSTER_NOT_FOUND":                  ErrClusterNotFound,
		"CLUSTER_ALREADY_REQUESTED_DELETION": ErrClusterNotFound,

		"DUPLICATE_CLUSTER_NAME": ErrClusterAlreadyExists,

		"USER_ALREADY_EXISTS": ErrUserAlreadyExists,
	}

	// Default to an error wrapping the Atlas error description
	err := errorsByCode[code]
	if err == nil {
		return errors.New("Atlas error: " + description)
	}

	return err
}
