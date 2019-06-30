package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/stringhelper"
)

func TestRegisterWorker(t *testing.T) {
	// Define returned data
	uniqueID := "my-unique-id"
	cert := "test-cert"
	key := "test-key"
	caCert := "test-cacert"

	// Define test data
	name := "my-worker"
	secret := "12345-test-secret"
	tags := []string{"tag1", "tag2", "tag3"}

	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Check request parameters
		if req.URL.String() != fmt.Sprintf("/api/%s/worker/register", gaia.APIVersion) {
			t.Fatalf("wrong request parameters provided: %s", req.URL.String())
		}

		// Check form values
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if req.Form.Get("name") != name {
			t.Fatalf("expected %s but got %s", name, req.Form.Get("name"))
		}
		if req.Form.Get("secret") != secret {
			t.Fatalf("expected %s but got %s", secret, req.Form.Get("secret"))
		}
		reqTags := req.Form["tags"]
		for _, tag := range tags {
			if !stringhelper.IsContainedInSlice(reqTags, tag, false) {
				t.Fatalf("expected tag %s to be in slice but it is not: %s", tag, reqTags)
			}
		}

		// Response
		resp := RegisterResponse{
			UniqueID: uniqueID,
			Cert:     cert,
			Key:      key,
			CACert:   caCert,
		}

		// Marshal
		mResp, err := json.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}

		// Return response
		if _, err := rw.Write(mResp); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	// Run call
	resp, err := RegisterWorker(server.URL, secret, name, tags)
	if err != nil {
		t.Fatal(err)
	}

	// Validate returned data
	if resp.UniqueID != uniqueID {
		t.Fatalf("expected %s but got %s", uniqueID, resp.UniqueID)
	}
	if resp.Cert != cert {
		t.Fatalf("expected %s but got %s", cert, resp.Cert)
	}
	if resp.Key != key {
		t.Fatalf("expected %s but got %s", key, resp.Key)
	}
	if resp.CACert != caCert {
		t.Fatalf("expected %s but got %s", caCert, resp.CACert)
	}
}
