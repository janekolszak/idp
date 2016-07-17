package rethinkdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	RETHINKDB_ADDRESS = "localhost:28015"
	TEST_DATABASE     = "usersstoretest"
)

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
