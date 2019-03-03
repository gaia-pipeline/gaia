package plugin

import (
	"context"

	proto "github.com/gaia-pipeline/protobuf"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GaiaPlugin is the Gaia plugin interface used for communication
// with the plugin.
type GaiaPlugin interface {
	GetJobs() (proto.Plugin_GetJobsClient, error)
	ExecuteJob(job *proto.Job) (*proto.JobResult, error)
}

// GaiaPluginClient represents gRPC client
type GaiaPluginClient struct {
	client proto.PluginClient
}

// GaiaPluginImpl represents the plugin implementation on client side.
type GaiaPluginImpl struct {
	Impl GaiaPlugin

	plugin.NetRPCUnsupportedPlugin
}

// GRPCServer is needed here to implement hashicorp
// plugin.Plugin interface. Real implementation is
// in the plugin(s).
func (p *GaiaPluginImpl) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
	// Real implementation defined in plugin
	return nil
}

// GaiaPluginClient is the passing method for the gRPC client.
func (p *GaiaPluginImpl) GaiaPluginClient(context context.Context, b *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GaiaPluginClient{client: proto.NewPluginClient(c)}, nil
}

// GetJobs requests all jobs from the plugin.
// We get a stream of proto.Job back.
func (m *GaiaPluginClient) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return m.client.GetJobs(context.Background(), &proto.Empty{})
}

// ExecuteJob triggers the execution of the given job in the plugin.
func (m *GaiaPluginClient) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return m.client.ExecuteJob(context.Background(), job)
}
