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
		fmt.Println(string(json))

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

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())

	// Decode the response from JSON
	json.NewDecoder(resp.Body).Decode(response)

	return nil
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
