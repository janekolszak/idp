package basic

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testFileName = "/tmp/idp_htpasswd_test"
	htpasswdFile = `# Comment
	                user1:hash
	                user2:hash
	                user3:hash
	                user4:hash`
)

var users = []string{"user1", "user2", "user3", "user4"}

func TestHtpasswd(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(testFileName, []byte(htpasswdFile), 0644)
	assert.Nil(err)

	// Load passwords
	var h Htpasswd

	err = h.Load(testFileName)
	assert.Nil(err)

	// t.Log("index:", len(h.(map[string]string)))
	for _, user := range users {
		hash, err := h.Get(user)
		assert.Nil(err)
		assert.Equal(hash, "hash", "Bad password")
	}
}
