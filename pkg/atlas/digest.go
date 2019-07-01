package atlas

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// digestParts will extract the parts neccessary to perform digest auth from
// the challenge response. These parts are the nonce, realm, and
// "quality of protection" (qop).
// These parts are stored in the "WWW-Authenticate" header as comma-separated
// key-value pairs on the form `key="value"`.
func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}

// getMD5 will calculate the MD5 hash of a string.
func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// getCnonce will generate a random nonce.
func getCnonce() string {
	b := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic("Error during cnonce generation: " + err.Error())
	}
	return fmt.Sprintf("%x", b)[:16]
}

// getDigestAuthrization generates a Authorization header value based on the
// parts from the server digest challenge.
// The generated header will comply with the format outlined in RFC 2617
// (https://tools.ietf.org/html/rfc2617).
func getDigestAuthrization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := "00000001"
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%s", qop="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response)
	return authorization
}
