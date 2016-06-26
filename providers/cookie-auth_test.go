package providers

import (
	"github.com/stretchr/testify/assert"
	// "net/http"
	// "net/http/httptest"
	"testing"
)

const (
	testFileName = "/tmp/idp_cookie_test.db3"
)

func TestCookieNew(t *testing.T) {
	assert := assert.New(t)

	c, err := NewCookieAuth(testFileName)
	assert.Nil(err)
	assert.NotNil(c)

}
