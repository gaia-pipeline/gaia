package pipelines

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/pipelinehelper"
	"github.com/gaia-pipeline/gaia/helper/stringhelper"
	"github.com/gaia-pipeline/gaia/security"
	"github.com/gaia-pipeline/gaia/services"
	"github.com/gaia-pipeline/gaia/workers/pipeline"
)

var (
	// errPipelineNotFound is thrown when a pipeline was not found with the given id
	errPipelineNotFound = errors.New("pipeline not found with the given id")

	// errInvalidPipelineID is thrown when the given pipeline id is not valid
	errInvalidPipelineID = errors.New("the given pipeline id is not valid")

	// errPipelineDelete is thrown when a pipeline binary could not be deleted
	errPipelineDelete = errors.New("pipeline could not be deleted. Perhaps you don't have the right permissions")

	// errPipelineRename is thrown when a pipeline binary could not be renamed
	errPipelineRename = errors.New("pipeline could not be renamed")

	// errWrongDockerValue is thrown when docker has been specified for a pipeline run but the value is invalid
	errWrongDockerValue = errors.New("invalid value for docker parameter")
)

// PipelineGitLSRemote checks for available git remote branches.
// This is the perfect way to check if we have access to a given repo.
// @Summary Check for repository access.
// @Description Checks for available git remote branches which in turn verifies repository access.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param PipelineGitLSRemoteRequest body gaia.GitRepo true "The repository details"
// @Success 200 {array} string "Available branches"
// @Failure 400 {string} string "Failed to bind body"
// @Failure 403 {string} string "No access"
// @Router /pipeline/gitlsremote [post]
func (pp *PipelineProvider) PipelineGitLSRemote(c echo.Context) error {
	repo := &gaia.GitRepo{}
	if err := c.Bind(repo); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Check for remote branches
	err := pipeline.GitLSRemote(repo)
	if err != nil {
		return c.String(http.StatusForbidden, err.Error())
	}

	// Return branches
	return c.JSON(http.StatusOK, repo.Branches)
}

// CreatePipeline accepts all data needed to create a pipeline.
// It then starts the create pipeline execution process async.
// @Summary Create pipeline.
// @Description Starts creating a pipeline given all the data asynchronously.
// @Tags pipelines
// @Accept json
// @Produce plain
// @Param CreatePipelineRequest body gaia.CreatePipeline true "Create pipeline details"
// @Success 200
// @Failure 400 {string} string "Failed to bind, validation error and invalid details"
// @Failure 500 {string} string "Internal error while saving create pipeline run"
// @Router /pipeline [post]
func (pp *PipelineProvider) CreatePipeline(c echo.Context) error {
	storeService, _ := services.StorageService()
	p := &gaia.CreatePipeline{}
	if err := c.Bind(p); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Validate pipeline name
	if err := pipeline.ValidatePipelineName(p.Pipeline.Name); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Set initial value
	p.Created = time.Now()
	p.StatusType = gaia.CreatePipelineRunning
	v4, err := uuid.NewV4()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	p.ID = uuid.Must(v4, nil).String()

	// Add pipeline type tag if not already existent
	if !stringhelper.IsContainedInSlice(p.Pipeline.Tags, p.Pipeline.Type.String(), true) {
		p.Pipeline.Tags = append(p.Pipeline.Tags, p.Pipeline.Type.String())
	}

	// Save this pipeline to our store
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot put pipeline into store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Cloning the repo and compiling the pipeline will be done async
	go pp.deps.PipelineService.CreatePipeline(p)

	return c.JSON(http.StatusOK, nil)
}

// CreatePipelineGetAll returns a json array of
// all pipelines which are about to get compiled and
// all pipelines which have been compiled.
// @Summary Get all create pipelines.
// @Description Get a list of all pipelines which are about to be compiled and which have been compiled.
// @Tags pipelines
// @Produce json
// @Success 200 {array} gaia.CreatePipeline
// @Failure 500 {string} string "Internal error while retrieving create pipeline data."
// @Router /pipeline/created [get]
func (pp *PipelineProvider) CreatePipelineGetAll(c echo.Context) error {
	// Get all create pipelines
	storeService, _ := services.StorageService()
	pipelineList, err := storeService.CreatePipelineGet()
	if err != nil {
		gaia.Cfg.Logger.Debug("cannot get create pipelines from store", "error", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return all create pipelines
	return c.JSON(http.StatusOK, pipelineList)
}

// PipelineNameAvailable looks up if the given pipeline name is
// available and valid.
// @Summary Pipeline name validation.
// @Description Looks up if the given pipeline name is available and valid.
// @Tags pipelines
// @Accept plain
// @Produce json
// @Param name query string true "The name of the pipeline to validate"
// @Success 200
// @Failure 400 {string} string "Pipeline name validation errors"
// @Router /pipeline/name [get]
func (pp *PipelineProvider) PipelineNameAvailable(c echo.Context) error {
	pName := c.QueryParam("name")
	if err := pipeline.ValidatePipelineName(pName); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, nil)
}

// PipelineGetAll returns all registered pipelines.
// @Summary Returns all registered pipelines.
// @Description Returns all registered pipelines.
// @Tags pipelines
// @Produce json
// @Success 200 {array} gaia.Pipeline
// @Router /pipeline/name [get]
func (pp *PipelineProvider) PipelineGetAll(c echo.Context) error {
	// Get all active pipelines
	pipelines := pipeline.GlobalActivePipelines.GetAll()

	// Obscure non-necessary information
	for id := range pipelines {
		obscurePipelineData(&pipelines[id])
	}

	// Return as json
	return c.JSON(http.StatusOK, pipelines)
}

// PipelineGet accepts a pipeline id and returns the pipeline object.
// @Summary Get pipeline information.
// @Description Get pipeline information based on ID.
// @Tags pipelines
// @Accept plain
// @Produce json
// @Param pipelineid query string true "The ID of the pipeline"
// @Success 200 {object} gaia.Pipeline
// @Failure 400 {string} string "The given pipeline id is not valid"
// @Failure 404 {string} string "Pipeline not found with the given id"
// @Router /pipeline/{pipelineid} [get]
func (pp *PipelineProvider) PipelineGet(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	for _, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			obscurePipelineData(&p)
			return c.JSON(http.StatusOK, p)
		}
	}

	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

// PipelineUpdate updates the given pipeline.
// @Summary Update pipeline.
// @Description Update a pipeline by its ID.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param PipelineUpdateRequest body gaia.Pipeline true "PipelineUpdate request"
// @Success 200 {string} string "Pipeline has been updated"
// @Failure 400 {string} string "Error while updating the pipeline"
// @Failure 404 {string} string "The pipeline with the given ID was not found"
// @Failure 500 {string} string "Internal error while updating and building the new pipeline information"
// @Router /pipeline/{pipelineid} [put]
func (pp *PipelineProvider) PipelineUpdate(c echo.Context) error {
	storeService, _ := services.StorageService()
	p := gaia.Pipeline{}
	if err := c.Bind(&p); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, pipe := range pipeline.GlobalActivePipelines.GetAll() {
		if pipe.ID == p.ID {
			foundPipeline = pipe
			break
		}
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusNotFound, errPipelineNotFound.Error())
	}

	// Check if the pipeline name was changed.
	if foundPipeline.Name != p.Name {
		// Pipeline name has been changed
		currentName := foundPipeline.Name

		// Rename binary
		err := pipeline.RenameBinary(foundPipeline, p.Name)
		if err != nil {
			return c.String(http.StatusInternalServerError, errPipelineRename.Error())
		}

		// Update name and exec path
		foundPipeline.Name = p.Name
		foundPipeline.ExecPath = pipeline.GetExecPath(p)

		// Update pipeline in store
		err = storeService.PipelinePut(&foundPipeline)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Update active pipelines
		pipeline.GlobalActivePipelines.ReplaceByName(currentName, foundPipeline)
	}

	// Check if the periodic scheduling has been changed.
	if !stringSliceEqual(foundPipeline.PeriodicSchedules, p.PeriodicSchedules) {
		// We prevent side effects here and make sure
		// that no scheduling is already running.
		if foundPipeline.CronInst != nil {
			foundPipeline.CronInst.Stop()
		}
		foundPipeline.CronInst = cron.New()

		// Iterate over all cron schedules.
		for _, schedule := range p.PeriodicSchedules {
			err := foundPipeline.CronInst.AddFunc(schedule, func() {
				_, err := pp.deps.Scheduler.SchedulePipeline(&foundPipeline, gaia.StartReasonScheduled, []*gaia.Argument{})
				if err != nil {
					gaia.Cfg.Logger.Error("cannot schedule pipeline from periodic schedule", "error", err, "pipeline", foundPipeline)
					return
				}

				// Log scheduling information
				gaia.Cfg.Logger.Info("pipeline has been automatically scheduled by periodic scheduling:", "name", foundPipeline.Name)
			})

			if err != nil {
				return c.String(http.StatusBadRequest, err.Error())
			}
		}

		// Update pipeline in store
		err := storeService.PipelinePut(&foundPipeline)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Start schedule process.
		foundPipeline.CronInst.Start()

		// Update active pipelines
		pipeline.GlobalActivePipelines.Replace(foundPipeline)
	}

	// Check if docker option has been updated
	if p.Docker != foundPipeline.Docker {
		foundPipeline.Docker = p.Docker

		// Update pipeline in store
		err := storeService.PipelinePut(&foundPipeline)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Update active pipelines
		pipeline.GlobalActivePipelines.Replace(foundPipeline)
	}

	return c.String(http.StatusOK, "Pipeline has been updated")
}

// stringSliceEqual is a small helper function
// which determines if two string slices are equal.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// PipelineDelete accepts a pipeline id and deletes it from the
// store. It also removes the binary inside the pipeline folder.
// @Summary Delete a pipeline.
// @Description Accepts a pipeline id and deletes it from the store. It also removes the binary inside the pipeline folder.
// @Tags pipelines
// @Accept plain
// @Produce plain
// @Param pipelineid query string true "The ID of the pipeline."
// @Success 200 {string} string "Pipeline has been deleted"
// @Failure 400 {string} string "Error while deleting the pipeline"
// @Failure 404 {string} string "The pipeline with the given ID was not found"
// @Failure 500 {string} string "Internal error while deleting and removing the pipeline from store and disk"
// @Router /pipeline/{pipelineid} [delete]
func (pp *PipelineProvider) PipelineDelete(c echo.Context) error {
	storeService, _ := services.StorageService()
	pipelineIDStr := c.Param("pipelineid")

	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	var deletedPipelineIndex int
	for index, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			foundPipeline = p
			deletedPipelineIndex = index
			break
		}
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusNotFound, errPipelineNotFound.Error())
	}

	// Stop any schedulers running
	if ct := foundPipeline.CronInst; ct != nil {
		ct.Stop()
	}

	// Delete pipeline binary
	err = pipeline.DeleteBinary(foundPipeline)
	if err != nil {
		return c.String(http.StatusInternalServerError, errPipelineDelete.Error())
	}

	// Delete pipeline from store
	err = storeService.PipelineDelete(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Remove from active pipelines
	if err := pipeline.GlobalActivePipelines.Remove(deletedPipelineIndex); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Pipeline has been deleted")
}

// PipelineTrigger allows for a remote running of a pipeline.
// This endpoint does not require authentication. It will use a TOKEN
// that is specific to a pipeline. It can only be used by the `auto`
// user.
// @Summary Trigger a pipeline.
// @Description Using a trigger token, start a pipeline run. This endpoint does not require authentication.
// @Tags pipelines
// @Accept plain
// @Produce plain
// @Security
// @Param pipelineid query string true "The ID of the pipeline."
// @Param pipelinetoken query string true "The trigger token for this pipeline."
// @Success 200 {string} string "Trigger successful for pipeline: {pipelinename}"
// @Failure 400 {string} string "Error while triggering pipeline"
// @Failure 403 {string} string "Invalid trigger token"
// @Router /pipeline/{pipelineid}/{pipelinetoken}/trigger [post]
func (pp *PipelineProvider) PipelineTrigger(c echo.Context) error {
	err := pp.PipelineTriggerAuth(c)
	if err != nil {
		return c.String(http.StatusForbidden, "User rejected")
	}

	// Check here against the pipeline's token.
	pipelineIDStr := c.Param("pipelineid")
	pipelineToken := c.Param("pipelinetoken")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			foundPipeline = p
			break
		}
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusBadRequest, "Pipeline not found.")
	}

	if foundPipeline.TriggerToken != pipelineToken {
		return c.String(http.StatusForbidden, "Invalid remote trigger token.")
	}

	var args []*gaia.Argument
	_ = c.Bind(&args)
	pipelineRun, err := pp.deps.Scheduler.SchedulePipeline(&foundPipeline, gaia.StartReasonRemote, args)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	} else if pipelineRun != nil {
		return c.String(http.StatusOK, "Trigger successful for pipeline: "+pipelineIDStr)
	}

	return c.String(http.StatusBadRequest, "Failed to trigger pipeline run.")
}

// PipelineResetToken generates a new remote trigger token for a given
// pipeline.
// @Summary Reset trigger token.
// @Description Generates a new remote trigger token for a given pipeline.
// @Tags pipelines
// @Accept plain
// @Produce plain
// @Param pipelineid query string true "The ID of the pipeline."
// @Success 200 {string} string "Trigger successful for pipeline: {pipelinename}"
// @Failure 400 {string} string "Invalid pipeline id"
// @Failure 404 {string} string "Pipeline not found"
// @Failure 500 {string} string "Internal storage error"
// @Router /pipeline/{pipelineid}/reset-trigger-token [put]
func (pp *PipelineProvider) PipelineResetToken(c echo.Context) error {
	// Check here against the pipeline's token.
	pipelineIDStr := c.Param("pipelineid")

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			foundPipeline = p
			break
		}
	}

	if foundPipeline.Name == "" {
		return c.String(http.StatusNotFound, "Pipeline not found.")
	}

	foundPipeline.TriggerToken = security.GenerateRandomUUIDV5()
	s, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error getting store service.")
	}
	err = s.PipelinePut(&foundPipeline)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error while saving pipeline.")
	}
	return c.String(http.StatusOK, "Token successfully reset. To see, please open the pipeline's view.")
}

// PipelineTriggerAuth is a barrier before remote trigger which checks if
// the user is `auto`.
func (pp *PipelineProvider) PipelineTriggerAuth(c echo.Context) error {
	// check headers
	s, err := services.StorageService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error getting store service.")
	}
	auto, err := s.UserGet("auto")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Auto user not found.")
	}

	username, password, ok := c.Request().BasicAuth()
	if !ok {
		return c.String(http.StatusForbidden, "No authentication provided.")
	}
	if username != auto.Username || password != auto.TriggerToken {
		return c.String(http.StatusBadRequest, "Auto username or password did not match.")
	}
	return nil
}

// PipelineStart starts a pipeline by the given id.
// It accepts arguments for the given pipeline.
// Afterwards it returns the created/scheduled pipeline run.
// @Summary Start a pipeline.
// @Description Starts a pipeline with a given ID and arguments for that pipeline and returns created/scheduled status.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineid query string true "The ID of the pipeline."
// @Param args body gaia.Argument false "Optional arguments of the pipeline."
// @Success 200 {object} gaia.PipelineRun
// @Failure 400 {string} string "Various failures regarding starting the pipeline like: invalid id, invalid docker value and schedule errors"
// @Failure 404 {string} string "Pipeline not found"
// @Router /pipeline/{pipelineid}/start [post]
func (pp *PipelineProvider) PipelineStart(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	// Decode content
	content := echo.Map{}
	if err := c.Bind(&content); err != nil {
		return c.String(http.StatusBadRequest, "invalid content provided in request")
	}

	// Look for arguments.
	// We do not check for errors here cause arguments are optional.
	var args []*gaia.Argument
	_ = c.Bind(&args)

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}
	var docker bool
	if _, ok := content["docker"]; ok {
		docker, ok = content["docker"].(bool)
		if !ok {
			return c.String(http.StatusBadRequest, errWrongDockerValue.Error())
		}
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			foundPipeline = p
			break
		}
	}

	// Overwrite docker setting
	foundPipeline.Docker = docker

	if foundPipeline.Name != "" {
		pipelineRun, err := pp.deps.Scheduler.SchedulePipeline(&foundPipeline, gaia.StartReasonManual, args)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		} else if pipelineRun != nil {
			return c.JSON(http.StatusCreated, pipelineRun)
		}
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

// PipelinePull does a pull on the remote repository
// which contains the code for this pipeline. This is so the user
// won't have to wait for polling or a hook.
// @Summary Update the underlying repository of the pipeline.
// @Description Pull new code using the repository of the pipeline.
// @Tags pipelines
// @Accept plain
// @Produce plain
// @Param pipelineid query string true "The ID of the pipeline."
// @Success 200
// @Failure 400 {string} string
// @Router /pipeline/{pipelineid}/pull [post]
func (pp *PipelineProvider) PipelinePull(c echo.Context) error {
	pipelineIDStr := c.Param("pipelineid")

	// Decode content
	content := echo.Map{}
	if err := c.Bind(&content); err != nil {
		return c.String(http.StatusBadRequest, "invalid content provided in request")
	}

	// Convert string to int because id is int
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	var docker bool
	if _, ok := content["docker"]; ok {
		docker, ok = content["docker"].(bool)
		if !ok {
			return c.String(http.StatusBadRequest, errWrongDockerValue.Error())
		}
	}

	// Look up pipeline for the given id
	var foundPipeline gaia.Pipeline
	for _, p := range pipeline.GlobalActivePipelines.GetAll() {
		if p.ID == pipelineID {
			foundPipeline = p
			break
		}
	}

	// Overwrite docker setting
	foundPipeline.Docker = docker

	if foundPipeline.Name != "" {
		uniqueFolder, err := pipelinehelper.GetLocalDestinationForPipeline(foundPipeline)
		if err != nil {
			gaia.Cfg.Logger.Error("Pipeline type invalid", "type", foundPipeline.Type)
			return err
		}
		if foundPipeline.Repo == nil {
			gaia.Cfg.Logger.Error("Git repo is missing")
			return errors.New("no git repository for pipeline")
		}
		foundPipeline.Repo.LocalDest = uniqueFolder
		if err := pp.deps.PipelineService.UpdateRepository(&foundPipeline); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.NoContent(http.StatusOK)
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

type getAllWithLatestRun struct {
	Pipeline    gaia.Pipeline    `json:"p"`
	PipelineRun gaia.PipelineRun `json:"r"`
}

// PipelineGetAllWithLatestRun returns the latest of all registered pipelines
// included with the latest run.
// @Summary Returns the latest run.
// @Description Returns the latest of all registered pipelines included with the latest run.
// @Tags pipelines
// @Produce json
// @Success 200 {object} getAllWithLatestRun
// @Failure 500 {string} string "Internal error while getting latest run"
// @Router /pipeline/latest [get]
func (pp *PipelineProvider) PipelineGetAllWithLatestRun(c echo.Context) error {
	// Get all active pipelines
	storeService, _ := services.StorageService()
	pipelines := pipeline.GlobalActivePipelines.GetAll()

	// Iterate all pipelines
	var pipelinesWithLatestRun []getAllWithLatestRun
	for _, p := range pipelines {
		// Get the latest run by the given pipeline id
		run, err := storeService.PipelineGetLatestRun(p.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// Append run if one exists
		g := getAllWithLatestRun{}
		obscurePipelineData(&p)
		g.Pipeline = p
		if run != nil {
			g.PipelineRun = *run
		}

		// Append
		pipelinesWithLatestRun = append(pipelinesWithLatestRun, g)
	}

	return c.JSON(http.StatusOK, pipelinesWithLatestRun)
}

// PipelineCheckPeriodicSchedules validates the added periodic schedules.
// @Summary Returns the latest run.
// @Description Returns the latest of all registered pipelines included with the latest run.
// @Tags pipelines
// @Accept json
// @Produce json
// @Param schedules body []string true "A list of valid cronjob specs"
// @Success 200
// @Failure 400 {string} string "Bind error and schedule errors"
// @Router /pipeline/periodicschedules [post]
func (pp *PipelineProvider) PipelineCheckPeriodicSchedules(c echo.Context) error {
	var pSchedules []string
	if err := c.Bind(&pSchedules); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Create new test cron spec parser.
	cr := cron.New()

	// Check every cron entry.
	for _, entry := range pSchedules {
		if err := cr.AddFunc(entry, func() {}); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	// All entries are valid.
	return c.JSON(http.StatusOK, nil)
}

// obscurePipelineData obscures pipeline data from the given pipeline object
func obscurePipelineData(p *gaia.Pipeline) {
	p.ExecPath = ""
}
