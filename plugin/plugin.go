package plugin

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/proto"
	plugin "github.com/hashicorp/go-plugin"
)

const (
	pluginMapKey = "Plugin"
)

var handshake = plugin.HandshakeConfig{
	ProtocolVersion: 1,
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
}

// NewPlugin creates a new instance of Plugin.
// One Plugin instance represents one connection to a plugin.
//
// It expects the start command to start the plugin and the log path (including file)
// where the output should be logged to.
func NewPlugin(command *exec.Cmd, logPath *string) (p *Plugin, err error) {
	// Allocate
	p = &Plugin{}

	// Create log file and open it.
	// We will close this file in the close method.
	if logPath != nil {
		p.logFile, err = os.OpenFile(
			*logPath,
			os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return nil, err
		}
	}

	// Create new writer
	p.writer = bufio.NewWriter(p.logFile)

	// Get new client
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshake,
		Plugins:          pluginMap,
		Cmd:              command,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Stderr:           p.writer,
	})

	return p, nil
}

// Connect starts the plugin, initiates the gRPC connection and looks up the plugin.
// It's up to the caller to call plugin.Close to shutdown the plugin
// and close the gRPC connection.
func (p *Plugin) Connect() error {
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

		// Convert proto object to gaia.Job struct
		j := gaia.Job{
			ID:          job.UniqueId,
			Title:       job.Title,
			Description: job.Description,
			Priority:    job.Priority,
			Status:      gaia.JobWaitingExec,
		}
		l = append(l, j)
	}

	// return list
	return l, nil
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
	}()
}
