package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb/memory"
	"github.com/stretchr/testify/assert"
)

const (
	loginform = `
<html>
<head>
</head>
<body>
<form method="post">
username <input type="text" name="username"><br>
password <input type="password" name="password" autocomplete="off"><br>
<input type="submit">
<hr>
{{.}}

<body>
</html>
`
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	s, err := memory.NewMemStore()
	assert.Nil(err)

	f, err := NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "",
		LoginPasswordField: "p",
		UserStore:          s,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	_, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "",
		UserStore:          s,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	_, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "u",
		UserStore:          s,
	})
	assert.Equal(core.ErrorInvalidConfig, err)
	assert.Nil(f)

	f, err = NewFormAuth(Config{
		LoginForm:          "",
		LoginUsernameField: "u",
		LoginPasswordField: "p",
		UserStore:          s,
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
		LoginForm:          loginform,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
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
		LoginForm:          loginform,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,
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
		LoginForm:          loginform,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,

		// Validation options:
		MinUsernameLength: 1,
		MaxUsernameLength: 100,
		MinPasswordLength: 1,
		MaxPasswordLength: 100,
		UsernamePattern:   ".*",
		PasswordPattern:   ".*",
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
		LoginForm:          loginform,
		LoginUsernameField: "username",
		LoginPasswordField: "password",
		UserStore:          userdb,

		// Validation options:
		MinUsernameLength: 1,
		MaxUsernameLength: 100,
		MinPasswordLength: 1,
		MaxPasswordLength: 100,
		UsernamePattern:   ".*",
		PasswordPattern:   ".*",
	})
	assert.Nil(err)

	data := url.Values{"username": {"bob"}, "password": {"thebuilder"}}
	r, err := http.NewRequest("POST", "/", strings.NewReader(data.Encode()))
	assert.Nil(err)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = provider.Check(r)
	assert.Equal(core.ErrorAuthenticationFailure, err)
}
