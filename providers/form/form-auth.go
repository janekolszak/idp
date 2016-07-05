package form

import (
	"net/http"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"
)

type Config struct {
	LoginForm          string
	LoginUsernameField string
	LoginPasswordField string
	UserStore          userdb.Store
}

type FormAuth struct {
	Config
}

func NewFormAuth(c Config) (*FormAuth, error) {
	if c.LoginUsernameField == "" ||
		c.LoginPasswordField == "" ||
		c.LoginUsernameField == c.LoginPasswordField {
		return nil, core.ErrorInvalidConfig
	}
	auth := FormAuth{Config: c}
	return &auth, nil
}

func (f *FormAuth) Check(r *http.Request) (user string, err error) {
	r.ParseForm()
	user = r.Form.Get(f.LoginUsernameField)
	password := r.Form.Get(f.LoginPasswordField)

	err = f.UserStore.Check(user, password)
	if err != nil {
		user = ""
		err = core.ErrorAuthenticationFailure
	}

	return
}

func (a *FormAuth) Respond(w http.ResponseWriter, r *http.Request) error {
	return nil
}
