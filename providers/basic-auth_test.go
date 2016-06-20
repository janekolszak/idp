package providers

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testFileName = "/tmp/idp_basicauth_test"
	htpasswdFile = `# Comment
	                user1:$2y$05$oHxU0ruGKexixZrm6uDFMOtSFKQIpesi.iW/K6CY/2fcwwZ7F13TK
	                user2:$2y$05$9ebkCN627LXsUqk9QTUQ/.bVRR1A5l1mRBprqioy7n52k11wnKvbO
	                user3:$2y$05$.dhsRo567Am71lJGYRTHROQkPSzRCCfr8Bu5CuOryW5TslPf8fp0i
	                user4:$2y$05$44nVGKd8jOGbequSveEJU.N5IKY1OLnBR0Ow0899ZpOZJe5LZrQRS`
)

var users = []string{"user1", "user2", "user3", "user4"}

func TestCheck(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(testFileName, []byte(htpasswdFile), 0644)
	assert.Nil(err)

	// Create the provider
	provider, err := MakeBasicAuth(testFileName, "example.com")
	assert.Nil(err)

	for _, user := range users {
		r, err := http.NewRequest("GET", "/", nil)
		assert.Nil(err)

		r.SetBasicAuth(user, "password")
		err = provider.Check(r)
		assert.Nil(err)

		r.SetBasicAuth(user, "badpassword")
		err = provider.Check(r)
		assert.NotNil(err)
	}
}

func TestNoHeader(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(testFileName, []byte(htpasswdFile), 0644)
	assert.Nil(err)

	// Create the provider
	provider, err := MakeBasicAuth(testFileName, "example.com")
	assert.Nil(err)

	r := &http.Request{}

	err = provider.Check(r)
	assert.NotNil(err)
}

func TestRespond(t *testing.T) {
	assert := assert.New(t)

	// Prepare file
	err := ioutil.WriteFile(testFileName, []byte(htpasswdFile), 0644)
	assert.Nil(err)

	// Create the provider
	provider, err := NewBasicAuth(testFileName, "example.com")
	assert.Nil(err)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)

	err = provider.Respond(w, r)
	assert.Nil(err)
	assert.Equal(w.HeaderMap["Www-Authenticate"], []string{`Basic realm="example.com"`}, "Bad header")
	assert.Equal(w.HeaderMap["Content-Type"], []string{`text/plain; charset=utf-8`}, "Bad header")
}
