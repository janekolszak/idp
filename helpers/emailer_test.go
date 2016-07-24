package helpers

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

var (
	testEmailerOpts = EmailerOpts{
		Host:         "mailtrap.io",
		Port:         2525,
		User:         "bffc19805c8b87",
		Password:     "5e86eb71d6ba55",
		From:         "5a7ed0268f-6f6f88@inbox.mailtrap.io",
		TextTemplate: template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, visit {{.URL}} to verify!")),
		HtmlTemplate: template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, click <a href={{.URL}}> here </a> to verify!")),
		Domain:       "example.com",
	}
)

func TestEmailer(t *testing.T) {
	assert := assert.New(t)

	e, err := NewEmailer(testEmailerOpts)
	assert.Nil(err)
	assert.NotNil(e)

	data := map[string]string{
		"Username": "Username",
		"URL":      "URL",
	}

	err = e.Send("to@example.com", data)
	assert.Nil(err)
}
