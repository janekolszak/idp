package basic

import (
	"fmt"
	"net/http"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"
)

// Basic Authentication checker.
// Expects Storage to return plain text passwords
type BasicAuth struct {
	UserStore userdb.Store
	Realm     string
}

func NewBasicAuth(users userdb.Store, realm string) (*BasicAuth, error) {
	b := BasicAuth{
		UserStore: users,
		Realm:     realm,
	}
	return &b, nil
}

func (c *BasicAuth) Check(r *http.Request) (user string, err error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		user = ""
		err = core.ErrorAuthenticationFailure
		return
	}

	err = c.UserStore.Check(user, pass)
	if err != nil {
		user = ""
		err = core.ErrorAuthenticationFailure
		return
	}

	return
}

func (c *BasicAuth) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, c.Realm))
	http.Error(w, "authorization failed", http.StatusUnauthorized)
	return nil
}

func (c *BasicAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
