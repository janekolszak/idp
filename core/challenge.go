package core

import (
	"encoding/gob"
	"fmt"
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

	// TODO: Add sessions.Session field

	Client   *hclient.Client
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
	session, err := c.idp.config.ChallengeStore.New(r, SessionCookieName)
	if err != nil {
		return err
	}

	fmt.Println("MaxAge", session.Options.MaxAge)
	session.Options = c.idp.createChallengeCookieOptions
	fmt.Println("MaxAge", session.Options.MaxAge)

	session.Values[SessionCookieName] = c
	return c.idp.config.ChallengeStore.Save(r, w, session)
}

func (c *Challenge) Delete(w http.ResponseWriter, r *http.Request) error {
	session, err := c.idp.config.ChallengeStore.Get(r, SessionCookieName)
	if err != nil {
		return err
	}

	session.Options = c.idp.deleteChallengeCookieOptions
	return c.idp.config.ChallengeStore.Save(r, w, session)
}

func (c *Challenge) RefuseAccess(w http.ResponseWriter, r *http.Request) error {
	err := c.Delete(w, r)
	if err != nil {
		return err
	}

	http.Redirect(w, r, c.Redirect+"&consent=false", http.StatusFound)

	return nil
}

func (c *Challenge) GrantAccessToAll(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()

	// TODO: Validate Challenge before using the data

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = c.Client.GetID()
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
