package helpers

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"github.com/gorilla/sessions"
	"net/http"
)

const (
	rememberMeValidatorLen = 32
)

var (
	rememberMeStore sessions.Store
)

// Implementation of https://paragonie.com/blog/2015/04/secure-authentication-php-with-long-term-persistence#title.2
type LoginCookie struct {
	Selector   string
	Validator  string
	CookieName string
}

func init() {
	// Gob is used by gorilla sessions
	gob.Register(&LoginCookie{})

	// TODO: Initialize somewhere else
	rememberMeStore = sessions.NewCookieStore([]byte("something-very-secret"))
}

func GetLoginCookie(r *http.Request, cookieName string) (*LoginCookie, error) {
	session, err := rememberMeStore.Get(r, cookieName)
	if err != nil {
		return nil, err
	}

	l, ok := session.Values["r"].(*LoginCookie)
	if !ok {
		return nil, errors.New("Bad remember me cookie format")
	}

	l.CookieName = cookieName

	return l, nil
}

// Compute the sha-256
func (l *LoginCookie) validatorHash() string {
	hasher := sha1.New()
	hasher.Write([]byte(l.Validator))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Fills Validator with a new value and return it's sha-256 hash.
// This hash, together with Selector need to be saved for later comparison.
func (l *LoginCookie) GenerateValidator() (string, error) {
	b := make([]byte, rememberMeValidatorLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	l.Validator = base64.URLEncoding.EncodeToString(b)

	return l.validatorHash(), nil
}

// Check if Validator value is valid...
func (l *LoginCookie) Check(value string) bool {
	return l.validatorHash() == value
}

func (l *LoginCookie) Save(w http.ResponseWriter, r *http.Request) error {
	session, err := rememberMeStore.Get(r, l.CookieName)
	if err != nil {
		return err
	}

	session.Values["r"] = l

	return session.Save(r, w)
}
