package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
)

type Worker struct {
	Host        string
	WorkerID    string
	ContainerID string
}

// NewWorker initiates a new worker instance.
func NewWorker(host string) *Worker {
	return &Worker{Host: host}
}

// SetupDockerWorker starts a Gaia worker inside a docker container, automatically
// connects it with this Gaia instance and sets a unique tag.
func (w *Worker) SetupDockerWorker(workerImage string) error {
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

	// Pull worker image from docker hub
	_, err = cli.ImagePull(ctx, workerImage, types.ImagePullOptions{})
	if err != nil {
		gaia.Cfg.Logger.Error("failed to pull worker image", "error", err)
		return err
	}

	// Retrieve the global worker registration secret
	v, err := services.VaultService(nil)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to access vault", "error", err)
		return err
	}
	if err = v.LoadSecrets(); err != nil {
		gaia.Cfg.Logger.Error("failed to load secrets from vault", "error", err)
		return err
	}
	workerSecret, err := v.Get(gaia.WorkerRegisterKey)
	if err != nil {
		gaia.Cfg.Logger.Error("failed to get global worker registration secret from vault", "error", err)
		return err
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: workerImage,
		Env: []string{
			"GAIA_HOST_URL=" + gaia.Cfg.WorkerHostURL,
			"GAIA_MODE=worker",
			"GAIA_WORKER_GRPC_HOST_URL=" + gaia.Cfg.WorkerGRPCHostURL,
			"GAIA_WORKER_TAGS=" + fmt.Sprintf("%s,dockerworker", w.WorkerID),
			"GAIA_WORKER_SECRET=" + string(workerSecret[:]),
		},
	}, &container.HostConfig{}, nil, "")

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		gaia.Cfg.Logger.Error("failed to start worker container", "error", err)
		return err
	}

	// Store container id for later processing
	w.ContainerID = resp.ID
	return nil
}

// KillDockerWorker disconnects the existing docker worker and
// kills the docker container.
func (w *Worker) KillDockerWorker() error {
	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		gaia.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	// Get memdb service
	db, err := services.MemDBService(nil)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot get memdb service from store", "error", err)
		return err
	}

	// Check if worker is still registered
	worker, err := db.GetWorker(w.WorkerID)
	if err != nil || worker == nil {
		gaia.Cfg.Logger.Warn("failed to deregister docker worker. It has been already deregistered!")
	}

	// Delete worker which basically indicates it is not registered anymore
	if worker != nil {
		if err := db.DeleteWorker(worker.UniqueID, true); err != nil {
			gaia.Cfg.Logger.Error("failed to delete docker worker", "error", err)
			return err
		}
	}

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
