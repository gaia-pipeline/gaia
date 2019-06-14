package store

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/gaia-pipeline/gaia"
	uuid "github.com/satori/go.uuid"
)

func TestInit(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreInit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserGet(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreUserGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u, true)
	if err != nil {
		t.Fatal(err)
	}

	user, err := store.UserGet("userdoesnotexist")
	if err != nil {
		t.Fatal(err)
	}
	if user != nil {
		t.Fatalf("user object is not nil. We expected nil!")
	}

	user, err = store.UserGet(u.Username)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatalf("Expected user %v. Got nil.", u.Username)
	}
}

func TestUserPut(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreUserPut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserAuth(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreUserAuth")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u, true)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Password field has been cleared after last UserPut
	u.Password = "12345!#+21+"
	r, err := store.UserAuth(u, true)
	if err != nil {
		t.Fatal(err)
		return
	}
	if r == nil {
		t.Fatalf("user not found or password invalid")
	}

	u = &gaia.User{}
	u.Username = "userdoesnotexist"
	u.Password = "wrongpassword"
	r, err = store.UserAuth(u, true)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("Expected nil object here. User shouldnt be valid")
	}

	u = &gaia.User{}
	u.Username = "testuser"
	u.Password = "wrongpassword"
	r, err = store.UserAuth(u, true)
	if err == nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("Expected nil object here. User shouldnt be valid")
	}
}

func TestCreatePipelinePut(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreCreatePipelinePut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	p := &gaia.CreatePipeline{
		ID:         uuid.Must(uuid.NewV4(), nil).String(),
		Created:    time.Now(),
		StatusType: gaia.CreatePipelineRunning,
		Pipeline: gaia.Pipeline{
			Repo: &gaia.GitRepo{},
		},
	}
	err = store.CreatePipelinePut(p)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreatePipelineGet(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreCreatePipelineGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var putIDs []string
	var getIDs []string

	for i := 0; i < 3; i++ {
		p := gaia.CreatePipeline{
			ID: uuid.Must(uuid.NewV4(), nil).String(),
			Pipeline: gaia.Pipeline{
				Repo: &gaia.GitRepo{},
			},
		}
		err = store.CreatePipelinePut(&p)
		if err != nil {
			t.Fatal(err)
		}
		putIDs = append(putIDs, p.ID)
	}

	pList, err := store.CreatePipelineGet()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pList {
		getIDs = append(getIDs, p.ID)
	}

	if len(putIDs) != len(getIDs) {
		t.Fatalf("expected %d pipelines, got %d", len(putIDs), len(getIDs))
	}

	sort.Strings(putIDs)
	sort.Strings(getIDs)

	for i := range putIDs {
		if putIDs[i] != getIDs[i] {
			t.Fatalf("the IDs do not match")
		}
	}

}

func TestPipelinePut(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelinePut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	p := &gaia.Pipeline{
		Name:    "Test Pipeline",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	err = store.PipelinePut(p)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID == 0 {
		t.Fatal("ID is 0, it should be a unique ID")
	}

	id := p.ID
	err = store.PipelinePut(p)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID != id {
		t.Fatal("ID should not be generated if it is already present")
	}

}

func TestPipelineGet(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	p := &gaia.Pipeline{
		Name:    "Test Pipeline",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	err = store.PipelinePut(p)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := store.PipelineGet(p.ID)
	if err != nil {
		t.Fatal(err)
	}

	if ret.Name != p.Name || ret.Type != p.Type {
		log.Fatal("the values do not match")
	}

}

func TestPipelineGetByName(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetByName")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	p := &gaia.Pipeline{
		Name:    "Test Pipeline",
		Type:    gaia.PTypeGolang,
		Created: time.Now(),
	}

	err = store.PipelinePut(p)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := store.PipelineGetByName("Test Pipeline")
	if err != nil {
		t.Fatal(err)
	}

	if ret.Name != p.Name || ret.Type != p.Type {
		log.Fatal("the values do not match")
	}

}

func TestPipelineGetRunHighestID(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetRunHighestID")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	pipeline := &gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.PTypeGolang,
	}

	err = store.PipelinePut(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	pipelineRun1 := &gaia.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     gaia.RunRunning,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
		StartDate:  time.Now(),
	}
	err = store.PipelinePutRun(pipelineRun1)
	if err != nil {
		t.Fatal(err)
	}

	pipelineRun2 := &gaia.PipelineRun{
		ID:         2,
		PipelineID: 1,
		Status:     gaia.RunRunning,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
		StartDate:  time.Now(),
	}
	err = store.PipelinePutRun(pipelineRun2)
	if err != nil {
		t.Fatal(err)
	}

	runHighestID, err := store.PipelineGetRunHighestID(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	if runHighestID != pipelineRun2.ID {
		t.Fatalf("expected ID %d, got %d", pipelineRun2.ID, runHighestID)
	}

}

func TestPipelinePutRun(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelinePutRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	run := gaia.PipelineRun{
		UniqueID:     uuid.Must(uuid.NewV4(), nil).String(),
		ID:           1,
		PipelineID:   1,
		ScheduleDate: time.Now(),
		Status:       gaia.RunNotScheduled,
	}

	err = store.PipelinePutRun(&run)
	if err != nil {
		t.Fatal(err)
	}

}

func TestPipelineGetScheduled(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetScheduled")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	pipelineRun1 := &gaia.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     gaia.RunNotScheduled,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
	}
	err = store.PipelinePutRun(pipelineRun1)
	if err != nil {
		t.Fatal(err)
	}

	pipelineRun2 := &gaia.PipelineRun{
		ID:         2,
		PipelineID: 1,
		Status:     gaia.RunNotScheduled,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
	}
	err = store.PipelinePutRun(pipelineRun2)
	if err != nil {
		t.Fatal(err)
	}

	runs, err := store.PipelineGetScheduled(10)
	if err != nil {
		t.Fatal(err)
	}

	if len(runs) != 2 {
		t.Fatalf("expected %d runs, got %d", 2, len(runs))
	}

	runs, err = store.PipelineGetScheduled(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(runs) != 1 {
		t.Fatalf("expected %d runs, got %d", 1, len(runs))
	}

}

func TestPipelineGetRunByPipelineIDAndID(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetRunByPipelineIDAndID")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	run := gaia.PipelineRun{
		UniqueID:     uuid.Must(uuid.NewV4(), nil).String(),
		ID:           1,
		PipelineID:   1,
		ScheduleDate: time.Now(),
		Status:       gaia.RunNotScheduled,
	}

	err = store.PipelinePutRun(&run)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := store.PipelineGetRunByPipelineIDAndID(run.PipelineID, run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if ret.UniqueID != run.UniqueID {
		t.Fatal("the unique IDs do not match")
	}
}

func TestPipelineGetAllRuns(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetAllRuns")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	var putUniqueIDs []string
	var getUniqueIDs []string

	pipeline := &gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.PTypeGolang,
	}

	err = store.PipelinePut(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		p := &gaia.PipelineRun{
			ID:         i,
			PipelineID: 1,
			Status:     gaia.RunNotScheduled,
			UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
		}
		err = store.PipelinePutRun(p)
		if err != nil {
			t.Fatal(err)
		}
		putUniqueIDs = append(putUniqueIDs, p.UniqueID)
	}

	pipelineRuns, err := store.PipelineGetAllRuns()
	if err != nil {
		t.Fatal(err)
	}

	for _, pr := range pipelineRuns {
		getUniqueIDs = append(getUniqueIDs, pr.UniqueID)
	}

	if len(putUniqueIDs) != len(getUniqueIDs) {
		t.Fatalf("expected %d runs, got %d", len(putUniqueIDs), len(getUniqueIDs))
	}

	sort.Strings(putUniqueIDs)
	sort.Strings(getUniqueIDs)

	for i := range putUniqueIDs {
		if putUniqueIDs[i] != getUniqueIDs[i] {
			t.Fatalf("the unique IDs do not match")
		}
	}

}

func TestPipelineGetLatestRun(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStorePipelineGetLatestRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	pipeline := &gaia.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: gaia.PTypeGolang,
	}

	err = store.PipelinePut(pipeline)
	if err != nil {
		t.Fatal(err)
	}

	pipelineRun1 := &gaia.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     gaia.RunRunning,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
		StartDate:  time.Now(),
	}
	err = store.PipelinePutRun(pipelineRun1)
	if err != nil {
		t.Fatal(err)
	}

	pipelineRun2 := &gaia.PipelineRun{
		ID:         2,
		PipelineID: 1,
		Status:     gaia.RunRunning,
		UniqueID:   uuid.Must(uuid.NewV4(), nil).String(),
		StartDate:  time.Now(),
	}
	err = store.PipelinePutRun(pipelineRun2)
	if err != nil {
		t.Fatal(err)
	}

	latestRun, err := store.PipelineGetLatestRun(1)
	if err != nil {
		t.Fatal(err)
	}

	if latestRun.UniqueID != pipelineRun2.UniqueID {
		t.Fatalf("expected unique ID %s, got %s", pipelineRun2.UniqueID, latestRun.UniqueID)
	}
}

func TestUserPermissionsPutGetDelete(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestUserPermissionsPutGetDelete")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	gaia.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Put user permissions
	perm := &gaia.UserPermission{
		Username: "michel",
		Groups:   []string{"my-group"},
		Roles:    []string{"my-role"},
	}
	if err := store.UserPermissionsPut(perm); err != nil {
		t.Fatal(err)
	}

	// Read user permissions
	storePerm, err := store.UserPermissionsGet("michel")
	if err != nil {
		t.Fatal(err)
	}

	// Validate
	if storePerm == nil {
		t.Fatal("expected user permissions but it is nil")
	}
	if storePerm.Username != "michel" {
		t.Fatalf("expected '%s' but got '%s'", "michel", storePerm.Username)
	}
	if len(storePerm.Roles) != 1 {
		t.Fatalf("expected %d but got %d", 1, len(storePerm.Roles))
	}
	if storePerm.Roles[0] != "my-role" {
		t.Fatalf("expected '%s' but got '%s'", "my-role", storePerm.Roles[0])
	}
	if len(storePerm.Groups) != 1 {
		t.Fatalf("expected %d but got %d", 1, len(storePerm.Groups))
	}
	if storePerm.Groups[0] != "my-group" {
		t.Fatalf("expected '%s' but got '%s'", "m-group", storePerm.Groups[0])
	}

	// Delete
	if err := store.UserPermissionsDelete("michel"); err != nil {
		t.Fatal(err)
	}

	// Validate
	storePerm, err = store.UserPermissionsGet("michel")
	if err != nil {
		t.Fatal(err)
	}
	if storePerm != nil {
		t.Fatalf("expected nil object but it is: %#v", storePerm)
	}

}
