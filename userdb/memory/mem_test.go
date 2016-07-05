package memory

import (
	"io/ioutil"
	"testing"

	"github.com/janekolszak/idp/core"
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

func TestNew(t *testing.T) {
	assert := assert.New(t)

	s, err := NewMemStore()
	assert.Nil(err)

	err = s.Check("bob", "")
	assert.Equal(core.ErrorNoSuchUser, err)

	err = s.Add("bob", "bob123")
	assert.Nil(err)

	err = s.Check("bob", "")
	assert.Equal(core.ErrorAuthenticationFailure, err)

	err = s.Check("bob", "bob123")
	assert.Nil(err)
}

func TestHtpasswd(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(htpasswdTestFileName, []byte(htpasswdFileContents), 0644)
	assert.Nil(err)

	s, err := NewMemStore()

	// Load passwords
	err = s.LoadHtpasswd(htpasswdTestFileName)
	assert.Nil(err)

	// t.Log("index:", len(h.(map[string]string)))
	for _, user := range htpasswdUsers {
		assert.Equal("hash", s.hashes[user])
	}
}
