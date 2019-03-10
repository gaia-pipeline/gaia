package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// RegisterResponse represents a response from API registration
// call.
type RegisterResponse struct {
	UniqueID string `json:"uniqueid"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"cacert"`
}

// RegisterWorker registers a new worker at a Gaia instance.
// It uses the given secret for authentication and returns certs
// which can be used for a future mTLS connection.
func RegisterWorker(host, secret string) (*RegisterResponse, error) {
	resp, err := http.PostForm(host,
		url.Values{"secret": {secret}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check the return code first
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to register worker. Return code was '%d' and message was: %s", resp.StatusCode, string(body))
	}

	// Unmarshal the json response
	regResp := RegisterResponse{}
	if err = json.Unmarshal(body, &regResp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal registration response: %s", string(body))
	}

	return &regResp, nil
}
