package form

import (
	"testing"

	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/userdb/memory"
	"github.com/stretchr/testify/assert"
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

func TestUnknownUser(t *testing.T) {

}
