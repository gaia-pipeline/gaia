package agent

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/agent/api"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	hclog "github.com/hashicorp/go-hclog"
)

type mockScheduler struct {
	scheduler.GaiaScheduler
}

type mockStore struct {
	worker *gaia.Worker
	store.GaiaStore
}

func (m *mockStore) WorkerGetAll() ([]*gaia.Worker, error) {
	return []*gaia.Worker{{UniqueID: "test12345"}}, nil
}
func (m *mockStore) WorkerDeleteAll() error         { return nil }
func (m *mockStore) WorkerPut(w *gaia.Worker) error { m.worker = w; return nil }

func TestInitAgent(t *testing.T) {
	ag := InitAgent(&mockScheduler{}, &mockStore{}, "")
	if ag == nil {
		t.Fatal("failed initiate agent")
	}
}

func TestSetupConnectionInfo(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestSetupConnectionInfo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	// setup test response data
	uniqueID := "unique-id"
	certBytes, err := ioutil.ReadFile("./fixtures/cert.pem")
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, err := ioutil.ReadFile("./fixtures/key.pem")
	if err != nil {
		t.Fatal(err)
	}
	caCertBytes, err := ioutil.ReadFile("./fixtures/cacert.pem")
	if err != nil {
		t.Fatal(err)
	}

	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Response
		resp := api.RegisterResponse{
			UniqueID: uniqueID,
			Cert:     base64.StdEncoding.EncodeToString(certBytes),
			Key:      base64.StdEncoding.EncodeToString(keyBytes),
			CACert:   base64.StdEncoding.EncodeToString(caCertBytes),
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

	// Set config
	gaia.Cfg = &gaia.Config{
		Logger:        hclog.NewNullLogger(),
		HomePath:      tmp,
		WorkerTags:    "tag1,tag2,tag3",
		WorkerHostURL: server.URL,
		WorkerSecret:  "secret12345",
	}

	// Run setup configuration with registration
	t.Run("registration-success", func(t *testing.T) {
		// Init agent
		store := &mockStore{}
		ag := InitAgent(&mockScheduler{}, store, tmp)

		// Run setup connection info
		clientTLS, err := ag.setupConnectionInfo()
		if err != nil {
			t.Fatal(err)
		}
		if clientTLS == nil {
			t.Fatal("clientTLS should be not nil")
		}

		// Validate worker object in store
		if store.worker.UniqueID != uniqueID {
			t.Fatalf("expected %s but got %s", uniqueID, store.worker.UniqueID)
		}
	})

	// Run setup configuration without registration
	t.Run("without-registration-success", func(t *testing.T) {
		// Init agent
		store := &mockStore{}
		ag := InitAgent(&mockScheduler{}, store, "./fixtures")

		// Run setup connection info
		clientTLS, err := ag.setupConnectionInfo()
		if err != nil {
			t.Fatal(err)
		}
		if clientTLS == nil {
			t.Fatal("clientTLS should be not nil")
		}

		// Validate worker object in store
		if store.worker.UniqueID != "test12345" {
			t.Fatalf("expected %s but got %s", uniqueID, store.worker.UniqueID)
		}
	})
}
