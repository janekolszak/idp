package cookie

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	testFileName = "/tmp/idp_cookie_test.db3"
)

var (
	users = []string{"user1", "user2", "user3", "user4"}
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	{
		c, err := NewCookieAuth(testFileName)
		assert.Nil(err)
		defer c.Close()
		assert.NotNil(c)
	}

	{
		c, err := NewCookieAuth(testFileName)
		assert.Nil(err)
		defer c.Close()
		assert.NotNil(c)
	}

	os.Remove(testFileName)

	{
		c, err := NewCookieAuth(testFileName)
		assert.Nil(err)
		defer c.Close()
		assert.NotNil(c)
	}

}

func TestAdd(t *testing.T) {
	assert := assert.New(t)

	c, err := NewCookieAuth(testFileName)
	assert.Nil(err)
	defer c.Close()
	assert.NotNil(c)

	for _, user := range users {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/", nil)
		err = c.Add(w, r, user)
		assert.Nil(err)

		// Should check true if this cookie appears
		requestToVerify := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}
		readUser, err := c.Check(requestToVerify)
		assert.Nil(err)
		assert.Equal(user, readUser)
	}
}
