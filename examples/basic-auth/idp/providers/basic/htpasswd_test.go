package basic

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	htpasswdTestFileName = "/tmp/idp_htpasswd_test"
	htpasswdFileContents = `# Comment
							user1:hash
							user2:hash
							user3:hash
							user4:hash`
)

var htpasswdUsers = []string{"user1", "user2", "user3", "user4"}

func TestHtpasswd(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(htpasswdTestFileName, []byte(htpasswdFileContents), 0644)
	assert.Nil(err)

	// Load passwords
	var h Htpasswd

	err = h.Load(htpasswdTestFileName)
	assert.Nil(err)

	// t.Log("index:", len(h.(map[string]string)))
	for _, user := range htpasswdUsers {
		hash, err := h.Get(user)
		assert.Nil(err)
		assert.Equal(hash, "hash", "Bad password")
	}
}
