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
	"syscall"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia/store/memdb"

	"github.com/gaia-pipeline/gaia/security"

	"github.com/gaia-pipeline/gaia/services"

	"github.com/pkg/errors"

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
	err error
}

func (ms *mockScheduler) SetPipelineJobs(p *gaia.Pipeline) error { return ms.err }
func (ms *mockScheduler) GetFreeWorkers() int32                  { return int32(0) }
func (ms *mockScheduler) Init()                                  {}
func (ms *mockScheduler) SchedulePipeline(p *gaia.Pipeline, args []*gaia.Argument) (*gaia.PipelineRun, error) {
	return nil, ms.err
}
func (ms *mockScheduler) StopPipelineRun(p *gaia.Pipeline, runid int) error { return ms.err }
func (ms *mockScheduler) CountScheduledRuns() int                           { return 0 }

type mockStore struct {
	worker       *gaia.Worker
	run          *gaia.PipelineRun
	mockPipeline *gaia.Pipeline
	store.GaiaStore
}

func (m *mockStore) WorkerGetAll() ([]*gaia.Worker, error) {
	return []*gaia.Worker{{UniqueID: "test12345"}}, nil
}
func (m *mockStore) WorkerDeleteAll() error         { return nil }
func (m *mockStore) WorkerPut(w *gaia.Worker) error { m.worker = w; return nil }
func (m *mockStore) PipelinePutRun(r *gaia.PipelineRun) error {
	if r.ID == 1 {
		m.run = r
	}
	return nil
}
func (m *mockStore) PipelinePut(pipeline *gaia.Pipeline) error { return nil }
func (m *mockStore) PipelineGet(id int) (pipeline *gaia.Pipeline, err error) {
	return m.mockPipeline, nil
}
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

// PipelinePut is a Mock implementation for pipelines
func (m *mockStore) CreatePipelinePut(createPipeline *gaia.CreatePipeline) error {
	return nil
}

var lis *bufconn.Listener
var tmpFolder string

const bufsize = 1024 * 1024

type mockWorkerInterface struct {
	pbRuns  []*pb.PipelineRun
	gitRepo *pb.GitRepo
}

func (mw *mockWorkerInterface) GetGitRepo(context.Context, *pb.PipelineID) (*pb.GitRepo, error) {
	return mw.gitRepo, nil
}

var mW *mockWorkerInterface

func (mw *mockWorkerInterface) GetWork(workInst *pb.WorkerInstance, serv pb.Worker_GetWorkServer) error {
	pipelinePath := filepath.Join(tmpFolder, "my-pipeline_golang")

	// Create a mock pipeline file
	err := ioutil.WriteFile(pipelinePath, []byte("test pipeline content"), 0777)
	if err != nil {
		return err
	}

	// Get SHA-Sum from mock file
	sha, err := filehelper.GetSHA256Sum(pipelinePath)
	if err != nil {
		return err
	}

	// Create a mock pipeline file
	cppPipelinePath := filepath.Join(tmpFolder, "my-cpp-pipeline_cpp")
	err = ioutil.WriteFile(cppPipelinePath, []byte("test pipeline content"), 0777)
	if err != nil {
		return err
	}

	// Create broken test file
	cppPipelineBrokenPath := filepath.Join(tmpFolder, "my-cpp-pipeline-broken_cpp")
	err = ioutil.WriteFile(cppPipelineBrokenPath, []byte("tes pip cont"), 0777)
	if err != nil {
		return err
	}

	// Get SHA-Sum from mock file
	shaCpp, err := filehelper.GetSHA256Sum(cppPipelinePath)
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
		{
			UniqueId:     "second-pipeline-run",
			PipelineType: gaia.PTypeCpp.String(),
			Status:       string(gaia.RunScheduled),
			ScheduleDate: time.Now().Unix(),
			Id:           2,
			PipelineName: "my-cpp-pipeline-broken_cpp",
			ShaSum:       shaCpp,
		},
	}

	for _, run := range testdata {
		if workInst.UniqueId == "my-failed-worker" {
			return errors.New("worker is not registered")
		}

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
	if pipelineRun.PipelineName == "my-cpp-pipeline-broken_cpp" {
		content, err := ioutil.ReadFile(filepath.Join(tmpFolder, "my-cpp-pipeline_cpp"))
		if err != nil {
			return err
		}

		err = serv.Send(&pb.FileChunk{
			Chunk: content,
		})
		if err != nil {
			return err
		}
		return nil
	}

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
	if !bytes.Equal(content.Chunk, []byte("test log file entry")) {
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
	mW.gitRepo = &pb.GitRepo{
		Url:            "https://github.com/gaia-pipeline/go-example",
		SelectedBranch: "refs/heads/master",
	}
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
		mStore := &mockStore{}
		ag := InitAgent(&mockScheduler{}, mStore, tmp)

		// Run setup connection info
		clientTLS, err := ag.setupConnectionInfo()
		if err != nil {
			t.Fatal(err)
		}
		if clientTLS == nil {
			t.Fatal("clientTLS should be not nil")
		}

		// Validate worker object in mStore
		if mStore.worker.UniqueID != uniqueID {
			t.Fatalf("expected %s but got %s", uniqueID, mStore.worker.UniqueID)
		}
	})

	// Run setup configuration without registration
	t.Run("without-registration-success", func(t *testing.T) {
		// Init agent
		mStore := &mockStore{}
		ag := InitAgent(&mockScheduler{}, mStore, "./fixtures")

		// Run setup connection info
		clientTLS, err := ag.setupConnectionInfo()
		if err != nil {
			t.Fatal(err)
		}
		if clientTLS == nil {
			t.Fatal("clientTLS should be not nil")
		}

		// Validate worker object in mStore
		if mStore.worker.UniqueID != "test12345" {
			t.Fatalf("expected %s but got %s", uniqueID, mStore.worker.UniqueID)
		}
	})
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

type memDBFake struct {
	memdb.GaiaMemDB
	err  error
	pair gaia.SHAPair
	ok   bool
}

func (m *memDBFake) UpsertSHAPair(pair gaia.SHAPair) error {
	return m.err
}
func (m *memDBFake) GetSHAPair(pipelineID string) (bool, gaia.SHAPair, error) {
	return m.ok, m.pair, m.err
}

func TestScheduleWorkSHAPairMismatch(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	mDB := memDBFake{}
	mDB.ok = true
	mDB.pair = gaia.SHAPair{Original: []byte("test"), Worker: []byte("nottest")}
	services.MockMemDBService(&mDB)
	ag := InitAgent(mScheduler, mStore, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Run mScheduler
	ag.scheduleWork()

	// Validate output from mScheduler
	if mStore.run == nil {
		t.Fatal("run is nil but should exist")
	}
}

func TestRebuildWorkerBinaryUnkownPipeline(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	services.MockStorageService(mStore)
	ag := InitAgent(mScheduler, mStore, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})
	p := gaia.Pipeline{
		Name: "test-pipeline",
		ID:   1,
		UUID: security.GenerateRandomUUIDV5(),
		// Setting this to avoid testing CreatePipeline again.
		Type: gaia.PTypeUnknown,
	}
	err = ag.rebuildWorkerBinary(ctx, &p)
	if err == nil {
		t.Fatal("was expecting unknown pipeline type error... got none.")
	}
}

func TestRebuildWorkerBinary(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	services.MockStorageService(mStore)
	ms := new(mockScheduler)
	services.MockSchedulerService(ms)
	ag := InitAgent(mScheduler, mStore, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
		HomePath:     tmpFolder,
		CAPath:       tmpFolder,
		DataPath:     tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})
	p := gaia.Pipeline{
		Name: "test-pipeline",
		ID:   1,
		UUID: security.GenerateRandomUUIDV5(),
		// Setting this to avoid testing CreatePipeline again.
		Type: gaia.PTypeGolang,
	}
	err = ag.rebuildWorkerBinary(ctx, &p)
	if err != nil {
		t.Fatal("was not expecting error, got one: ", err)
	}
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
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	mDB := memDBFake{}
	services.MockMemDBService(&mDB)
	ag := InitAgent(mScheduler, mStore, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Run mScheduler
	ag.scheduleWork()

	// Validate output from mScheduler
	if mStore.run == nil {
		t.Fatal("run is nil but should exist")
	}
	if mStore.run.ID != 1 {
		t.Fatalf("expected 1 but got %d", mStore.run.ID)
	}
	if mStore.run.UniqueID != "first-pipeline-run" {
		t.Fatalf("expected 'first-pipeline-run' but got %s", mStore.run.UniqueID)
	}
	if len(mStore.run.Jobs) != 1 {
		t.Fatalf("expected 1 but got %d", len(mStore.run.Jobs))
	}
	if mStore.run.Jobs[0].Title != "Test job 1" {
		t.Fatalf("expected 'Test job 1' but got %s", mStore.run.Jobs[0].Title)
	}
	if len(mStore.run.Jobs[0].Args) != 1 {
		t.Fatalf("expected 1 but got %d", len(mStore.run.Jobs[0].Args))
	}
	if mStore.run.Jobs[0].Args[0].Key != "key" {
		t.Fatalf("expected 'key' but got %s", mStore.run.Jobs[0].Args[0].Key)
	}
}

func TestScheduleWorkExecFormatError(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Init agent
	mStore := &mockStore{}
	mStore.mockPipeline = &gaia.Pipeline{
		Name: "test-pipeline",
		ID:   1,
		UUID: security.GenerateRandomUUIDV5(),
		// Setting this to avoid testing CreatePipeline again.
		Type: gaia.PTypeUnknown,
	}
	mScheduler := &mockScheduler{}
	mScheduler.err = errors.New("exec format error")
	mDB := memDBFake{}
	services.MockMemDBService(&mDB)
	ag := InitAgent(mScheduler, mStore, "")
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
		HomePath:     tmpFolder,
		CAPath:       tmpFolder,
		DataPath:     tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Run mScheduler
	ag.scheduleWork()

	// Validate output from mScheduler
	if mStore.run != nil {
		t.Fatal("run should not exist.")
	}
}

func TestScheduleWork_RecvError(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)

	// Create test certificate files
	certFilePath := filepath.Join(tmpFolder, "cert.pem")
	if err := ioutil.WriteFile(certFilePath, []byte("test cert"), 0777); err != nil {
		t.Fatal(err)
	}
	keyFilePath := filepath.Join(tmpFolder, "key.pem")
	if err := ioutil.WriteFile(keyFilePath, []byte("test key"), 0777); err != nil {
		t.Fatal(err)
	}
	caCertFilePath := filepath.Join(tmpFolder, "caCert.pem")
	if err := ioutil.WriteFile(caCertFilePath, []byte("test ca cert"), 0777); err != nil {
		t.Fatal(err)
	}

	// Init agent
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	ag := InitAgent(mScheduler, mStore, tmpFolder)
	ag.sigs = make(chan os.Signal, 1)
	ag.client = client
	ag.self = &pb.WorkerInstance{UniqueId: "my-failed-worker"}
	gaia.Cfg = &gaia.Config{
		PipelinePath: tmpFolder,
	}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Gaia",
	})

	// Run mScheduler
	ag.scheduleWork()

	// Validate output from mScheduler
	select {
	case sig := <-ag.sigs:
		if sig != syscall.SIGTERM {
			t.Fatalf("expected SIGTERM syscall but got %#v", sig)
		}
	default:
		t.Fatal("signal channel is empty or blocked")
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
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	ag := InitAgent(mScheduler, mStore, "")
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
	mStore := &mockStore{}
	mScheduler := &mockScheduler{}
	ag := InitAgent(mScheduler, mStore, "")
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
	if err := os.MkdirAll(logFileFolder, 0700); err != nil {
		t.Fatal(err)
	}

	// Create log file
	if err := ioutil.WriteFile(logFilePath, []byte("test log file entry"), 0777); err != nil {
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
