package plugin

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/scheduler"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/protobuf"
	plugin "github.com/hashicorp/go-plugin"
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
	pluginMapKey: &PluginGRPCImpl{},
}

// timeFormat is the logging time format.
const timeFormat = "2006/01/02 15:04:05"

// Plugin represents a single plugin instance which uses gRPC
// to connect to exactly one plugin.
type Plugin struct {
	// Client is an instance of the go-plugin client.
	client *plugin.Client

	// Client protocol instance used to open gRPC connections.
	clientProtocol plugin.ClientProtocol

	// Interface to the connected plugin.
	pluginConn PluginGRPC

	// Log file where all output is stored.
	logFile *os.File

	// Writer used to write logs from execution to file or buffer
	writer *bufio.Writer
	buffer *bytes.Buffer

	// CA instance used to handle certificates
	ca security.CAAPI

	// Created certificates path for pipeline run
	certPath       string
	keyPath        string
	serverCertPath string
	serverKeyPath  string
}

// NewPlugin creates a new instance of Plugin.
// One Plugin instance represents one connection to a plugin.
func (p *Plugin) NewPlugin(ca security.CAAPI) scheduler.Plugin {
	return &Plugin{ca: ca}
}

// Init prepares the log path, set's up new certificates for both gaia and
// plugin, and prepares the go-plugin client.
//
// It expects the start command for the plugin and the path where
// the log file should be stored.
//
// It's up to the caller to call plugin.Close to shutdown the plugin
// and close the gRPC connection.
func (p *Plugin) Init(command *exec.Cmd, logPath *string) error {
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
		p.writer = bufio.NewWriter(p.logFile)
	} else {
		// If no path is provided, write output to buffer
		p.buffer = new(bytes.Buffer)
		p.writer = bufio.NewWriter(p.buffer)
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
		Stderr:           p.writer,
		TLSConfig:        tlsConfig,
	})

	// Connect via gRPC
	p.clientProtocol, err = p.client.Client()
	if err != nil {
		p.writer.Flush()
		return fmt.Errorf("%s\n\n--- output ---\n%s", err.Error(), p.buffer.String())
	}

	return nil
}

// Validate validates the interface of the plugin.
func (p *Plugin) Validate() error {
	// Request the plugin
	raw, err := p.clientProtocol.Dispense(pluginMapKey)
	if err != nil {
		return err
	}

	// Convert plugin to interface
	if pC, ok := raw.(PluginGRPC); ok {
		p.pluginConn = pC
		return nil
	}

	return errors.New("plugin is not compatible with plugin interface")
}

// Execute triggers the execution of one single job
// for the given plugin.
func (p *Plugin) Execute(j *gaia.Job) error {
	// Transform arguments
	args := []*proto.Argument{}
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
		p.writer.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, resultObj.Message))
	} else if err != nil {
		// An error occured during the send or somewhere else.
		// The job itself usually does not return an error here.
		// We mark the job as failed.
		j.Status = gaia.JobFailed

		// Generate error message and attach it to logs.
		timeString := time.Now().Format(timeFormat)
		p.writer.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, err.Error()))
	} else {
		j.Status = gaia.JobSuccess
	}

	return nil
}

// GetJobs receives all implemented jobs from the given plugin.
func (p *Plugin) GetJobs() ([]gaia.Job, error) {
	l := []gaia.Job{}

	// Get the stream
	stream, err := p.pluginConn.GetJobs()
	if err != nil {
		return nil, err
	}

	// receive all jobs
	pList := []*proto.Job{}
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
		args := []gaia.Argument{}
		for _, arg := range job.Args {
			a := gaia.Argument{
				Description: arg.Description,
				Key:         arg.Key,
				Type:        arg.Type,
			}

			args = append(args, a)
		}

		// add proto object to separate list to rebuild dep later.
		pList = append(pList, job)

		// Convert proto object to gaia.Job struct
		j := gaia.Job{
			ID:          job.UniqueId,
			Title:       job.Title,
			Description: job.Description,
			Status:      gaia.JobWaitingExec,
			Args:        args,
		}
		l = append(l, j)
	}

	// Rebuild dependency tree
	for id, job := range l {
		for _, pJob := range pList {
			if job.ID == pJob.UniqueId {
				l[id].DependsOn = rebuildDepTree(pJob.Dependson, l)
			}
		}
	}

	// return list
	return l, nil
}

// FlushLogs flushes the logs.
func (p *Plugin) FlushLogs() error {
	return p.writer.Flush()
}

// rebuildDepTree resolves related depenendencies and returns
// list of pointers to dependent jobs.
func rebuildDepTree(dep []uint32, l []gaia.Job) []*gaia.Job {
	depTree := []*gaia.Job{}
	for _, jobHash := range dep {
		for id, job := range l {
			if job.ID == jobHash {
				depTree = append(depTree, &l[id])
			}
		}
	}
	return depTree
}

// Close shutdown the plugin and kills the gRPC connection.
// Remember to call this when you call plugin.Connect.
func (p *Plugin) Close() {
	// We start the kill command in a goroutine because kill
	// is blocking until the subprocess successfully exits.
	// The user should not wait for this.
	go func() {
		p.client.Kill()

		// Flush the writer
		p.writer.Flush()

		// Close log file
		p.logFile.Close()

		// Cleanup certificates
		p.ca.CleanupCerts(p.certPath, p.keyPath)
		p.ca.CleanupCerts(p.serverCertPath, p.serverKeyPath)
	}()
}
