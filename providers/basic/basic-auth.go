package basic

import (
	"fmt"
	"net/http"

	"github.com/janekolszak/idp/core"
	"golang.org/x/crypto/bcrypt"
)

// Basic Authentication checker.
// Expects Storage to return plain text passwords
type BasicAuth struct {
	Htpasswd Htpasswd
	Realm    string
}

func NewBasicAuth(htpasswdFileName string, realm string) (*BasicAuth, error) {
	b := new(BasicAuth)

	err := b.Htpasswd.Load(htpasswdFileName)
	if err != nil {
		return nil, err
	}

	b.Realm = realm

	return b, nil
}

func (c *BasicAuth) Check(r *http.Request) (user string, err error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		user = ""
		err = core.ErrorAuthenticationFailure
		return
	}

	hash, err := c.Htpasswd.Get(user)
	if err != nil {
		// Prevent timing attack
		bcrypt.CompareHashAndPassword([]byte{}, []byte(pass))
		user = ""
		err = core.ErrorAuthenticationFailure
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		user = ""
		err = core.ErrorAuthenticationFailure
	}

	return
}

func (c *BasicAuth) WriteError(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, c.Realm))
	http.Error(w, "authorization failed", http.StatusUnauthorized)
	return nil
}
