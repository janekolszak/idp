package cookie

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testFileName = "/tmp/idp_cookie_test.db3"
)

var (
	users = []string{"user1", "user2", "user3", "user4"}
)

func TestAdd(t *testing.T) {
	assert := assert.New(t)

	store, err := NewDBStore("sqlite3", testFileName)
	assert.Nil(err)
	defer store.Close()

	c := CookieAuth{
		Store: store,
	}

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
