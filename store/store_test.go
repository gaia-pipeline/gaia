package store

import (
	"fmt"
	"os"
	"testing"

	"github.com/michelvocks/gaia"
)

var store *Store
var config *gaia.Config

func TestMain(m *testing.M) {
	store = NewStore()
	config = &gaia.Config{}
	config.DataPath = "data"
	config.Bolt.Path = "test.db"
	config.Bolt.Mode = 0600

	r := m.Run()

	// cleanup
	err := os.Remove("data")
	if err != nil {
		fmt.Printf("cannot remove data folder: %s\n", err.Error())
		r = 1
	}
	os.Exit(r)
}

func TestInit(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup
	err = os.Remove("data/test.db")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserGet(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u)
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

	// cleanup
	err = os.Remove("data/test.db")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserPut(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup
	err = os.Remove("data/test.db")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserAuth(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserPut(u)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Password field has been cleared after last UserPut
	u.Password = "12345!#+21+"
	r, err := store.UserAuth(u)
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
	r, err = store.UserAuth(u)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("Expected nil object here. User shouldnt be valid")
	}

	u = &gaia.User{}
	u.Username = "testuser"
	u.Password = "wrongpassword"
	r, err = store.UserAuth(u)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		t.Fatalf("Expected nil object here. User shouldnt be valid")
	}

	// cleanup
	err = os.Remove("data/test.db")
	if err != nil {
		t.Fatal(err)
	}
}
