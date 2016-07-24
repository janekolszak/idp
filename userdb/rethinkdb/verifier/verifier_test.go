package verifier

import (
	"os"
	"testing"

	"github.com/janekolszak/idp/userdb/rethinkdb/store"
	"github.com/stretchr/testify/assert"
	r "gopkg.in/dancannon/gorethink.v2"
)

const (
	RETHINKDB_ADDRESS = "localhost:28015"
	TEST_DATABASE     = "verifyuserstest"
	TEST_USER_ID      = "testUserID"
	TEST_USER_NAME    = "testUser Name"
	TEST_USER_EMAIL   = "joe@doe"
)

var (
	session  *r.Session
	testUser = &store.User{
		FirstName: "Joe",
		LastName:  "Doe",
		Username:  "joe",
		Email:     TEST_USER_EMAIL,
	}
	testUserPassword = "testPassword"
)

func Cleanup() {
	testUser.ID = ""
	r.DB(TEST_DATABASE).TableDrop("verifyEmails").Exec(session)
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

	SetupWorkerTests()

	os.Exit(m.Run())
}

func TestNewVerifier(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	verifier, err := NewVerifier(session)
	assert.Nil(err)
	assert.NotNil(verifier)
}

func TestVerifierPush(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	verifier, err := NewVerifier(session)
	assert.Nil(err)
	assert.NotNil(verifier)

	id, err := verifier.Push(TEST_USER_ID, TEST_USER_NAME, TEST_USER_EMAIL)
	assert.Nil(err)
	assert.NotEqual(id, "")

	count, err := verifier.Count()
	assert.Nil(err)
	assert.Equal(int(count), 1, "Should be equal")
}

func TestVerifierVerify(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	verifier, err := NewVerifier(session)
	assert.Nil(err)
	assert.NotNil(verifier)

	code, err := verifier.Push(TEST_USER_ID, TEST_USER_NAME, TEST_USER_EMAIL)
	assert.Nil(err)
	assert.NotEqual(code, "")

	count, err := verifier.Count()
	assert.Nil(err)
	assert.Equal(int(count), 1, "Should be equal")

	userID, err := verifier.Verify(code)
	assert.Nil(err)
	assert.Equal(userID, TEST_USER_ID)

	count, err = verifier.Count()
	assert.Nil(err)
	assert.Equal(int(count), 0, "Should be equal")
	// id, err = verifier.Push(TEST_USER_ID,TEST_USER_NAME, TEST_USER_EMAIL)
	// assert.Nil(err)
	// assert.NotEqual(id, "")
}

// func TestVerify(t *testing.T) {
// 	assert := assert.New(t)
// 	Cleanup()

// 	store, err := NewStore(session)
// 	assert.Nil(err)
// 	assert.NotNil(store)

// 	id, err := store.Insert(testUser, testUserPassword)
// 	assert.Nil(err)
// 	assert.NotEqual(id, "")

// 	// Second insert with the same username or email should fail
// 	id, err = store.Insert(testUser, testUserPassword)
// 	assert.NotNil(err)
// }
