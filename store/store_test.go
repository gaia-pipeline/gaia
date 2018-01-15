package store

import (
	"os"
	"testing"

	"github.com/michelvocks/gaia"
)

var store *Store
var config *gaia.Config

func TestMain(m *testing.M) {
	store = NewStore()
	config = &gaia.Config{}
	config.Bolt.Path = "test.db"
	config.Bolt.Mode = 0600

	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup
	err = os.Remove("test.db")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserUpdate(t *testing.T) {
	err := store.Init(config)
	if err != nil {
		t.Fatal(err)
	}

	u := &gaia.User{}
	u.Username = "testuser"
	u.Password = "12345!#+21+"
	u.DisplayName = "Test"
	err = store.UserUpdate(u)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup
	err = os.Remove("test.db")
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
	err = store.UserUpdate(u)
	if err != nil {
		t.Fatal(err)
		return
	}

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
	err = os.Remove("test.db")
	if err != nil {
		t.Fatal(err)
	}
}
