package verifier

import (
	"os"
	"testing"

	"github.com/janekolszak/idp/userdb/rethinkdb"
	"github.com/stretchr/testify/assert"
	r "gopkg.in/dancannon/gorethink.v2"
)

const (
	RETHINKDB_ADDRESS = "localhost:28015"
	TEST_DATABASE     = "verifyuserstest"
	TEST_USER_ID      = "testUserID"
)

var (
	session  *r.Session
	testUser = &rethinkdb.User{
		FirstName: "Joe",
		LastName:  "Doe",
		Username:  "joe",
		Email:     "joe@example.com",
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

	id, err := verifier.PushVerification("userID", "joe@example.com")
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

	code, err := verifier.PushVerification(TEST_USER_ID, "joe@example.com")
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
	// id, err = verifier.PushVerification(TEST_USER_ID, "joe@example.com")
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
