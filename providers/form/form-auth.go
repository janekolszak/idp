package form

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"

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

	Username     Complexity
	Password     Complexity
	UserStore    userdb.UserStore
	UserVerifier userdb.UserVerifier

	// Directory with all needed html templates
	TemplateDir string
}

type FormAuth struct {
	Config
	templates *template.Template
}

func NewFormAuth(c Config) (*FormAuth, error) {
	if c.LoginUsernameField == "" ||
		c.LoginPasswordField == "" ||
		c.LoginUsernameField == c.LoginPasswordField {
		return nil, core.ErrorInvalidConfig
	}

	if len(c.Password.Patterns) == 0 {
		c.Password.Patterns = []string{".*"}
	}

	auth := FormAuth{Config: c}

	govalidator.TagMap["password"] = govalidator.Validator(func(str string) bool {
		return auth.Config.Password.Validate(str)
	})

	var err error
	auth.templates, err = template.ParseGlob(filepath.Join(c.TemplateDir, "*.html"))
	if err != nil {
		return nil, err
	}

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

func (f *FormAuth) Register(r *http.Request) (id string, err error) {
	// Parse and validate posted form
	data, err := NewRegisterPOST(r)
	if err != nil {
		return
	}

	user := userdb.User{
		Username:  data.Username,
		Email:     data.Email,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	}

	id, err = f.UserStore.Insert(&user, data.Password)
	if err != nil {
		return
	}

	_, err = f.UserVerifier.Push(id, data.Username, data.Email)
	return
}

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

func (f *FormAuth) WriteRegister(w http.ResponseWriter, r *http.Request) error {
	return f.templates.ExecuteTemplate(w, "register.html", nil)
}
