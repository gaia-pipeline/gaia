package agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/filehelper"
	"github.com/gaia-pipeline/gaia/store"
	"github.com/gaia-pipeline/gaia/workers/agent/api"
	pb "github.com/gaia-pipeline/gaia/workers/proto"
	"github.com/gaia-pipeline/gaia/workers/scheduler"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type mockScheduler struct {
	scheduler.GaiaScheduler
}

func (ms *mockScheduler) SetPipelineJobs(p *gaia.Pipeline) error { return nil }
func (ms *mockScheduler) GetFreeWorkers() int32                  { return int32(0) }

type mockStore struct {
	worker *gaia.Worker
	run    *gaia.PipelineRun
	store.GaiaStore
}

func (m *mockStore) WorkerGetAll() ([]*gaia.Worker, error) {
	return []*gaia.Worker{{UniqueID: "test12345"}}, nil
}
func (m *mockStore) WorkerDeleteAll() error                                  { return nil }
func (m *mockStore) WorkerPut(w *gaia.Worker) error                          { m.worker = w; return nil }
func (m *mockStore) PipelinePutRun(r *gaia.PipelineRun) error                { m.run = r; return nil }
func (m *mockStore) PipelinePut(pipeline *gaia.Pipeline) error               { return nil }
func (m *mockStore) PipelineGet(id int) (pipeline *gaia.Pipeline, err error) { return nil, nil }
func (m *mockStore) PipelineGetAllRuns() ([]gaia.PipelineRun, error) {
	runs := []gaia.PipelineRun{
		{
			ID:         1,
			UniqueID:   "first-pipeline-run",
			PipelineID: 1,
			Jobs: []*gaia.Job{
				{
					ID:    1,
					Title: "first-job",
				},
				{
					ID:    2,
					Title: "second-job",
				},
			},
		},
		{
			ID:         2,
			UniqueID:   "second-pipeline-run",
			PipelineID: 2,
			Jobs: []*gaia.Job{
				{
					ID:    1,
					Title: "first-job",
				},
				{
					ID:    2,
					Title: "second-job",
				},
			},
		},
	}
	return runs, nil
}

var lis *bufconn.Listener
var tmpFolder string

const bufsize = 1024 * 1024

type mockWorkerInterface struct {
	pbRuns []*pb.PipelineRun
}

var mW *mockWorkerInterface

func (mw *mockWorkerInterface) GetWork(workInst *pb.WorkerInstance, serv pb.Worker_GetWorkServer) error {
	pipelinePath := filepath.Join(tmpFolder, "my-pipeline_golang")

	// Create a mock pipeline file
	err := ioutil.WriteFile(pipelinePath, []byte("test pipeline content"), 777)
	if err != nil {
		return err
	}

	// Get SHA-Sum from mock file
	sha, err := filehelper.GetSHA256Sum(pipelinePath)
	if err != nil {
		return err
	}

	testdata := []*pb.PipelineRun{
		{
			UniqueId:     "first-pipeline-run",
			PipelineType: gaia.PTypeGolang.String(),
			Status:       string(gaia.RunScheduled),
			ScheduleDate: time.Now().Unix(),
			Id:           1,
			PipelineName: "my-pipeline_golang",
			ShaSum:       sha,
			Jobs: []*pb.Job{
				{
					UniqueId:    1,
					Description: "Test job 1",
					Status:      string(gaia.JobWaitingExec),
					Title:       "Test job 1",
					Args: []*pb.Argument{
						{
							Description: "test argument",
							Key:         "key",
							Type:        "textbox",
						},
					},
				},
			},
		},
	}

	for _, run := range testdata {
		if err := serv.Send(run); err != nil {
			return err
		}
	}
	return nil
}

func (mw *mockWorkerInterface) UpdateWork(ctx context.Context, pipelineRun *pb.PipelineRun) (*empty.Empty, error) {
	mw.pbRuns = append(mw.pbRuns, pipelineRun)
	return &empty.Empty{}, nil
}

func (mw *mockWorkerInterface) StreamBinary(pipelineRun *pb.PipelineRun, serv pb.Worker_StreamBinaryServer) error {
	err := serv.Send(&pb.FileChunk{
		Chunk: []byte("test byte chunk\n"),
	})
	if err != nil {
		return err
	}

	err = serv.Send(&pb.FileChunk{
		Chunk: []byte("another byte chunk"),
	})
	return err
}

func (mw *mockWorkerInterface) StreamLogs(stream pb.Worker_StreamLogsServer) error {
	content, err := stream.Recv()
	if err != nil {
		return err
	}
	if bytes.Compare(content.Chunk, []byte("test log file entry")) != 0 {
		return fmt.Errorf("log file content is not the same: %s", string(content.Chunk[:]))
	}
	return nil
}

func (mw *mockWorkerInterface) Deregister(ctx context.Context, workInst *pb.WorkerInstance) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func init() {
	// Create tmp folder
	var err error
	if tmpFolder, err = ioutil.TempDir("", "AgentTestDir"); err != nil {
		log.Fatal(err)
	}

	lis = bufconn.Listen(bufsize)
	s := grpc.NewServer()
	mW = &mockWorkerInterface{}
	pb.RegisterWorkerServer(s, mW)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(tmpFolder)
	}()
}

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
	caCertBytes, err := ioutil.ReadFile("./fixtures/caCert.pem")
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

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

func TestScheduleWork(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	store := &mockStore{}
	scheduler := &mockScheduler{}
	ag := InitAgent(scheduler, store, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Run scheduler
	ag.scheduleWork()

	// Validate output from scheduler
	if store.run == nil {
		t.Fatal("run is nil but should be exist")
	}
	if store.run.ID != 1 {
		t.Fatalf("expected 1 but got %d", store.run.ID)
	}
	if store.run.UniqueID != "first-pipeline-run" {
		t.Fatalf("expected 'first-pipeline-run' but got %s", store.run.UniqueID)
	}
	if len(store.run.Jobs) != 1 {
		t.Fatalf("expected 1 but got %d", len(store.run.Jobs))
	}
	if store.run.Jobs[0].Title != "Test job 1" {
		t.Fatalf("expected 'Test job 1' but got %s", store.run.Jobs[0].Title)
	}
	if len(store.run.Jobs[0].Args) != 1 {
		t.Fatalf("expected 1 but got %d", len(store.run.Jobs[0].Args))
	}
	if store.run.Jobs[0].Args[0].Key != "key" {
		t.Fatalf("expected 'key' but got %s", store.run.Jobs[0].Args[0].Key)
	}
}

func TestStreamBinary(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	store := &mockStore{}
	scheduler := &mockScheduler{}
	ag := InitAgent(scheduler, store, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	run := &pb.PipelineRun{
		UniqueId: "first-pipeline-run",
	}
	pipelinePath := filepath.Join(tmpFolder, "test-pipeline")

	if err := ag.streamBinary(run, pipelinePath); err != nil {
		t.Fatal(err)
	}

	// Check content of file
	content, err := ioutil.ReadFile(pipelinePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content[:]) != "test byte chunk\nanother byte chunk" {
		t.Fatalf("wrong content in the streamed file: %s", string(content[:]))
	}
}

func TestUpdateWork(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	store := &mockStore{}
	scheduler := &mockScheduler{}
	ag := InitAgent(scheduler, store, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath:  tmpFolder,
		WorkspacePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Create test log folder
	logFileFolder := filepath.Join(gaia.Cfg.WorkspacePath, "1", "1", gaia.LogsFolderName)
	logFilePath := filepath.Join(logFileFolder, gaia.LogsFileName)
	if err := os.MkdirAll(logFileFolder, 500); err != nil {
		t.Fatal(err)
	}

	// Create log file
	if err := ioutil.WriteFile(logFilePath, []byte("test log file entry"), 777); err != nil {
		t.Fatal(err)
	}

	// Run update work
	ag.updateWork()

	// Verify the updated work
	if len(mW.pbRuns) != 2 {
		t.Fatalf("updated work should be 2 but is %d", len(mW.pbRuns))
	}
	if mW.pbRuns[0].UniqueId != "first-pipeline-run" {
		t.Fatalf("expected 'first-pipeline-run' but got %s", mW.pbRuns[0].UniqueId)
	}
	if len(mW.pbRuns[0].Jobs) != 2 {
		t.Fatalf("expected 2 but got %d", len(mW.pbRuns[0].Jobs))
	}
	if mW.pbRuns[1].UniqueId != "second-pipeline-run" {
		t.Fatalf("expected 'second-pipeline-run' but got %s", mW.pbRuns[1].UniqueId)
	}
	if len(mW.pbRuns[1].Jobs) != 2 {
		t.Fatalf("expected 2 but got %d", len(mW.pbRuns[1].Jobs))
	}
}
