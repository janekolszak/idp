package form

import (
	"html/template"
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
	err = r.ParseForm()
	if err != nil {
		return
	}

	user = r.Form.Get(f.LoginUsernameField)
	password := r.Form.Get(f.LoginPasswordField)

	err = f.UserStore.Check(user, password)
	if err != nil {
		user = ""
		err = core.ErrorAuthenticationFailure
	}

	return
}

func (a *FormAuth) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	msg := ""
	if r.Method == "POST" && err != nil {
		switch err {
		case core.ErrorAuthenticationFailure:
			msg = "Authentication failed"

		default:
			msg = "An error occurred"
		}
	}
	t := template.Must(template.New("tmpl").Parse(a.LoginForm))
	return t.Execute(w, msg)
}

func (a *FormAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
