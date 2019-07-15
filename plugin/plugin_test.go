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
	"time"

	"github.com/gaia-pipeline/gaia"
	proto "github.com/gaia-pipeline/protobuf"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/metadata"
)

type fakeCAAPI struct{}

func (c *fakeCAAPI) CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error) {
	return "", "", nil
}
func (c *fakeCAAPI) CreateSignedCert() (string, string, error) { return "", "", nil }
func (c *fakeCAAPI) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	return &tls.Config{}, nil
}
func (c *fakeCAAPI) CleanupCerts(crt, key string) error { return nil }
func (c *fakeCAAPI) GetCACertPath() (string, string)    { return "", "" }

type fakeClientProtocol struct{}

func (cp *fakeClientProtocol) Dispense(s string) (interface{}, error) { return &fakeGaiaPlugin{}, nil }
func (cp *fakeClientProtocol) Ping() error                            { return nil }
func (cp *fakeClientProtocol) Close() error                           { return nil }

type fakeGaiaPlugin struct{}

func (p *fakeGaiaPlugin) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return &fakeJobsClient{}, nil
}
func (p *fakeGaiaPlugin) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return &proto.JobResult{}, nil
}

type fakeJobsClient struct {
	counter int
}

func (jc *fakeJobsClient) Recv() (*proto.Job, error) {
	j := &proto.Job{
		Args: []*proto.Argument{
			{
				Key:   "key",
				Value: "value",
			},
		},
	}

	if jc.counter == 0 {
		jc.counter++
		return j, nil
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
	p := &GoPlugin{}
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
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	err := p.Init(exec.Command("echo", "world"), &logpath)
	if err == nil {
		t.Fatal("was expecting an error. non happened")
	}
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
	p := &GoPlugin{clientProtocol: new(fakeClientProtocol)}
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
	p := &GoPlugin{pluginConn: new(fakeGaiaPlugin)}
	buf := new(bytes.Buffer)
	p.writer = bufio.NewWriter(buf)
	j := &gaia.Job{
		Args: []*gaia.Argument{
			{
				Key:   "key",
				Value: "value",
			},
		},
	}
	err := p.Execute(j)
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
	p := &GoPlugin{pluginConn: new(fakeGaiaPlugin)}
	buf := new(bytes.Buffer)
	p.writer = bufio.NewWriter(buf)
	_, err := p.GetJobs()
	if err != nil {
		t.Fatal(err)
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
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	_ = p.Init(exec.Command("echo", "world"), &logpath)
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
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	_ = p.Init(exec.Command("echo", "world"), &logpath)
	err := p.FlushLogs()
	if err != nil {
		t.Fatal(err)
	}
}
