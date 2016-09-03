package cookie

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	RETHINKDB_ADDRESS = "localhost:28015"
	TEST_DATABASE     = "cookietests"
)

func TestNewRethinkDBStore(t *testing.T) {
	assert := assert.New(t)

	store, err := NewRethinkDBStore(RETHINKDB_ADDRESS, TEST_DATABASE)
	assert.Nil(err)
	assert.NotNil(store)
}

func TestRethinkDBStoreSimple(t *testing.T) {
	assert := assert.New(t)

	store, err := NewRethinkDBStore(RETHINKDB_ADDRESS, TEST_DATABASE)
	assert.Nil(err)
	assert.NotNil(store)

	now := time.Now()
	selector, err := store.Insert("user", "hash", now)
	assert.Nil(err)
	assert.NotEqual(selector, "")

	user, hash, expiration, err := store.Get(selector)
	assert.Nil(err)
	assert.Equal(user, "user")
	assert.Equal(hash, "hash")
	assert.Equal(expiration.Day(), now.Day())

	now = time.Now()
	err = store.Update(selector, "user", "hash2", now)
	assert.Nil(err)

	user, hash, expiration, err = store.Get(selector)
	assert.Nil(err)
	assert.Equal(user, "user")
	assert.Equal(hash, "hash2")
	assert.Equal(expiration.Day(), now.Day())

	err = store.DeleteAll()
	assert.Nil(err)

}

func TestRethinkDBStoreDeleteSelector(t *testing.T) {
	assert := assert.New(t)

	store, err := NewRethinkDBStore(RETHINKDB_ADDRESS, TEST_DATABASE)

	selector, err := store.Insert("user", "hash", time.Now())
	assert.Nil(err)
	assert.NotEqual(selector, "")

	_, _, _, err = store.Get(selector)
	assert.Nil(err)

	err = store.DeleteSelector(selector)
	assert.Nil(err)

	_, _, _, err = store.Get(selector)
	assert.NotNil(err)
}

func TestRethinkDBStoreDeleteUser(t *testing.T) {
	assert := assert.New(t)

	store, err := NewRethinkDBStore(RETHINKDB_ADDRESS, TEST_DATABASE)

	selector, err := store.Insert("user", "hash", time.Now())
	assert.Nil(err)
	assert.NotEqual(selector, "")

	_, _, _, err = store.Get(selector)
	assert.Nil(err)

	err = store.DeleteUser("user")
	assert.Nil(err)

	_, _, _, err = store.Get(selector)
	assert.NotNil(err)
}
