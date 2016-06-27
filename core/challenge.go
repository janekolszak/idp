package core

import (
	"crypto/rsa"
	"encoding/gob"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"net/http"
	"time"
)

const (
	SessionCookieName = "challenge"
)

var (
	challengeStore sessions.Store
)

type Challenge struct {
	token      *jwt.Token
	consentKey *rsa.PrivateKey

	// TODO: Add sessions.Session field

	// TODO: Remove
	TokenStr string

	// Set in the challenge endpoint, after authenticated.
	User   string
	Client string
	Scopes []string
}

func init() {
	// Gob is used by gorilla sessions
	gob.Register(&Challenge{})
}

func GetChallenge(w http.ResponseWriter, r *http.Request) (*Challenge, error) {
	session, err := challengeStore.Get(r, SessionCookieName)
	if err != nil {
		return nil, err
	}

	c, ok := session.Values[SessionCookieName].(*Challenge)
	if !ok {
		return nil, ErrorBadChallengeCookie
	}

	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Challenge) Save(w http.ResponseWriter, r *http.Request) error {
	session, err := challengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}
	session.Values[SessionCookieName] = c
	return session.Save(r, w)
}

func (c *Challenge) GrantAccess(w http.ResponseWriter, r *http.Request, subject string, scopes []string) error {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	challengeClaims := c.token.Claims.(jwt.MapClaims)

	claims["aud"] = challengeClaims["aud"]
	claims["exp"] = now.Add(time.Minute * 5).Unix()
	claims["iat"] = now.Unix()
	claims["scp"] = scopes
	claims["sub"] = subject

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(c.consentKey)
	if err != nil {
		return err
	}

	// Remove the challenge data from the cookie
	// TODO: Remove the cookie instead
	session, err := challengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}
	delete(session.Values, SessionCookieName)
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	http.Redirect(w, r, challengeClaims["redir"].(string)+"&consent="+tokenString, http.StatusFound)

	return nil
}
