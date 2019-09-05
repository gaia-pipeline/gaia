package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
)

// DockerWorker represents the data structure of a docker worker.
type DockerWorker struct {
	Host          string
	WorkerID      string
	ContainerID   string
	PipelineRunID string
}

// NewDockerWorker initiates a new worker instance.
func NewDockerWorker(host string, pipelineRunID string) *DockerWorker {
	return &DockerWorker{Host: host, PipelineRunID: pipelineRunID}
}

// SetupDockerWorker starts a Gaia worker inside a docker container, automatically
// connects it with this Gaia instance and sets a unique tag.
func (w *DockerWorker) SetupDockerWorker(workerImage string, workerSecret string) error {
	// Generate a unique id for this worker
	w.WorkerID = security.GenerateRandomUUIDV5()

	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		gaia.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	// Define small helper function which creates the docker container
	createContainer := func() (container.ContainerCreateCreatedBody, error) {
		return cli.ContainerCreate(ctx, &container.Config{
			Image: workerImage,
			Env: []string{
				"GAIA_WORKER_HOST_URL=" + gaia.Cfg.DockerWorkerHostURL,
				"GAIA_MODE=worker",
				"GAIA_WORKER_GRPC_HOST_URL=" + gaia.Cfg.DockerWorkerGRPCHostURL,
				"GAIA_WORKER_TAGS=" + fmt.Sprintf("%s,dockerworker", w.WorkerID),
				"GAIA_WORKER_SECRET=" + workerSecret,
			},
		}, &container.HostConfig{}, nil, "")
	}

	// Create container
	resp, err := createContainer()
	if err != nil {
		gaia.Cfg.Logger.Error("failed to create worker container", "error", err)
		gaia.Cfg.Logger.Info("try to pull docker image and then try it again...")

		// Pull worker image
		_, err = cli.ImagePull(ctx, workerImage, types.ImagePullOptions{})
		if err != nil {
			gaia.Cfg.Logger.Error("failed to pull worker image", "error", err)
			return err
		}

		// Try to create the container again
		resp, err = createContainer()
		if err != nil {
			gaia.Cfg.Logger.Error("failed to create worker container after pull", "error", err)
			return err
		}
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		gaia.Cfg.Logger.Error("failed to start worker container", "error", err)
		return err
	}

	// Store container id for later processing
	w.ContainerID = resp.ID
	return nil
}

// IsDockerWorkerRunning checks if the docker worker is running.
func (w *DockerWorker) IsDockerWorkerRunning() bool {
	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		gaia.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return false
	}
	cli.NegotiateAPIVersion(ctx)

	// Check if docker worker container is still running
	resp, err := cli.ContainerInspect(ctx, w.ContainerID)
	if err != nil {
		return false
	}

	if resp.State.Running {
		return true
	}
	return false
}

// KillDockerWorker disconnects the existing docker worker and
// kills the docker container.
func (w *DockerWorker) KillDockerWorker() error {
	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		gaia.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	// Kill container
	if err := cli.ContainerRemove(ctx, w.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	}); err != nil {
		gaia.Cfg.Logger.Error("failed to remove docker worker", "error", err, "containerid", w.ContainerID)
		return err
	}
	return nil
}