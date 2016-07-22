package verifier

import (
	// "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"html/template"
	// "time"
)

var (
	testWorkerOpts WorkerOpts
)

func SetupWorkerTests() {
	// Credentials for mailtrap.io account
	testWorkerOpts = WorkerOpts{
		Session:         session,
		Host:            "mailtrap.io",
		Port:            2525,
		User:            "bffc19805c8b87",
		Password:        "5e86eb71d6ba55",
		From:            "5a7ed0268f-6f6f88@inbox.mailtrap.io",
		ContentType:     "text/plain",
		EmailTemplate:   template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, click {{.Username}} to verify!")),
		EndpointAddress: "https://example.com/verify",
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

	verifier.PushVerification("userID", "username", "to@example.com")

	w.Stop()
}
