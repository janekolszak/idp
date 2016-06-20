package core

import (
	"crypto/rsa"
	jwt "github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

type Challenge struct {
	token      *jwt.Token
	consentKey *rsa.PrivateKey

	// Is client in a predefined list?
	trusted bool

	// TODO: Remove
	TokenStr string

	// Set in the challenge endpoint, after authenticated
	User   string
	Scopes []string
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

	http.Redirect(w, r, challengeClaims["redir"].(string)+"&consent="+tokenString, http.StatusFound)

	return nil
}
