package store

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"time"

	bolt "github.com/coreos/bbolt"

	"github.com/gaia-pipeline/gaia"
	"github.com/gaia-pipeline/gaia/helper/resourcehelper"
	"github.com/gaia-pipeline/gaia/security"
)

var (
	// Name of the bucket where we store user objects
	userBucket = []byte("Users")

	// Where we store all users permissions
	userPermsBucket = []byte("UserPermissions")

	// Name of the bucket where we store information about pipelines
	pipelineBucket = []byte("Pipelines")

	// Name of the bucket where we store information about pipelines
	// which are not yet compiled (create pipeline)
	createPipelineBucket = []byte("CreatePipelines")

	// Name of the bucket where we store all pipeline runs.
	pipelineRunBucket = []byte("PipelineRun")

	// Name of the bucket where we store information about settings
	settingsBucket = []byte("Settings")

	// Name of the bucket where we store all worker.
	workerBucket = []byte("Worker")

	// SHA pair bucket.
	shaPairBucket = []byte("SHAPair")

	authPolicyResources = []byte("authorization.rbac.resources")
	authPolicyBindings  = []byte("authorization.rbac.bindings")
)

const (
	// Username and password of the first admin user
	adminUsername = "admin"
	adminPassword = "admin"
	autoUsername  = "auto"
	autoPassword  = "auto"

	// Bolt database file name
	boltDBFileName = "gaia.db"
)

// BoltStore represents the access type for store
type BoltStore struct {
	db             *bolt.DB
	rbacMarshaller resourcehelper.Marshaller
}

// GaiaStore is the interface that defines methods needed to store
// pipeline and user related information.
type GaiaStore interface {
	Init(dataPath string) error
	Close() error
	CreatePipelinePut(createPipeline *gaia.CreatePipeline) error
	CreatePipelineGet() (listOfPipelines []gaia.CreatePipeline, err error)
	PipelinePut(pipeline *gaia.Pipeline) error
	PipelineGet(id int) (pipeline *gaia.Pipeline, err error)
	PipelineGetByName(name string) (pipline *gaia.Pipeline, err error)
	PipelineGetRunHighestID(pipeline *gaia.Pipeline) (id int, err error)
	PipelinePutRun(r *gaia.PipelineRun) error
	PipelineGetScheduled(limit int) ([]*gaia.PipelineRun, error)
	PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*gaia.PipelineRun, error)
	PipelineGetAllRuns() ([]gaia.PipelineRun, error)
	PipelineGetAllRunsByPipelineID(pipelineID int) ([]gaia.PipelineRun, error)
	PipelineGetLatestRun(pipelineID int) (*gaia.PipelineRun, error)
	PipelineGetRunByID(runID string) (*gaia.PipelineRun, error)
	PipelineDelete(id int) error
	PipelineRunDelete(uniqueID string) error
	UserPut(u *gaia.User, encryptPassword bool) error
	UserAuth(u *gaia.User, updateLastLogin bool) (*gaia.User, error)
	UserGet(username string) (*gaia.User, error)
	UserGetAll() ([]gaia.User, error)
	UserDelete(u string) error
	UserPermissionsPut(perms *gaia.UserPermission) error
	UserPermissionsGet(username string) (*gaia.UserPermission, error)
	UserPermissionsDelete(username string) error
	SettingsPut(config *gaia.StoreConfig) error
	SettingsGet() (*gaia.StoreConfig, error)
	WorkerPut(w *gaia.Worker) error
	WorkerGetAll() ([]*gaia.Worker, error)
	WorkerDelete(id string) error
	WorkerDeleteAll() error
	WorkerGet(id string) (*gaia.Worker, error)
	UpsertSHAPair(pair gaia.SHAPair) error
	GetSHAPair(pipelineID int) (bool, gaia.SHAPair, error)
	RBACPolicyResourcePut(spec gaia.RBACPolicyResourceV1) error
	RBACPolicyResourceGet(name string) (gaia.RBACPolicyResourceV1, error)
	RBACPolicyBindingsPut(username string, policy string) error
	RBACPolicyBindingsGet(username string) (map[string]interface{}, error)
}

// Compile time interface compliance check for BoltStore. If BoltStore
// wouldn't implement GaiaStore this line wouldn't compile.
var _ GaiaStore = (*BoltStore)(nil)

// NewBoltStore creates a new instance of store.
func NewBoltStore() *BoltStore {
	s := &BoltStore{
		rbacMarshaller: resourcehelper.NewMarshaller(),
	}

	return s
}

// Init creates the data folder if not exists,
// generates private key and bolt database.
// This should be called only once per database
// because bolt holds a lock on the database file.
func (s *BoltStore) Init(dataPath string) error {
	// Open connection to bolt database
	path := filepath.Join(dataPath, boltDBFileName)
	// Give boltdb 5 seconds to try and open up a db file.
	// If another process is already holding that file, this will time-out.
	db, err := bolt.Open(path, gaia.Cfg.Bolt.Mode, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return err
	}
	s.db = db

	// Setup database
	return s.setupDatabase()
}

// Close closes the active boltdb connection.
func (s *BoltStore) Close() error {
	return s.db.Close()
}

type setup struct {
	bs  *BoltStore
	err error
}

// Create bucket if not exists function
func (s *setup) update(bucketName []byte) {
	// update is a no-op in case there was already an error
	if s.err != nil {
		return
	}
	c := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	}
	s.err = s.bs.db.Update(c)
}

// setupDatabase create all buckets in the db.
// Additionally, it makes sure that the admin user exists.
func (s *BoltStore) setupDatabase() error {
	// Create bucket if not exists function
	setP := &setup{
		bs:  s,
		err: nil,
	}

	// Make sure buckets exist
	setP.update(userBucket)
	setP.update(userPermsBucket)
	setP.update(pipelineBucket)
	setP.update(createPipelineBucket)
	setP.update(pipelineRunBucket)
	setP.update(settingsBucket)
	setP.update(workerBucket)
	setP.update(shaPairBucket)
	setP.update(authPolicyResources)
	setP.update(authPolicyBindings)

	if setP.err != nil {
		return setP.err
	}

	// Make sure that the user "admin" does exist
	admin, err := s.UserGet(adminUsername)
	if err != nil {
		return err
	}

	// Create admin user if we cannot find it
	if admin == nil {
		err = s.UserPut(&gaia.User{
			DisplayName: adminUsername,
			Username:    adminUsername,
			Password:    adminPassword,
		}, true)

		if err != nil {
			return err
		}
	}

	err = s.CreatePermissionsIfNotExisting()
	if err != nil {
		return err
	}

	u, err := s.UserGet(autoUsername)
	if err != nil {
		return err
	}

	if u == nil {
		triggerToken := security.GenerateRandomUUIDV5()
		auto := gaia.User{
			DisplayName:  "Auto User",
			JwtExpiry:    0,
			Password:     autoPassword,
			Tokenstring:  "",
			TriggerToken: triggerToken,
			Username:     autoUsername,
			LastLogin:    time.Now(),
		}
		err = s.UserPut(&auto, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
