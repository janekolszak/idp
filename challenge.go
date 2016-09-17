// Package for handling challenge requests from Hydra(https://github.com/ory-am/hydra).
package idp

import (
	"encoding/gob"
	jwt "github.com/dgrijalva/jwt-go"
	// "github.com/gorilla/sessions"
	hclient "github.com/ory-am/hydra/client"
	"net/http"
	"time"
)

const (
	SessionCookieName = "challenge"
)

type Challenge struct {
	// Parent IDP that got the challenge
	idp *IDP

	// Hydra's client
	Client *hclient.Client

	// Time of expiration
	Expires time.Time

	// Redirect URL
	Redirect string

	// Requested scopes
	Scopes []string

	// Set in the challenge endpoint, after authenticated.
	User string
}

func init() {
	// Gob is used by gorilla sessions
	gob.Register(&Challenge{})
}

// Saves the Challenge to it's session store
func (c *Challenge) Save(w http.ResponseWriter, r *http.Request) error {
	session, err := c.idp.config.ChallengeStore.New(r, SessionCookieName)
	if err != nil {
		return err
	}

	session.Options = c.idp.createChallengeCookieOptions
	session.Values[SessionCookieName] = c

	return c.idp.config.ChallengeStore.Save(r, w, session)
}

// Update the Challenge, e.g. add user representation
func (c *Challenge) Update(w http.ResponseWriter, r *http.Request) error {
	session, err := c.idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}

	session.Options = c.idp.createChallengeCookieOptions
	session.Values[SessionCookieName] = c

	return c.idp.config.ChallengeStore.Save(r, w, session)
}

// Deletes the challenge from the store
func (c *Challenge) Delete(w http.ResponseWriter, r *http.Request) error {
	session, err := c.idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}

	session.Options = c.idp.deleteChallengeCookieOptions
	return c.idp.config.ChallengeStore.Save(r, w, session)
}

// User refused access to requested scopes, forward the desicion to Hydra via redirection.
func (c *Challenge) RefuseAccess(w http.ResponseWriter, r *http.Request) error {
	err := c.Delete(w, r)
	if err != nil {
		return err
	}

	http.Redirect(w, r, c.Redirect+"&consent=false", http.StatusFound)

	return nil
}

// User granted access to requested scopes, forward the desicion to Hydra via redirection.
func (c *Challenge) GrantAccessToAll(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = c.Client.GetID()
	claims["exp"] = now.Add(time.Minute * 5).Unix()
	claims["iat"] = now.Unix()
	claims["scp"] = c.Scopes
	claims["sub"] = c.User

	// Sign and get the complete encoded token as a string
	key, err := c.idp.getConsentKey()
	if err != nil {
		return err
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		return err
	}

	// Delete the cookie
	err = c.Delete(w, r)
	if err != nil {
		return err
	}

	// All this work might have taken too long (fetching key may be time consuming)
	// so check token expiration
	if c.Expires.Before(time.Now()) {
		return ErrorChallengeExpired
	}

	http.Redirect(w, r, c.Redirect+"&consent="+tokenString, http.StatusFound)

	return nil
}
