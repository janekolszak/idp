package verifier

import (
	// "os"
	"testing"

	"github.com/stretchr/testify/assert"
	// "time"
)

// TODO: More tests

func TestNewWorker(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	w, err := NewWorker(session)
	assert.Nil(err)
	assert.NotNil(w)
}

func TestWorkerSimple(t *testing.T) {
	assert := assert.New(t)
	Cleanup()

	verifier, err := NewVerifier(session)
	assert.Nil(err)
	assert.NotNil(verifier)

	w, err := NewWorker(session)
	assert.Nil(err)
	assert.NotNil(w)

	w.Start()
	verifier.PushVerification("userID", "email")
	w.Stop()
}
