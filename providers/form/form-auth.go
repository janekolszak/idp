package form

import (
	"github.com/asaskevich/govalidator"
	"html/template"
	"net/http"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"
)

type Config struct {
	LoginForm          string
	LoginUsernameField string
	LoginPasswordField string
	MinUsernameLength  int
	MaxUsernameLength  int
	UsernamePattern    string
	MinPasswordLength  int
	MaxPasswordLength  int
	PasswordPattern    string
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

	if c.UsernamePattern == "" {
		c.UsernamePattern = ".*"
	}

	if c.PasswordPattern == "" {
		c.PasswordPattern = ".*"
	}

	auth := FormAuth{Config: c}
	return &auth, nil
}

func (f *FormAuth) Check(r *http.Request) (user string, err error) {
	user = r.FormValue(f.LoginUsernameField)
	if !govalidator.IsByteLength(user, f.Config.MinUsernameLength, f.Config.MaxUsernameLength) ||
		!govalidator.Matches(user, f.Config.UsernamePattern) {
		err = core.ErrorBadRequest
		return
	}

	password := r.FormValue(f.LoginPasswordField)
	if !govalidator.IsByteLength(password, f.Config.MinPasswordLength, f.Config.MaxPasswordLength) ||
		!govalidator.Matches(password, f.Config.PasswordPattern) {
		err = core.ErrorBadRequest
		return
	}

	err = f.UserStore.Check(user, password)
	if err != nil {
		user = ""
		err = core.ErrorAuthenticationFailure
	}

	return
}

func (f *FormAuth) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	msg := ""
	if r.Method == "POST" && err != nil {
		switch err {
		case core.ErrorAuthenticationFailure:
			msg = "Authentication failed"

		default:
			msg = "An error occurred"
		}
	}
	t := template.Must(template.New("tmpl").Parse(f.LoginForm))
	return t.Execute(w, msg)
}

func (f *FormAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
