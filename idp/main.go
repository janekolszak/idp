package main

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/mendsley/gojwk"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Structure of the .hydra.yml file shared with Hydra via a volume
type Config struct {
	Port         int    `yaml:"port"`
	Issuer       string `yaml:"issuer"`
	ConsentURL   string `yaml:"consent_url"`
	ClusterURL   string `yaml:"cluster_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// Form to parse when granting a consent
type Consent struct {
	Challenge string `form:"challenge" binding:"required"`
}

var (
	// Configuration file
	config Config

	// Http client form communicating with Hydra
	client *http.Client

	// Key for challenge JWT verification
	verificationKey *rsa.PublicKey

	// Key for signing the consent JWT
	consentKey *rsa.PrivateKey
)

// IdP has its credentials preconfigured by Hydra.
// This function parses the yaml file with that information
func readConfig() {
	data, err := ioutil.ReadFile("/root/.hydra.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

// Gets the requested key from Hydra
func getKey(client *http.Client, set string, kind string) *gojwk.Key {
	url := os.Getenv("HYDRA_URL") + "/keys/" + set + "/" + kind
	fmt.Printf("Url:         %s\n", url)

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	key, err := gojwk.Unmarshal(body)
	if err != nil {
		panic(err)
	}

	return key.Keys[0]
}

// Use the credentials to login to Hydra
func loginToHydra() {
	credentials := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     os.Getenv("HYDRA_URL") + "/oauth2/token",
		Scopes:       []string{"core", "hydra.keys.get"},
	}

	// Skip verifying the certificate
	// TODO: Remove when Hydra implements passing key-cert pairs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c)

	token, err := credentials.Token(ctx)
	if err != nil {
		panic(err)
	}

	// Here is the token we got from Hydra
	fmt.Printf("Client Credentials flow completed, got:\n")
	fmt.Printf("AccessToken: %s\n", token.AccessToken)
	fmt.Printf("TokenType:   %s\n", token.TokenType)
	fmt.Printf("Expiry:      %s\n\n", token.Expiry.Local())

	client = credentials.Client(ctx)
}

// Downloads the hydra's public key
func getVerificationKey() {
	fmt.Printf("Getting the JWK (needed for JWT verification):\n")

	key := getKey(client, "consent.challenge", "public")
	fmt.Printf("Key type:    %s\n\n", key.Kty)

	publicKey, err := key.DecodePublicKey()
	if err != nil {
		panic(err)
	}

	verificationKey = publicKey.(*rsa.PublicKey)
}

// Downloads the private key used for signing the consent
func getConsentKey() {
	fmt.Printf("Getting the JWK (needed for JWT signing):\n")

	key := getKey(client, "consent.endpoint", "private")
	fmt.Printf("Key type:    %s\n\n", key.Kty)

	k, err := key.DecodePrivateKey()
	if err != nil {
		panic(err)
	}

	consentKey = k.(*rsa.PrivateKey)
}

// Parse and verify the challenge JWT
func getChallengeToken(challengeString string) *jwt.Token {
	token, err := jwt.Parse(challengeString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return verificationKey, nil
	})

	if err != nil || !token.Valid {
		panic(err)
	}

	return token
}

// Generate the consent
func generateConsentToken(challenge *jwt.Token, subject string, scopes []string) string {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["aud"] = challenge.Claims["aud"]
	token.Claims["exp"] = now.Add(time.Minute * 5).Unix()
	token.Claims["iat"] = now.Unix()
	token.Claims["scp"] = scopes
	token.Claims["sub"] = subject

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(consentKey)
	if err != nil {
		panic(err)
	}

	return tokenString
}

// Start serving the consent endpoint
func serveConsentEndpoint() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		var data Consent
		if c.Bind(&data) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "BadRequest"})
		}

		challengeToken := getChallengeToken(data.Challenge)

		consentTokenString := generateConsentToken(challengeToken, "joe@joe", []string{"read", "write"})

		fmt.Printf("Access granted!\n")

		// TODO: Redirect only after checking user's credentials.
		c.Redirect(http.StatusFound, challengeToken.Claims["redir"].(string)+"&consent="+consentTokenString)
	})

	r.Run(":3000")
}

func main() {
	// Read the configuration file
	readConfig()

	// Obtain a token from Hydra
	// and get the http.Client that automatically passes that token to requests
	loginToHydra()

	// Obtain the public key used to verify consent challenge JWT
	getVerificationKey()

	// Obtain the private key used to sign consents
	getConsentKey()

	// Start serving the consent endpoint
	serveConsentEndpoint()
}
