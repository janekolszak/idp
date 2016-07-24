package form

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb"
)

type LoginFormContext struct {
	Msg         string
	SubmitURI   string
	RegisterURI string
}

type Config struct {
	LoginForm          string
	LoginUsernameField string
	LoginPasswordField string

	RegisterUsernameField        string
	RegisterPasswordField        string
	RegisterPasswordConfirmField string

	Username  Complexity
	Password  Complexity
	UserStore userdb.Store
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

// func (f *FormAuth) Register(r *http.Request) (user string, err error) {
// 	user = r.FormValue(f.RegisterUsernameField)
// 	password := r.FormValue(f.RegisterPasswordField)
// 	confirm := r.FormValue(f.RegisterPasswordConfirmField)

// 	if password != confirm {
// 		err = core.ErrorPasswordMismatch
// 	}

// 	if !f.Config.Password.Validate(password) {
// 		err = core.ErrorComplexityFailed
// 	}

// 	if !f.Config.Username.Validate(user) {
// 		err = core.ErrorComplexityFailed
// 	}

// 	if err != nil {
// 		user = ""
// 		return
// 	}

// 	err = f.UserStore.Add(user, password)
// 	return
// }

func (f *FormAuth) WriteError(w http.ResponseWriter, r *http.Request, err error) error {
	query := url.Values{}
	query["challenge"] = []string{r.URL.Query().Get("challenge")}
	context := LoginFormContext{
		SubmitURI:   r.URL.RequestURI(),
		RegisterURI: fmt.Sprintf("/register?%s", query.Encode()),
	}

	if r.Method == "POST" && err != nil {
		switch err {
		case core.ErrorAuthenticationFailure:
			context.Msg = "Authentication failed"

		default:
			context.Msg = "An error occurred"
		}
	}
	t := template.Must(template.New("tmpl").Parse(f.LoginForm))
	return t.Execute(w, context)
}

func (f *FormAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}
