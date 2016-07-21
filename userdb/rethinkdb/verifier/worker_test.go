package verifier

import (
	// "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"html/template"

	"net/smtp"
	// "time"
)

var (
	testWorkerOpts WorkerOpts
)

func SetupWorkerTests() {
	// Credentials for mailtrap.io account
	testWorkerOpts = WorkerOpts{
		Session:       session,
		HostAddress:   "mailtrap.io:2525",
		Sender:        "",
		SmtpAuth:      smtp.CRAMMD5Auth("bffc19805c8b87", "5e86eb71d6ba55"),
		EmailTemplate: template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, click {{.URL}} to verify")),
	}
}

// TODO: More tests

func TestNewWorker(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	w, err := NewWorker(testWorkerOpts)
	assert.Nil(err)
	assert.NotNil(w)
}

func TestWorkerSimple(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	verifier, err := NewVerifier(session)
	assert.Nil(err)
	assert.NotNil(verifier)

	w, err := NewWorker(testWorkerOpts)
	assert.Nil(err)
	assert.NotNil(w)

	err = w.Start()
	assert.Nil(err)

	verifier.PushVerification("userID", "username", "to@mailtrap.io")

	w.Stop()
}
