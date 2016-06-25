package helpers

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRememberMeNew(t *testing.T) {
	assert := assert.New(t)

	rm, err := NewRememberMe("", "")
	assert.Nil(err)
	assert.NotEqual(rm.Selector, "")

	rm, err = NewRememberMe("1234", "")
	assert.Equal(rm.Selector, "1234")
}

func TestRememberMeValidator(t *testing.T) {
	assert := assert.New(t)

	rm, err := NewRememberMe("1", "")
	assert.Nil(err)

	a, err := rm.GenerateValidator()
	assert.Nil(err)

	assert.True(rm.Check(a))

	b, err := rm.GenerateValidator()
	assert.Nil(err)
	assert.True(rm.Check(b))
	assert.False(rm.Check(a))
}

func TestRememberMeSave(t *testing.T) {
	assert := assert.New(t)

	rm, err := NewRememberMe("1", "remember")
	assert.Nil(err)

	_, err = rm.GenerateValidator()
	assert.Nil(err)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	err = rm.Save(w, r)
	assert.Nil(err)
	assert.NotEqual(w.HeaderMap["Set-Cookie"], "")
}

func TestRememberMeGet(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)

	rma, err := NewRememberMe("1", "remember")
	assert.Nil(err)

	hash, err := rma.GenerateValidator()
	assert.Nil(err)

	err = rma.Save(w, r)
	assert.Nil(err)

	request := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}

	rmb, err := GetRememberMe(request, "remember")
	assert.Nil(err)
	assert.Equal(rma.Selector, rmb.Selector)
	assert.Equal(rma.Validator, rmb.Validator)

	assert.True(rma.Check(hash))
	assert.True(rmb.Check(hash))
}
