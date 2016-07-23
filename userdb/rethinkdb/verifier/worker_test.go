package verifier

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
	"time"
)

var (
	testWorkerOpts WorkerOpts
)

func SetupWorkerTests() {
	// Credentials for a test mailtrap.io account
	testWorkerOpts = WorkerOpts{
		Session:            session,
		Host:               "mailtrap.io",
		Port:               2525,
		User:               "bffc19805c8b87",
		Password:           "5e86eb71d6ba55",
		From:               "5a7ed0268f-6f6f88@inbox.mailtrap.io",
		TextTemplate:       template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, visit {{.URL}} to verify!")),
		HtmlTemplate:       template.Must(template.New("tmpl").Parse("Hi! {{.Username}}, click <a href={{.URL}}> here </a> to verify!")),
		EndpointAddress:    "https://example.com/verify",
		VerificationMaxAge: time.Millisecond * 1,
		CleanupInterval:    time.Millisecond * 500,
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

func TestWorkerCleanup(t *testing.T) {
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
	defer w.Stop()

	verifier.PushVerification("userID", "username", "to@example.com")

	time.Sleep(3 * time.Second)

	// Old verification should be removed
	count, err := verifier.Count()
	assert.Nil(err)
	assert.Equal(0, int(count))
}
