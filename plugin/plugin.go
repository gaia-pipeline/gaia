package plugin

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"

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

// Plugin represents a single plugin instance which uses gRPC
// to connect to exactly one plugin.
type Plugin struct {
	// Client instance used to open gRPC connections.
	client *plugin.Client

	// Interface to the connected plugin.
	pluginConn PluginGRPC

	// Log file where all output is stored.
	logFile *os.File

	// Writer used to write logs from execution to file
	writer *bufio.Writer

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

// Connect prepares the log path, starts the plugin, initiates the
// gRPC connection and looks up the plugin.
// It's up to the caller to call plugin.Close to shutdown the plugin
// and close the gRPC connection.
//
// It expects the start command for the plugin and the path where
// the log file should be stored.
func (p *Plugin) Connect(command *exec.Cmd, logPath *string) error {
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
	}

	// Create new writer
	p.writer = bufio.NewWriter(p.logFile)

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
	gRPCClient, err := p.client.Client()
	if err != nil {
		return err
	}

	// Request the plugin
	raw, err := gRPCClient.Dispense(pluginMapKey)
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
	// Create new proto job object and just set the id.
	// The rest is currently not important.
	job := &proto.Job{
		UniqueId: j.ID,
	}

	// Execute the job
	_, err := p.pluginConn.ExecuteJob(job)

	// Flush logs
	p.writer.Flush()

	return err
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

		// add proto object to separate list to rebuild dep later.
		pList = append(pList, job)

		// Convert proto object to gaia.Job struct
		j := gaia.Job{
			ID:          job.UniqueId,
			Title:       job.Title,
			Description: job.Description,
			Status:      gaia.JobWaitingExec,
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
