package cookie

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/helpers"

	"net/http"
)

const (
	rememberMeCookieName = "remember"
)

type CookieAuth struct {
	Store Store
}

func (c *CookieAuth) Check(r *http.Request) (user string, err error) {
	l, err := helpers.GetLoginCookie(r, rememberMeCookieName)
	if err != nil {
		return
	}

	// TODO: Validate selector, shouldn't be too long etc.

	var hash string
	user, hash, err = c.Store.Get(l.Selector)
	if err != nil {
		return
	}

	if !l.Check(hash) {
		err = core.ErrorBadRequest
	}

	return
}

// TODO: Selector should be created by the database, here it's automatically generated
func (c *CookieAuth) Add(w http.ResponseWriter, r *http.Request, user string) (err error) {
	l, err := helpers.NewLoginCookie("", rememberMeCookieName)
	if err != nil {
		return
	}

	hash, err := l.GenerateValidator()
	if err != nil {
		return
	}

	//TODO: Reuse selector

	// First save to the database
	err = c.Store.Upsert(l.Selector, user, hash)
	if err != nil {
		return
	}

	// Then save to the cookie
	err = l.Save(w, r)
	return
}

func (c *CookieAuth) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	return nil
}

func (c *CookieAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
