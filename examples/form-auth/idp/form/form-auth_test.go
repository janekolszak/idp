package form

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb/memory"

	// "fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	loginPage = `
<html>
<head>
</head>
<body>
    <form method="post">
        Username
        <input type="text" name="username">
        <br> Password
        <input type="password" name="password" autocomplete="off">
        <br>
        <input type="submit">
    </form>
    <body>
</html>
`

	registerPage = `
<html>
<head>
</head>
<body>
    <form method="post">
        Name
        <input type="text" name="name">
        <br> Last Name
        <input type="text" name="lastname">
        <br> Email
        <input type="text" name="email">
        <br> Username
        <input type="text" name="username">
        <br> Password
        <input type="password" name="password" autocomplete="off">
        <br> Confirmed Password
        <input type="password" name="confirmedpassword" autocomplete="off">
        <br>
        <input type="submit">
    </form>
    <body>
</html>
`

	resetPage = `
<html>
<head>
</head>
<body>
Welcome {{.Username}}! Please type in the new password.
<br>
    <form method="post">
        Password
        <input type="password" name="password" autocomplete="off">
        <br>
        Confirm Password
        <input type="password" name="confirmedpassword" autocomplete="off">
        <br>
        <input type="submit">
    </form>
    <body>
</html>
`

	verifyPage = `
<html>
<head>
</head>
<body>
Welcome {{.Username}}
Your email is verified.
<body>
</html>
`
)

var (
	testTemplates string
)

func TestMain(m *testing.M) {
	var err error
	testTemplates, err = ioutil.TempDir("", "idp-form-auth-templates-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(testTemplates) // clean up

	// Write temp files
	tmpfn := filepath.Join(testTemplates, "login.html")
	if err := ioutil.WriteFile(tmpfn, []byte(loginPage), 0666); err != nil {
		panic(err)
	}

	tmpfn = filepath.Join(testTemplates, "register.html")
	if err := ioutil.WriteFile(tmpfn, []byte(registerPage), 0666); err != nil {
		panic(err)
	}

	tmpfn = filepath.Join(testTemplates, "reset.html")
	if err = ioutil.WriteFile(tmpfn, []byte(resetPage), 0666); err != nil {
		panic(err)
	}

	tmpfn = filepath.Join(testTemplates, "verify.html")
	if err := ioutil.WriteFile(tmpfn, []byte(verifyPage), 0666); err != nil {
		panic(err)
	}

	status := m.Run()
	os.RemoveAll(testTemplates)
	os.Exit(status)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)

	s, err := memory.NewMemStore()
	assert.Nil(err)

	f, err := NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "",
		LoginPasswordField: "p",
		UserStore:          s,
		TemplateDir:        testTemplates,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	_, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "",
		UserStore:          s,
		TemplateDir:        testTemplates,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	_, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "u",
		UserStore:          s,
		TemplateDir:        testTemplates,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	f, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "p",
		UserStore:          s,
		TemplateDir:        testTemplates,
	})
	assert.Nil(err)
	assert.NotNil(f)
}

func createUsers(assert *assert.Assertions) *memory.Store {
	userdb, err := memory.NewMemStore()
	assert.Nil(err)

	err = userdb.Add("bob", "bob123")
	assert.Nil(err)

	return userdb
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	userdb := createUsers(assert)

	// Create the provider
	provider, err := NewFormAuth(Config{
		LoginForm:          loginPage,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
		TemplateDir:        testTemplates,
	})
	assert.Nil(err)

	r, err := http.NewRequest("GET", "/", nil)
	assert.Nil(err)

	u, err := provider.Check(r)
	assert.Equal(core.ErrorAuthenticationFailure, err)
	assert.Equal("", u)
}

func TestNoHeader(t *testing.T) {
	assert := assert.New(t)
	userdb := createUsers(assert)

	// Create the provider
	provider, err := NewFormAuth(Config{
		LoginForm:          loginPage,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
		TemplateDir:        testTemplates,
	})
	assert.Nil(err)

	r := &http.Request{}
	_, err = provider.Check(r)
	assert.NotNil(err)
}

func TestPostSuccess(t *testing.T) {
	assert := assert.New(t)
	userdb := createUsers(assert)

	// Create the provider
	provider, err := NewFormAuth(Config{
		LoginForm:          loginPage,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
		TemplateDir:        testTemplates,

		// Validation options:
		Username: Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
		Password: Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
	})
	assert.Nil(err)

	data := url.Values{"username": {"bob"}, "password": {"bob123"}}
	r, err := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	assert.Nil(err)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = provider.Check(r)
	assert.Nil(err)
}

func TestPostFail(t *testing.T) {
	assert := assert.New(t)
	userdb := createUsers(assert)

	// Create the provider
	provider, err := NewFormAuth(Config{
		LoginForm:          loginPage,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
		TemplateDir:        testTemplates,

		// Validation options:
		Username: Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
		Password: Complexity{
			MinLength: 1,
			MaxLength: 100,
			Patterns:  []string{".*"},
		},
	})
	assert.Nil(err)

	data := url.Values{"username": {"bob"}, "password": {"thebuilder"}}
	r, err := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	assert.Nil(err)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = provider.Check(r)
	assert.Equal(core.ErrorAuthenticationFailure, err)
}
