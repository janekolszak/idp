package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
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
	config     Config
	client     *http.Client
	verifyKey  *rsa.PublicKey
	consentKey *rsa.PrivateKey
)

func getChallengeToken(challengeString string) *jwt.Token {
	// Verify JWK
	token, err := jwt.Parse(challengeString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return verifyKey, nil
	})

	if err != nil {
		panic(err)
	}

	if !token.Valid {
		panic(nil)
	}

	return token
}

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

func handleConsent(c *gin.Context) {
	var data Consent
	if c.Bind(&data) == nil {
		// panic(nil)
	}
	challengeToken := getChallengeToken(data.Challenge)
	consentTokenString := generateConsentToken(challengeToken, "joe@joe", []string{"read", "write"})

	fmt.Printf("Granted access!\n")

	c.Redirect(http.StatusFound, challengeToken.Claims["redir"].(string)+"&consent="+consentTokenString)
}

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

func readConfig() {
	// IdP is has credentials preconfigured by hydra.
	// Let's parse the yaml file with that information

	source, err := ioutil.ReadFile("/root/.hydra.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
}

func loginToHydra() {
	// Use the credentials to login to hydra
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

	// Here are the tokens we got from Hydra
	fmt.Printf("Client Credentials flow completed, got:\n")
	fmt.Printf("AccessToken: %s\n", token.AccessToken)
	fmt.Printf("TokenType:   %s\n", token.TokenType)
	fmt.Printf("Expiry:      %s\n", token.Expiry.Local())
	fmt.Printf("\n\n")

	client = credentials.Client(ctx)
}

func getVerifyKey() {
	fmt.Printf("Getting the JWK (needed for JWT verification):\n")
	key := getKey(client, "consent.challenge", "public")
	buf, _ := json.Marshal(key)
	fmt.Printf("Key:         %s\n", string(buf))

	publicKey, err := key.DecodePublicKey()
	if err != nil {
		panic(err)
	}

	verifyKey = publicKey.(*rsa.PublicKey)
}

func getConsentKey() {
	fmt.Printf("Getting the JWK (needed for JWT signing):\n")
	key := getKey(client, "consent.endpoint", "private")
	buf, _ := json.Marshal(key)
	fmt.Printf("Key:         %s\n", string(buf))

	k, err := key.DecodePrivateKey()
	if err != nil {
		panic(err)
	}

	consentKey = k.(*rsa.PrivateKey)
}

func main() {

	// Read the configuration file
	readConfig()

	// Obtain a token from Hydra
	// and get the http.Client that automatically passes that token to requests
	loginToHydra()

	// Obtain the public key used to verify consent challenge JWT
	getVerifyKey()

	// Obtain the private key used to sign consents
	getConsentKey()

	r := gin.Default()
	r.GET("/", handleConsent)

	r.Run(":3000")
}
