package helpers

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"github.com/gorilla/sessions"
	"github.com/satori/go.uuid"
	"net/http"
)

const (
	rememberMeValidatorLen = 32
)

var (
	rememberMeStore sessions.Store
)

type RememberMe struct {
	cookieName string

	// Selector is assigned by
	Selector  string
	Validator string
}

func init() {
	// Gob is used by gorilla sessions
	gob.Register(&RememberMe{})

	// TODO: Initialize somewhere else
	rememberMeStore = sessions.NewCookieStore([]byte("something-very-secret"))
}

func NewRememberMe(selector string, cookieName string) (*RememberMe, error) {
	var rm = new(RememberMe)

	rm.Selector = selector
	rm.cookieName = cookieName

	if selector == "" {
		uniqueID := uuid.NewV1()
		rm.Selector = uniqueID.String()
	}

	return rm, nil
}

func GetRememberMe(r *http.Request, cookieName string) (*RememberMe, error) {
	session, err := rememberMeStore.Get(r, cookieName)
	if err != nil {
		return nil, err
	}

	rm, ok := session.Values["r"].(*RememberMe)
	if !ok {
		return nil, errors.New("Bad remember me cookie format")
	}

	rm.cookieName = cookieName

	return rm, nil
}

// Compute the sha-256
func (rm *RememberMe) validatorHash() string {
	hasher := sha1.New()
	hasher.Write([]byte(rm.Validator))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Fills Validator with a new value and return it's sha-256 hash.
// This hash, together with Selector need to be saved for later comparison.
func (rm *RememberMe) GenerateValidator() (string, error) {
	b := make([]byte, rememberMeValidatorLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	rm.Validator = base64.URLEncoding.EncodeToString(b)

	return rm.validatorHash(), nil
}

// Check if Validator value is valid...
func (rm *RememberMe) Check(value string) bool {
	return rm.validatorHash() == value
}

func (rm *RememberMe) Save(w http.ResponseWriter, r *http.Request) error {
	session, err := rememberMeStore.Get(r, rm.cookieName)
	if err != nil {
		return err
	}

	session.Values["r"] = rm

	return session.Save(r, w)
}
