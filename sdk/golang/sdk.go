package golang

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/gaia-pipeline/gaia/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// coreProtocolVersion is the protocol version of the plugin system itself.
const coreProtocolVersion = 1

// ProtocolVersion currently in use by Gaia
const ProtocolVersion = 1

// ProtocolType is the type used to communicate.
const ProtocolType = "grpc"

var (
	// ErrorJobNotFound is returned when a given job id was not found
	// locally.
	ErrorJobNotFound = errors.New("job not found in plugin")

	// ErrorExitPipeline is used to safely exit the pipeline (actually not an error).
	// Prevents the pipeline to be marked as 'failed'.
	ErrorExitPipeline = errors.New("pipeline exit requested by job")
)

// CachedJobs holds a list of JobsWrapper for later processing
var cachedJobs Jobs

// Args are the args passed from client.
// They are injected before job will be started.
var Args map[string]string

// GRPCServer is the plugin gRPC implementation.
type GRPCServer struct{}

// GetJobs streams all given jobs back.
func (GRPCServer) GetJobs(empty *proto.Empty, stream proto.Plugin_GetJobsServer) error {
	for _, job := range cachedJobs {
		err := stream.Send(&job.Job)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecuteJob receives a job and executes it.
// Returns a JobResult object which gives information about job execution.
func (GRPCServer) ExecuteJob(ctx context.Context, j *proto.Job) (*proto.JobResult, error) {
	job := cachedJobs.Get((*j).UniqueId)
	if job == nil {
		return nil, ErrorJobNotFound
	}

	// Set passed arguments
	Args = job.Job.Args

	// Execute Job
	err := job.FuncPointer()

	// Generate result object only when we got an error
	r := &proto.JobResult{}
	if err != nil {
		// Check if job wants to force exit pipeline.
		// We will exit the pipeline but not mark as 'failed'.
		if err == ErrorExitPipeline {
			r.ExitPipeline = true
		} else {
			// We got an error. Pipeline is now marked as 'failed'.
			r.ExitPipeline = true
			r.Failed = true
		}

		// Set log message and job id
		r.Message = err.Error()
		r.UniqueId = job.Job.UniqueId
	}

	return r, err
}

// Serve initiates the gRPC Server and listens...forever.
// This method should be last called in the plugin main function.
func Serve(j Jobs) {
	// Cache the jobs list for later processing
	cachedJobs = j

	// Get unix listener
	lis, err := serverListenerUnix()
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	// implement health service
	health := health.NewServer()
	health.SetServingStatus("plugin", healthpb.HealthCheckResponse_SERVING)

	// Create new gRPC server and register services
	s := grpc.NewServer()
	proto.RegisterPluginServer(s, &GRPCServer{})
	healthpb.RegisterHealthServer(s, health)

	// Register reflection service on gRPC server
	reflection.Register(s)

	// Output the address and service name to stdout.
	// hashicorp go-plugin will use that to establish connection.
	fmt.Printf("%d|%d|%s|%s|%s\n",
		coreProtocolVersion,
		ProtocolVersion,
		lis.Addr().Network(),
		lis.Addr().String(),
		ProtocolType)
	os.Stdout.Sync()

	// Listen
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func serverListenerUnix() (net.Listener, error) {
	tf, err := ioutil.TempFile("", "plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
