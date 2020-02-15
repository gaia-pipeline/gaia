package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gaia-pipeline/gaia/workers/server"

	"github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
	"github.com/gaia-pipeline/flag"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/gaia-pipeline/gaia/workers/agent"
)

var (
	echoInstance *echo.Echo
)

const (
	// Version is the current version of gaia.
	Version = "0.2.3"

	dataFolder      = "data"
	pipelinesFolder = "pipelines"
	workspaceFolder = "workspace"
)

var fs *flag.FlagSet

func init() {
	// set configuration file name for run-time arguments

	// set a prefix for environment properties so they are distinct to Gaia
	fs = flag.NewFlagSetWithEnvPrefix(os.Args[0], "GAIA", 0)

	// set the configuration filename
	fs.String("config", ".gaia_config", "this describes the name of the config file to use")

	// command line arguments
	fs.StringVar(&gaia.Cfg.ListenPort, "port", "8080", "Listen port for Gaia")
	fs.StringVar(&gaia.Cfg.HomePath, "home-path", "", "Path to the Gaia home folder where all data will be stored")
	fs.StringVar(&gaia.Cfg.Hostname, "hostname", "https://localhost", "The host's name under which Gaia is deployed at e.g.: https://gaia-pipeline.io")
	fs.StringVar(&gaia.Cfg.VaultPath, "vault-path", "", "Path to the Gaia vault folder. By default, will be stored inside the home folder")
	fs.IntVar(&gaia.Cfg.Worker, "concurrent-worker", 2, "Number of concurrent worker the Gaia instance will use to execute pipelines in parallel")
	fs.StringVar(&gaia.Cfg.JwtPrivateKeyPath, "jwt-private-key-path", "", "A RSA private key used to sign JWT tokens used for Web UI authentication")
	fs.StringVar(&gaia.Cfg.CAPath, "ca-path", "", "Path where the generated CA certificate files will be saved")
	fs.BoolVar(&gaia.Cfg.DevMode, "dev", false, "If true, Gaia will be started in development mode. Don't use this in production!")
	fs.BoolVar(&gaia.Cfg.VersionSwitch, "version", false, "If true, will print the version and immediately exit")
	fs.BoolVar(&gaia.Cfg.Poll, "pipeline-poll", false, "If true, Gaia will periodically poll pipeline repositories, watch for changes and rebuild them accordingly")
	fs.IntVar(&gaia.Cfg.PVal, "pipeline-poll-interval", 1, "The interval in minutes in which to poll source repositories for changes")
	fs.StringVar(&gaia.Cfg.ModeRaw, "mode", "server", "The mode which Gaia should be started in. Possible options are server and worker")
	fs.StringVar(&gaia.Cfg.WorkerName, "worker-name", "", "The name of the worker which will be displayed at the primary instance. Only used in worker mode or for docker runs")
	fs.StringVar(&gaia.Cfg.WorkerHostURL, "worker-host-url", "http://127.0.0.1:8080", "The host url of an Gaia primary instance to connect to. Only used in worker mode or for docker runs")
	fs.StringVar(&gaia.Cfg.WorkerGRPCHostURL, "worker-grpc-host-url", "127.0.0.1:8989", "The host url of an Gaia primary instance gRPC interface used for worker connection. Only used in worker mode or for docker runs")
	fs.StringVar(&gaia.Cfg.WorkerSecret, "worker-secret", "", "The secret which is used to register a worker at an Gaia primary instance. Only used in worker mode")
	fs.StringVar(&gaia.Cfg.WorkerServerPort, "worker-server-port", "8989", "Listen port for Gaia primary worker gRPC communication. Only used in server mode")
	fs.StringVar(&gaia.Cfg.WorkerTags, "worker-tags", "", "Comma separated list of custom tags for this worker. Only used in worker mode")
	fs.BoolVar(&gaia.Cfg.PreventPrimaryWork, "prevent-primary-work", false, "If true, prevents the scheduler to schedule work on this Gaia primary instance. Only used in server mode")
	fs.BoolVar(&gaia.Cfg.AutoDockerMode, "auto-docker-mode", false, "If true, by default runs all pipelines in a docker container")
	fs.StringVar(&gaia.Cfg.DockerHostURL, "docker-host-url", "unix:///var/run/docker.sock", "Docker daemon host url which is used to build and run pipelines in a docker container")
	fs.StringVar(&gaia.Cfg.DockerRunImage, "docker-run-image", "gaiapipeline/gaia:latest", "Docker image repository name with tag which will be used for running pipelines in a docker container")
	fs.StringVar(&gaia.Cfg.DockerWorkerHostURL, "docker-worker-host-url", "http://127.0.0.1:8080", "The host url of the primary/worker API endpoint used for docker worker communication")
	fs.StringVar(&gaia.Cfg.DockerWorkerGRPCHostURL, "docker-worker-grpc-host-url", "127.0.0.1:8989", "The host url of the primary/worker gRPC endpoint used for docker worker communication")

	// Default values
	gaia.Cfg.Bolt.Mode = 0600
}

// Start initiates all components of Gaia and starts the server/agent.
func Start() (err error) {
	// Parse command line flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err.Error() == "flag: help requested" {
			return nil
		}
	}

	// Check version switch
	if gaia.Cfg.VersionSwitch {
		fmt.Printf("Gaia Version: V%s\n", Version)
		return
	}

	// Initialize shared logger
	gaia.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Gaia",
	})

	// Determine the mode in which Gaia should be started
	switch gaia.Cfg.ModeRaw {
	case "server":
		gaia.Cfg.Mode = gaia.ModeServer
	case "worker":
		gaia.Cfg.Mode = gaia.ModeWorker
	default:
		gaia.Cfg.Logger.Error("unsupported mode used", "mode", gaia.Cfg.Mode)
		return errors.New("unsupported mode used")
	}

	// Find path for gaia home folder if not given by parameter
	if gaia.Cfg.HomePath == "" {
		// Find executable path
		execPath, err := findExecutablePath()
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find executable path", "error", err.Error())
			return err
		}
		gaia.Cfg.HomePath = execPath
		gaia.Cfg.Logger.Debug("executable path found", "path", execPath)
	}

	// Set data path, workspace path and pipeline path relative to home folder and create it
	// if not exist.
	gaia.Cfg.DataPath = filepath.Join(gaia.Cfg.HomePath, dataFolder)
	err = os.MkdirAll(gaia.Cfg.DataPath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", gaia.Cfg.DataPath)
		return
	}
	gaia.Cfg.PipelinePath = filepath.Join(gaia.Cfg.HomePath, pipelinesFolder)
	err = os.MkdirAll(gaia.Cfg.PipelinePath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create folder", "error", err.Error(), "path", gaia.Cfg.PipelinePath)
		return
	}
	gaia.Cfg.WorkspacePath = filepath.Join(gaia.Cfg.HomePath, workspaceFolder)
	err = os.MkdirAll(gaia.Cfg.WorkspacePath, 0700)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create data folder", "error", err.Error(), "path", gaia.Cfg.WorkspacePath)
		return
	}

	// Check CA path
	if gaia.Cfg.CAPath == "" {
		// Set default to data folder
		gaia.Cfg.CAPath = gaia.Cfg.DataPath
	}

	// Initialize the certificate manager service
	_, err = services.CertificateService()
	if err != nil {
		gaia.Cfg.Logger.Error("cannot create CA", "error", err.Error())
		return
	}

	// Initialize store
	store, err := services.StorageService()
	if err != nil {
		return
	}

	// Initialize MemDB
	db, err := services.MemDBService(store)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize memdb service", "error", err.Error())
		return err
	}
	if err = db.SyncStore(); err != nil {
		return err
	}

	var jwtKey interface{}
	// Check JWT key is set
	if gaia.Cfg.JwtPrivateKeyPath == "" {
		gaia.Cfg.Logger.Warn("using auto-generated key to sign jwt tokens, do not use in production")
		jwtKey = make([]byte, 64)
		_, err = rand.Read(jwtKey.([]byte))
		if err != nil {
			gaia.Cfg.Logger.Error("error auto-generating jwt key", "error", err.Error())
			return
		}
	} else {
		keyData, err := ioutil.ReadFile(gaia.Cfg.JwtPrivateKeyPath)
		if err != nil {
			gaia.Cfg.Logger.Error("could not read jwt key file", "error", err.Error())
			return err
		}
		jwtKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			gaia.Cfg.Logger.Error("could not parse jwt key file", "error", err.Error())
			return err
		}
	}
	gaia.Cfg.JWTKey = jwtKey

	// Initialize echo instance
	echoInstance = echo.New()

	// Initiating Vault
	if gaia.Cfg.VaultPath == "" {
		// Set default to data folder
		gaia.Cfg.VaultPath = gaia.Cfg.DataPath
	}
	v, err := services.DefaultVaultService()
	if err != nil {
		gaia.Cfg.Logger.Error("error initiating vault")
		return err
	}
	if err = v.LoadSecrets(); err != nil {
		gaia.Cfg.Logger.Error("error loading secrets from vault")
		return err
	}

	// Generate global worker secret if it does not exist
	_, err = v.Get(gaia.WorkerRegisterKey)
	if err != nil {
		// Secret hasn't been generated yet
		gaia.Cfg.Logger.Info("global worker registration secret has not been generated yet. Will generate it now...")
		secret := []byte(security.GenerateRandomUUIDV5())

		// Store secret in vault
		v.Add(gaia.WorkerRegisterKey, secret)
		if err := v.SaveSecrets(); err != nil {
			gaia.Cfg.Logger.Error("failed to store secret into vault", "error", err.Error())
			return err
		}
	}

	// Initialize handlers
	err = handlers.InitHandlers(echoInstance)
	if err != nil {
		gaia.Cfg.Logger.Error("cannot initialize handlers", "error", err.Error())
		return err
	}

	// Initialize scheduler
	scheduler, err := services.SchedulerService()
	if err != nil {
		return
	}

	// Allocate SIG channel
	exitChan := make(chan os.Signal, 1)

	// Register the signal channel
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker gRPC server.
	// We need this in both modes (server and worker) for docker worker to run.
	workerServer := server.InitWorkerServer()
	go func() {
		if err := workerServer.Start(); err != nil {
			gaia.Cfg.Logger.Error("failed to start gRPC worker server", "error", err)
			exitChan <- syscall.SIGTERM
		}
	}()

	cleanUpFunc := func() {}
	switch gaia.Cfg.Mode {
	case gaia.ModeServer:
		// Start ticker. Periodic job to check for new plugins.
		pipeline.InitTicker()

		// Start API server
		go func() {
			err := echoInstance.Start(":" + gaia.Cfg.ListenPort)
			if err != nil {
				gaia.Cfg.Logger.Error("failed to start echo listener", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()
	case gaia.ModeWorker:
		// Start API server
		go func() {
			err := echoInstance.Start(":" + gaia.Cfg.ListenPort)
			if err != nil {
				gaia.Cfg.Logger.Error("failed to start echo listener", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()

		// Start agent
		ag := agent.InitAgent(exitChan, scheduler, store, gaia.Cfg.HomePath)
		go func() {
			cleanUpFunc, err = ag.StartAgent()
			if err != nil {
				gaia.Cfg.Logger.Error("failed to start agent", "error", err)
				exitChan <- syscall.SIGTERM
			}
		}()
	}

	// Wait for exit signal
	<-exitChan
	gaia.Cfg.Logger.Info("exit signal received. Exiting...")

	// Run clean up func
	cleanUpFunc()
	return
}

// findExecutablePath returns the absolute path for the current
// process.
func findExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}
