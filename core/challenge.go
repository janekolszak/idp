package core

import (
	"encoding/gob"
	jwt "github.com/dgrijalva/jwt-go"
	// "github.com/gorilla/sessions"
	"net/http"
	"time"
)

const (
	SessionCookieName = "challenge"
)

type Challenge struct {
	// Parent IDP that got the challenge
	idp *IDP

	// TODO: Add sessions.Session field

	Client   string
	Expires  time.Time
	Redirect string
	Scopes   []string

	// Set in the challenge endpoint, after authenticated.
	User string
}

func init() {
	// Gob is used by gorilla sessions
	gob.Register(&Challenge{})
}

// Saves the challenge to it's session store
func (c *Challenge) Save(w http.ResponseWriter, r *http.Request) error {
	session, err := c.idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}
	session.Values[SessionCookieName] = c
	return session.Save(r, w)
}

func (c *Challenge) GrantAccessToAll(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()

	// TODO: Validate Challenge before using the data

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = c.Client
	claims["exp"] = now.Add(time.Minute * 5).Unix()
	claims["iat"] = now.Unix()
	claims["scp"] = c.Scopes
	claims["sub"] = c.User

	// Sign and get the complete encoded token as a string
	key, err := c.idp.GetConsentKey()
	if err != nil {
		return err
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		return err
	}

	// Remove the cookie
	session, err := c.idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}
	delete(session.Values, SessionCookieName)
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	// All this work could have been too long (fetching key might be time consuming)
	if c.Expires.Before(time.Now()) {
		return ErrorChallengeExpired
	}

	http.Redirect(w, r, c.Redirect+"&consent="+tokenString, http.StatusFound)

	return nil
}
