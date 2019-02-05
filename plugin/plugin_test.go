package plugin

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gaia-pipeline/gaia"
	proto "github.com/gaia-pipeline/protobuf"
	hclog "github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/metadata"
)

type fakeCAAPI struct{}

func (c *fakeCAAPI) CreateSignedCert() (string, string, error) { return "", "", nil }
func (c *fakeCAAPI) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	return &tls.Config{}, nil
}
func (c *fakeCAAPI) CleanupCerts(crt, key string) error { return nil }
func (c *fakeCAAPI) GetCACertPath() (string, string)    { return "", "" }

type fakeClientProtocol struct{}

func (cp *fakeClientProtocol) Dispense(s string) (interface{}, error) { return &fakePluginGRPC{}, nil }
func (cp *fakeClientProtocol) Ping() error                            { return nil }
func (cp *fakeClientProtocol) Close() error                           { return nil }

type fakePluginGRPC struct{}

func (p *fakePluginGRPC) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return &fakeJobsClient{}, nil
}
func (p *fakePluginGRPC) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return &proto.JobResult{}, nil
}

type fakeJobsClient struct {
	counter int
}

func (jc *fakeJobsClient) Recv() (*proto.Job, error) {
	if jc.counter == 0 {
		jc.counter++
		return &proto.Job{}, nil
	}
	return nil, io.EOF
}
func (jc *fakeJobsClient) Header() (metadata.MD, error) { return nil, nil }
func (jc *fakeJobsClient) Trailer() metadata.MD         { return nil }
func (jc *fakeJobsClient) CloseSend() error             { return nil }
func (jc *fakeJobsClient) Context() context.Context     { return nil }
func (jc *fakeJobsClient) SendMsg(m interface{}) error  { return nil }
func (jc *fakeJobsClient) RecvMsg(m interface{}) error  { return nil }

func TestNewPlugin(t *testing.T) {
	p := &Plugin{}
	p.NewPlugin(new(fakeCAAPI))
}

func TestInit(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	emptyPlugin := &Plugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	err := p.Init(exec.Command("echo", "world"), &logpath)
	if !strings.Contains(err.Error(), "Unrecognized remote plugin message") {
		// Sometimes go-plugin throws this error instead...
		if !strings.Contains(err.Error(), "plugin exited before we could connect") {
			t.Fatalf("Error should contain 'Unrecognized remote plugin message' but was '%s'", err.Error())
		}
	}
}

func TestValidate(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	p := &Plugin{clientProtocol: new(fakeClientProtocol)}
	err := p.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecute(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	p := &Plugin{pluginConn: new(fakePluginGRPC)}
	buf := new(bytes.Buffer)
	p.writer = bufio.NewWriter(buf)
	err := p.Execute(&gaia.Job{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetJobs(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	p := &Plugin{pluginConn: new(fakePluginGRPC)}
	buf := new(bytes.Buffer)
	p.writer = bufio.NewWriter(buf)
	_, err := p.GetJobs()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRebuildDepTree(t *testing.T) {
	l := []gaia.Job{
		{ID: 12345},
		{ID: 1234},
		{ID: 123},
	}
	dep := []uint32{1234, 123}
	depTree := rebuildDepTree(dep, l)
	if len(depTree) != 2 {
		t.Fatalf("dependency length should be 2 but is %d", len(depTree))
	}
	for _, depJob := range depTree {
		if depJob.ID != 1234 && depJob.ID != 123 {
			t.Fatalf("wrong dependency detected %d", depJob.ID)
		}
	}
}

func TestClose(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	emptyPlugin := &Plugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	p.Init(exec.Command("echo", "world"), &logpath)
	p.Close()
}

func TestFlushLogs(t *testing.T) {
	gaia.Cfg = &gaia.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})
	emptyPlugin := &Plugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	p.Init(exec.Command("echo", "world"), &logpath)
	err := p.FlushLogs()
	if err != nil {
		t.Fatal(err)
	}
}
