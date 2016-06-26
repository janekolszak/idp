package helpers

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginCookieNew(t *testing.T) {
	assert := assert.New(t)

	l, err := NewLoginCookie("", "")
	assert.Nil(err)
	assert.NotEqual(l.Selector, "")

	l, err = NewLoginCookie("1234", "")
	assert.Equal(l.Selector, "1234")
}

func TestLoginCookieValidator(t *testing.T) {
	assert := assert.New(t)

	l, err := NewLoginCookie("1", "")
	assert.Nil(err)

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

	l, err := NewLoginCookie("1", "remember")
	assert.Nil(err)

	_, err = l.GenerateValidator()
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

	la, err := NewLoginCookie("1", "remember")
	assert.Nil(err)

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
