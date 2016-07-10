package form

import (
	"html/template"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"
)

type Complexity struct {
	MinLength int
	MaxLength int
	Patterns  []string
}

func (c *Complexity) Validate(s string) bool {
	if !govalidator.IsByteLength(s, c.MinLength, c.MaxLength) {
		return false
	}
	for _, p := range c.Patterns {
		if !govalidator.Matches(s, p) {
			return false
		}
	}
	return true
}

type Config struct {
	LoginForm          string
	LoginUsernameField string
	LoginPasswordField string
	Username           Complexity
	Password           Complexity
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

	if len(c.Username.Patterns) == 0 {
		c.Username.Patterns = []string{".*"}
	}

	if len(c.Password.Patterns) == 0 {
		c.Password.Patterns = []string{".*"}
	}

	auth := FormAuth{Config: c}
	return &auth, nil
}

func (f *FormAuth) Check(r *http.Request) (user string, err error) {
	user = r.FormValue(f.LoginUsernameField)
	if !f.Config.Username.Validate(user) {
		user = ""
		err = core.ErrorBadRequest
		return
	}

	password := r.FormValue(f.LoginPasswordField)
	if !f.Config.Password.Validate(password) {
		user = ""
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
