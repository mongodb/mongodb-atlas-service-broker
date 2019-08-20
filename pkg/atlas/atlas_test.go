package atlas

import (
	"encoding/json"
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

// setupTest will set up an Atlas client with a mock HTTP client. The HTTP
// client will use a mock HTTP server which only responds to the specified path
// and the specified method. The HTTP server will simulate the digest
// authentication and return the specified status and response.
func setupTest(t *testing.T, expectedPath string, method string, status int, response interface{}) (*HTTPClient, *httptest.Server) {
	const groupID = "group"
	const publicKey = "pubkey"
	const privateKey = "privkey"

	fullPath := fmt.Sprintf("%s/groups/%s%s", publicAPIPath, groupID, expectedPath)

	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fullPath, req.URL.String())
		assert.Equal(t, method, req.Method)

		// If auth header is missing we return 401 to trigger the digest process
		if len(req.Header["Authorization"]) == 0 {
			rw.WriteHeader(401)
			return
		}

		rw.WriteHeader(status)

		if response != nil {
			data, _ := json.Marshal(response)
			rw.Write(data)
		} else {
			rw.Write([]byte{})
		}
	}))

	atlas := NewClient(s.URL, groupID, publicKey, privateKey)
	atlas.HTTP = s.Client()

	return atlas, s
}

func errorResponse(code string) interface{} {
	return struct {
		Code string `json:"errorCode"`
	}{
		code,
	}
}
