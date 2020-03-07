package plugin

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	proto "github.com/Skarlso/protobuf"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/hashicorp/go-plugin"
)

const (
	pluginMapKey = "Plugin"

	// env variable key names for TLS cert path
	serverCertEnv = "GAIA_PLUGIN_CERT"
	serverKeyEnv  = "GAIA_PLUGIN_KEY"
	rootCACertEnv = "GAIA_PLUGIN_CA_CERT"
)

var handshake = plugin.HandshakeConfig{
	ProtocolVersion: 2,
	MagicCookieKey:  "GAIA_PLUGIN",
	// This cookie should never be changed again
	MagicCookieValue: "FdXjW27mN6XuG2zDBP4LixXUwDAGCEkidxwqBGYpUhxiWHzctATYZvpz4ZJdALmh",
}

var pluginMap = map[string]plugin.Plugin{
	pluginMapKey: &GaiaPluginImpl{},
}

// timeFormat is the logging time format.
const timeFormat = "2006/01/02 15:04:05"

// GaiaLogWriter represents a concurrent safe log writer which can be shared with go-plugin.
type GaiaLogWriter struct {
	mu     sync.RWMutex
	buffer *bytes.Buffer
	writer *bufio.Writer
}

// Write locks and writes to the underlying writer.
func (g *GaiaLogWriter) Write(p []byte) (n int, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.Write(p)
}

// Flush locks and flushes the underlying writer.
func (g *GaiaLogWriter) Flush() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.Flush()
}

// WriteString locks and passes on the string to write to the underlying writer.
func (g *GaiaLogWriter) WriteString(s string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.WriteString(s)
}

// GoPlugin represents a single plugin instance which uses gRPC
// to connect to exactly one plugin.
type GoPlugin struct {
	// Client is an instance of the go-plugin client.
	client *plugin.Client

	// Client protocol instance used to open gRPC connections.
	clientProtocol plugin.ClientProtocol

	// Interface to the connected plugin.
	pluginConn GaiaPlugin

	// Log file where all output is stored.
	logFile *os.File

	// Writer used to write logs from execution to file or buffer
	logger GaiaLogWriter

	// CA instance used to handle certificates
	ca security.CAAPI

	// Created certificates path for pipeline run
	certPath       string
	keyPath        string
	serverCertPath string
	serverKeyPath  string
}

// Plugin represents the plugin implementation.
type Plugin interface {
	// NewPlugin creates a new instance of plugin
	NewPlugin(ca security.CAAPI) Plugin

	// Init initializes the go-plugin client and generates a
	// new certificate pair for gaia and the plugin/pipeline.
	Init(command *exec.Cmd, logPath *string) error

	// Validate validates the plugin interface.
	Validate() error

	// Execute executes one job of a pipeline.
	Execute(j *gaia.Job) error

	// GetJobs returns all real jobs from the pipeline.
	GetJobs() ([]*gaia.Job, error)

	// FlushLogs flushes the logs.
	FlushLogs() error

	// Close closes the connection and cleans open file writes.
	Close()
}

// NewPlugin creates a new instance of Plugin.
// One Plugin instance represents one connection to a plugin.
func (p *GoPlugin) NewPlugin(ca security.CAAPI) Plugin {
	return &GoPlugin{ca: ca}
}

// Init prepares the log path, set's up new certificates for both gaia and
// plugin, and prepares the go-plugin client.
//
// It expects the start command for the plugin and the path where
// the log file should be stored.
//
// It's up to the caller to call plugin.Close to shutdown the plugin
// and close the gRPC connection.
func (p *GoPlugin) Init(command *exec.Cmd, logPath *string) error {
	// Initialise the logger
	p.logger = GaiaLogWriter{}

	// Create log file and open it.
	// We will close this file in the close method.
	if logPath != nil {
		var err error
		p.logFile, err = os.OpenFile(
			*logPath,
			os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return err
		}

		// Create new writer
		p.logger.writer = bufio.NewWriter(p.logFile)
	} else {
		// If no path is provided, write output to buffer
		p.logger.buffer = new(bytes.Buffer)
		p.logger.writer = bufio.NewWriter(p.logger.buffer)
	}

	// Create and sign a new pair of certificates for the server
	var err error
	p.serverCertPath, p.serverKeyPath, err = p.ca.CreateSignedCert()
	if err != nil {
		return err
	}

	// Expose path of server certificates as well as public CA cert.
	// This allows the plugin to grab the certificates.
	caCert, _ := p.ca.GetCACertPath()
	command.Env = append(command.Env, serverCertEnv+"="+p.serverCertPath)
	command.Env = append(command.Env, serverKeyEnv+"="+p.serverKeyPath)
	command.Env = append(command.Env, rootCACertEnv+"="+caCert)

	// Create and sign a new pair of certificates for the client
	p.certPath, p.keyPath, err = p.ca.CreateSignedCert()
	if err != nil {
		return err
	}

	// Generate TLS config
	tlsConfig, err := p.ca.GenerateTLSConfig(p.certPath, p.keyPath)
	if err != nil {
		return err
	}

	// Get new client
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshake,
		Plugins:          pluginMap,
		Cmd:              command,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Stderr:           &p.logger,
		TLSConfig:        tlsConfig,
	})

	// Connect via gRPC
	p.clientProtocol, err = p.client.Client()
	if err != nil {
		_ = p.logger.Flush()
		return fmt.Errorf("%s\n\n--- output ---\n%s", err.Error(), p.logger.buffer.String())
	}

	return nil
}

// Validate validates the interface of the plugin.
func (p *GoPlugin) Validate() error {
	// Request the plugin
	raw, err := p.clientProtocol.Dispense(pluginMapKey)
	if err != nil {
		return err
	}

	// Convert plugin to interface
	if pC, ok := raw.(GaiaPlugin); ok {
		p.pluginConn = pC
		return nil
	}

	return errors.New("plugin is not compatible with plugin interface")
}

// Execute triggers the execution of one single job
// for the given plugin.
func (p *GoPlugin) Execute(j *gaia.Job) error {
	// Transform arguments
	var args []*proto.Argument
	for _, arg := range j.Args {
		a := &proto.Argument{
			Key:   arg.Key,
			Value: arg.Value,
		}

		args = append(args, a)
	}

	// Create new proto job object.
	job := &proto.Job{
		UniqueId: j.ID,
		Args:     args,
	}

	// Execute the job
	resultObj, err := p.pluginConn.ExecuteJob(job)

	// Check and set job status
	if resultObj != nil && resultObj.ExitPipeline {
		// ExitPipeline is true that indicates that the job failed.
		j.Status = gaia.JobFailed

		// Failed was set so the pipeline will now be marked as failed.
		if resultObj.Failed {
			j.FailPipeline = true
		}

		// Generate error message and attach it to logs.
		timeString := time.Now().Format(timeFormat)
		_, _ = p.logger.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, resultObj.Message))
	} else if err != nil {
		// An error occurred during the send or somewhere else.
		// The job itself usually does not return an error here.
		// We mark the job as failed.
		j.Status = gaia.JobFailed

		// Generate error message and attach it to logs.
		timeString := time.Now().Format(timeFormat)
		_, _ = p.logger.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, err.Error()))
	} else {
		// We set up the job's output if there was any
		outs := resultObj.GetOutput()
		if outs != nil {
			o := make([]*gaia.Output, 0)
			for _, out := range outs {
				o = append(o, &gaia.Output{
					Key:   out.GetKey(),
					Value: out.GetValue(),
				})
			}
			j.Outs = o
		}
		j.Status = gaia.JobSuccess
	}

	return nil
}

// GetJobs receives all implemented jobs from the given plugin.
func (p *GoPlugin) GetJobs() ([]*gaia.Job, error) {
	l := make([]*gaia.Job, 0)

	// Get the stream
	stream, err := p.pluginConn.GetJobs()
	if err != nil {
		return nil, err
	}

	// receive all jobs
	pList := make([]*proto.Job, 0)
	jobsMap := make(map[uint32]*gaia.Job)
	for {
		job, err := stream.Recv()

		// Got all jobs
		if err == io.EOF {
			break
		}

		// Error during stream
		if err != nil {
			return nil, err
		}

		// Transform arguments
		args := make([]*gaia.Argument, 0, len(job.Args))
		for _, arg := range job.Args {
			a := &gaia.Argument{
				Description: arg.Description,
				Key:         arg.Key,
				Type:        arg.Type,
			}

			args = append(args, a)
		}

		outs := make([]*gaia.Output, 0, len(job.GetOuts()))
		for _, out := range job.GetOuts() {
			o := &gaia.Output{
				Key:   out.GetKey(),
				Value: out.GetValue(),
			}
			outs = append(outs, o)
		}

		// add proto object to separate list to rebuild dep later.
		pList = append(pList, job)

		// Convert proto object to gaia.Job struct
		j := &gaia.Job{
			ID:          job.UniqueId,
			Title:       job.Title,
			Description: job.Description,
			Status:      gaia.JobWaitingExec,
			Args:        args,
			Outs:        outs,
		}
		l = append(l, j)
		jobsMap[j.ID] = j
	}

	// Rebuild dependencies
	for _, pbJob := range pList {
		// Get job
		j := jobsMap[pbJob.UniqueId]

		// Iterate all dependencies
		j.DependsOn = make([]*gaia.Job, 0, len(pbJob.Dependson))
		for _, depJob := range pbJob.Dependson {
			// Get dependency
			depJ := jobsMap[depJob]

			// Set dependency
			j.DependsOn = append(j.DependsOn, depJ)
		}
	}

	// return list
	return l, nil
}

// FlushLogs flushes the logs.
func (p *GoPlugin) FlushLogs() error {
	return p.logger.Flush()
}

// Close shutdown the plugin and kills the gRPC connection.
// Remember to call this when you call plugin.Connect.
func (p *GoPlugin) Close() {
	// We start the kill command in a goroutine because kill
	// is blocking until the subprocess successfully exits.
	// The user should not wait for this.
	go func() {
		p.client.Kill()

		// Flush the writer
		_ = p.logger.Flush()

		// Close log file
		_ = p.logFile.Close()

		// Cleanup certificates
		_ = p.ca.CleanupCerts(p.certPath, p.keyPath)
		_ = p.ca.CleanupCerts(p.serverCertPath, p.serverKeyPath)
	}()
}
