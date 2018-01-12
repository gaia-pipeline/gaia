package plugin

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/michelvocks/gaia/proto"
	"google.golang.org/grpc"
)

// PluginGRPC is the Gaia plugin interface used for communication
// with the plugin.
type PluginGRPC interface {
	GetJobs() (*proto.Plugin_GetJobsClient, error)
	ExecuteJob(job *proto.Job) (*proto.Empty, error)
}

// GRPCClient represents gRPC client
type GRPCClient struct {
	client proto.PluginClient
}

// PluginGRPCImpl represents the plugin implementation on client side.
type PluginGRPCImpl struct {
	Impl PluginGRPC
	plugin.NetRPCUnsupportedPlugin
}

// GRPCServer is needed here to implement hashicorp
// plugin.Plugin interface. Real implementation is
// in the plugin(s).
func (p *PluginGRPCImpl) GRPCServer(s *grpc.Server) error {
	// Real implementation defined in plugin
	return nil
}

// GRPCClient is the passing method for the gRPC client.
func (p *PluginGRPCImpl) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewPluginClient(c)}, nil
}

// GetJobs requests all jobs from the plugin.
// We get a stream of proto.Job back.
func (m *GRPCClient) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return m.client.GetJobs(context.Background(), &proto.Empty{})
}

// ExecuteJob triggers the execution of the given job in the plugin.
func (m *GRPCClient) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return m.client.ExecuteJob(context.Background(), job)
}
