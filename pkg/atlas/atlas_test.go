package atlas

import (
	"encoding/json"
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T, expectedPath string, method string, status int, response interface{}) *HTTPClient {
	const groupID = "group"
	const publicKey = "pubkey"
	const privateKey = "privkey"

	fullPath := fmt.Sprintf("/groups/%s%s", groupID, expectedPath)

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

	atlas, err := NewClient(s.URL, groupID, publicKey, privateKey)
	atlas.HTTP = s.Client()

	assert.NoError(t, err)

	return atlas
}

func errorResponse(code string) interface{} {
	return struct {
		Code string `json:"errorCode"`
	}{
		code,
	}
}
