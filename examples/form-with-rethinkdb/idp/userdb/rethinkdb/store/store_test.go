package store

import (
	"os"
	"testing"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"

	"github.com/stretchr/testify/assert"
	r "gopkg.in/dancannon/gorethink.v2"
)

const (
	RETHINKDB_ADDRESS = "localhost:28015"
	TEST_DATABASE     = "usersstoretest"
)

var (
	session  *r.Session
	testUser = &userdb.User{
		FirstName: "Joe",
		LastName:  "Doe",
		Username:  "joe",
		Email:     "joe@example.com",
	}
	testUserPassword = "testPassword"
)

func Cleanup() error {
	testUser.ID = ""
	return r.DBDrop(TEST_DATABASE).Exec(session)
}

func TestMain(m *testing.M) {
	var err error
	session, err = r.Connect(r.ConnectOpts{
		Address:  RETHINKDB_ADDRESS,
		Database: TEST_DATABASE,
	})

	if err != nil {
		panic(err)
	}
	defer session.Close()

	os.Exit(m.Run())
}

func TestNewStore(t *testing.T) {
	assert := assert.New(t)

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	// Second insert with the same username or email should fail
	id, err = store.Insert(testUser, testUserPassword)
	assert.NotNil(err)
}

func TestGetWithUsername(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	// Get the user having his username
	user, err := store.GetWithUsername(testUser.Username)
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal(user.FirstName, testUser.FirstName)
	assert.Equal(user.LastName, testUser.LastName)
	assert.Equal(user.Email, testUser.Email)
}

func TestGetWithID(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	// Get the user having his internal id
	user, err := store.GetWithID(id)
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal(user.FirstName, testUser.FirstName)
	assert.Equal(user.LastName, testUser.LastName)
	assert.Equal(user.Email, testUser.Email)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	userUpdated := &userdb.User{
		ID:        id,
		FirstName: "Ferris",
		LastName:  "Bueller",
		Username:  "righteousdude",
		Email:     "ferris@example.com",
	}

	err = store.Update(userUpdated)
	assert.Nil(err)

	// Check user data changed
	user, err := store.GetWithID(id)
	assert.NotNil(user)
	assert.Equal(user.FirstName, userUpdated.FirstName)
	assert.Equal(user.LastName, userUpdated.LastName)
	assert.Equal(user.Email, userUpdated.Email)

	// Password intact
	err = store.Check(user.Username, testUserPassword)
	assert.Nil(err)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	err = store.DeleteWithID(id)
	assert.Nil(err)

	_, err = store.GetWithID(id)
	assert.NotNil(err)
}

func TestSetIsVerifiedWithID(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	user, err := store.GetWithID(id)
	assert.Nil(err)
	assert.Equal(user.IsVerified, false)

	err = store.SetIsVerifiedWithID(id)
	assert.Nil(err)

	user, err = store.GetWithID(id)
	assert.Nil(err)
	assert.Equal(user.IsVerified, true)
}

func TestCheck(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(Cleanup())

	store, err := NewStore(session)
	assert.Nil(err)
	assert.NotNil(store)

	// No user
	err = store.Check(testUser.Username, testUserPassword)
	assert.Equal(err, core.ErrorAuthenticationFailure)

	id, err := store.Insert(testUser, testUserPassword)
	assert.Nil(err)
	assert.NotEqual(id, "")

	// Good password
	err = store.Check(testUser.Username, testUserPassword)
	assert.Nil(err)

	// Bad password
	err = store.Check(testUser.Username, testUserPassword+"stuff")
	assert.Equal(err, core.ErrorAuthenticationFailure)
}
