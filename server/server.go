package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gaia-pipeline/gaia/workers/pipeline"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gaia-pipeline/flag"
	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/handlers"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/agent"
	"github.com/gaia-pipeline/gaia/workers/server"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/labstack/echo"
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

	// set a prefix for environment properties so they are destinct to Gaia
	fs = flag.NewFlagSetWithEnvPrefix(os.Args[0], "GAIA", 0)

	// set the configuration filename
	fs.String("config", ".gaia_config", "this describes the name of the config file to use")

	// command line arguments
	fs.StringVar(&gaia.Cfg.ListenPort, "port", "8080", "Listen port for gaia")
	fs.StringVar(&gaia.Cfg.HomePath, "homepath", "", "Path to the gaia home folder")
	fs.StringVar(&gaia.Cfg.Hostname, "hostname", "https://localhost", "The host's name under which gaia is deployed at e.g.: https://gaia-pipeline.com")
	fs.StringVar(&gaia.Cfg.VaultPath, "vaultpath", "", "Path to the gaia vault folder")
	fs.IntVar(&gaia.Cfg.Worker, "worker", 2, "Number of worker gaia will use to execute pipelines in parallel")
	fs.StringVar(&gaia.Cfg.JwtPrivateKeyPath, "jwtPrivateKeyPath", "", "A RSA private key used to sign JWT tokens")
	fs.StringVar(&gaia.Cfg.CAPath, "capath", "", "Folder path where the generated CA certificate files will be saved")
	fs.BoolVar(&gaia.Cfg.DevMode, "dev", false, "If true, gaia will be started in development mode. Don't use this in production!")
	fs.BoolVar(&gaia.Cfg.VersionSwitch, "version", false, "If true, will print the version and immediately exit")
	fs.BoolVar(&gaia.Cfg.Poll, "poll", false, "Instead of using a Webhook, keep polling git for changes on pipelines")
	fs.IntVar(&gaia.Cfg.PVal, "pval", 1, "The interval in minutes in which to poll vcs for changes")
	fs.StringVar(&gaia.Cfg.ModeRaw, "mode", "server", "The mode in which gaia should be started. Possible options are server and worker")
	fs.StringVar(&gaia.Cfg.WorkerHostURL, "hosturl", "http://localhost:8080", "The host url of an gaia instance to connect to. Only used in worker mode")
	fs.StringVar(&gaia.Cfg.WorkerGRPCHostURL, "grpchosturl", "localhost:8989", "The host url of an gaia instance gRPC interface used for worker connection. Only used in worker mode")
	fs.StringVar(&gaia.Cfg.WorkerSecret, "workersecret", "", "The secret which is used to register a worker at an gaia instance")
	fs.StringVar(&gaia.Cfg.WorkerServerPort, "workerserverport", "8989", "Listen port for gaia worker gRPC communication")
	fs.StringVar(&gaia.Cfg.WorkerTags, "workertags", "", "Comma separated list of custom tags for this worker")

	// Default values
	gaia.Cfg.Bolt.Mode = 0600
}

// Start initiates all components of Gaia and starts the server/agent.
func Start() (err error) {
	// Parse command line flgs
	fs.Parse(os.Args[1:])

	// Check version switch
	if gaia.Cfg.VersionSwitch {
		fmt.Printf("Gaia Version: V%s\n", Version)
		os.Exit(0)
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
		// Find executeable path
		execPath, err := findExecutablePath()
		if err != nil {
			gaia.Cfg.Logger.Error("cannot find executeable path", "error", err.Error())
			return err
		}
		gaia.Cfg.HomePath = execPath
		gaia.Cfg.Logger.Debug("executeable path found", "path", execPath)
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

	if gaia.Cfg.Mode == gaia.ModeServer {
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
		v, err := services.VaultService(nil)
		if err != nil {
			gaia.Cfg.Logger.Error("error initiating vault")
			return err
		}
		if err = v.LoadSecrets(); err != nil {
			gaia.Cfg.Logger.Error("error loading secrets from vault")
			return err
		}

		// Generate global worker secret if it does not exist
		secret, err := v.Get(gaia.WorkerRegisterKey)
		if err != nil {
			// Secret hasn't been generated yet
			gaia.Cfg.Logger.Info("global worker registration secret has not been generated yet. Will generate it now...")
			secret = []byte(security.GenerateRandomUUIDV5())

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
	}

	// Initialize scheduler
	scheduler, err := services.SchedulerService()
	if err != nil {
		return
	}

	switch gaia.Cfg.Mode {
	case gaia.ModeServer:
		// Start ticker. Periodic job to check for new plugins.
		pipeline.InitTicker()

		// Start worker gRPC server
		workerServer := server.InitWorkerServer()
		go workerServer.Start()

		// Start listen
		echoInstance.Logger.Fatal(echoInstance.Start(":" + gaia.Cfg.ListenPort))
	case gaia.ModeWorker:
		// Start worker main loop and block until SIGINT or SIGTERM has been received
		ag := agent.InitAgent(scheduler, store)
		err := ag.StartAgent()
		if err != nil {
			gaia.Cfg.Logger.Error("failed to start agent", "error", err)
			return err
		}
	}
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
