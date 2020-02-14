package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	hclog "github.com/hashicorp/go-hclog"
)

func TestStart(t *testing.T) {
	// Create tmp folder
	tmpFolder, err := ioutil.TempDir("", "TestStart")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpFolder)

	gaia.Cfg = &gaia.Config{
		Mode:              gaia.ModeServer,
		WorkerGRPCHostURL: "myhost:12345",
		HomePath:          tmpFolder,
		DataPath:          tmpFolder,
		CAPath:            tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Init worker server
	server := InitWorkerServer()

	// Start server
	errChan := make(chan error)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()
	time.Sleep(3 * time.Second)
	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}
}
