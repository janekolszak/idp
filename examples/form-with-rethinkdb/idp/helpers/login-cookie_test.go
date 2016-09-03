package helpers

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginCookieValidator(t *testing.T) {
	assert := assert.New(t)

	l := LoginCookie{}

	a, err := l.GenerateValidator()
	assert.Nil(err)

	assert.True(l.Check(a))

	b, err := l.GenerateValidator()
	assert.Nil(err)
	assert.True(l.Check(b))
	assert.False(l.Check(a))
}

func TestLoginCookieSave(t *testing.T) {
	assert := assert.New(t)

	l := LoginCookie{
		Selector:   "1",
		CookieName: "remember",
	}

	_, err := l.GenerateValidator()
	assert.Nil(err)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	err = l.Save(w, r)
	assert.Nil(err)
	assert.NotEqual(w.HeaderMap["Set-Cookie"], "")
}

func TestLoginCookieGet(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)

	la := LoginCookie{
		Selector:   "1",
		CookieName: "remember",
	}

	hash, err := la.GenerateValidator()
	assert.Nil(err)

	err = la.Save(w, r)
	assert.Nil(err)

	request := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}

	lb, err := GetLoginCookie(request, "remember")
	assert.Nil(err)
	assert.Equal(la.Selector, lb.Selector)
	assert.Equal(la.Validator, lb.Validator)

	assert.True(la.Check(hash))
	assert.True(lb.Check(hash))
}

func TestLoginCookieNoCookieGet(t *testing.T) {
	assert := assert.New(t)
	r, err := http.NewRequest("GET", "/", nil)

	l, err := GetLoginCookie(r, "remember")
	assert.Nil(l)
	assert.NotNil(err)
}
