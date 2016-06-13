package core

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"time"
)

type IdP struct {
	Port          int    `yaml:"port"`
	ClientID      string `yaml:"client_id"`
	ClientSecret  string `yaml:"client_secret"`
	HydraAddress  string `yaml:"token_endpoint"`
	TokenEndpoint string `yaml:"token_endpoint"`

	// Checks if a user-password pair is valid
	Provider Provider

	// Http client form communicating with Hydra
	client *http.Client
	// Key for challenge JWT verification
	verificationKey *rsa.PublicKey

	// Key for signing the consent JWT
	consentKey *rsa.PrivateKey
}

// Gets the requested key from Hydra
func (idp *IdP) getKey(set string, kind string) (*gojwk.Key, error) {
	url := idp.HydraAddress + "/keys/" + set + "/" + kind

	resp, err := idp.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	key, err := gojwk.Unmarshal(body)
	if err != nil {
		return nil, err
	}

	return key.Keys[0], nil
}

// Downloads the hydra's public key
func (idp *IdP) getVerificationKey() error {
	jwk, err := idp.getKey("consent.challenge", "public")
	if err != nil {
		return err
	}

	key, err := jwk.DecodePublicKey()
	if err != nil {
		return err
	}

	idp.verificationKey = key.(*rsa.PublicKey)

	return err
}

// Downloads the private key used for signing the consent
func (idp *IdP) getConsentKey() error {
	jwk, err := idp.getKey("consent.endpoint", "private")
	if err != nil {
		return err
	}

	key, err := jwk.DecodePrivateKey()
	if err != nil {
		return err
	}

	idp.consentKey = key.(*rsa.PrivateKey)

	return err
}

func (idp *IdP) login() error {
	// Use the credentials to login to Hydra
	credentials := clientcredentials.Config{
		ClientID:     idp.ClientID,
		ClientSecret: idp.ClientSecret,
		TokenURL:     idp.HydraAddress + "/oauth2/token",
		Scopes:       []string{"core", "hydra.keys.get"},
	}

	// Skip verifying the certificate
	// TODO: Remove when Hydra implements passing key-cert pairs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c)

	// Prefetch the token - tests the connection``
	_, err := credentials.Token(ctx)
	if err != nil {
		return err
	}

	idp.client = credentials.Client(ctx)

	return nil
}

func (idp *IdP) Connect() error {
	err := idp.login()
	if err != nil {
		return err
	}

	err = idp.getVerificationKey()
	if err != nil {
		return err
	}

	err = idp.getConsentKey()
	if err != nil {
		return err
	}

	return err
}

// Parse and verify the challenge JWT
func (idp *IdP) getChallengeToken(challengeString string) (*jwt.Token, error) {
	token, err := jwt.Parse(challengeString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return idp.verificationKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("Empty token")
	}

	return token, nil
}

// Generate the consent
func (idp *IdP) generateConsentToken(challenge *jwt.Token, subject string, scopes []string) (string, error) {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["aud"] = challenge.Claims["aud"]
	token.Claims["exp"] = now.Add(time.Minute * 5).Unix()
	token.Claims["iat"] = now.Unix()
	token.Claims["scp"] = scopes
	token.Claims["sub"] = subject

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(idp.consentKey)
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func (idp *IdP) GetConsentGET() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idp.Provider.Respond(w, r)
	})
}

func (idp *IdP) GetConsentPOST() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("New request")
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		challengeTokenStr := r.Form.Get("challenge")
		if challengeTokenStr == "" {
			http.Error(w, "No challenge token", http.StatusBadRequest)
			return
		}

		fmt.Println("New request")
		challengeToken, err := idp.getChallengeToken(challengeTokenStr)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Checking!\n")
		// TODO: Check session cookie if present
		// TODO: Check credentials if present
		// TODO: Get the credentials from the form
		err = idp.Provider.Check(r)
		if err != nil {
			fmt.Println(err.Error())
			// TODO: Log the real error
			if err == ErrorAuthenticationFailure {
				fmt.Printf("Authentication Failure, responding!\n")
				idp.Provider.Respond(w, r)
				return
			}
			// Bad credentials
			fmt.Println(err.Error())
			http.Error(w, "Bad Credentials", http.StatusBadRequest)
			return
		}

		consentTokenStr, err := idp.generateConsentToken(challengeToken, "joe@joe", []string{"read", "write"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("Access granted!\n")

		// TODO: Redirect only after checking user's credentials.
		http.Redirect(w, r, challengeToken.Claims["redir"].(string)+"&consent="+consentTokenStr, http.StatusFound)
	})
}
