package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"

	"crypto/tls"
	"net/http"

	"github.com/mendsley/gojwk"
)

// Structure of the .hydra.yml file shared with Hydra
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

func handleConsent(c *gin.Context) {
	var data Consent
	if c.Bind(&data) == nil {
	}

	c.JSON(200, gin.H{"challenge": data.Challenge})
}

func getKey(client *http.Client, set string, kind string) string {
	url := os.Getenv("HYDRA_URL") + "/keys/" + set + "/" + kind
	fmt.Printf("Url:       %s\n\n", url)

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func main() {

	// IdP is has credentials preconfigured by hydra.
	// Let's parse the yaml file with that information
	var config Config
	source, err := ioutil.ReadFile("/root/.hydra.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}

	// Use the credentials to login to hydra
	credentials := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     os.Getenv("HYDRA_URL") + "/oauth2/token",
		Scopes:       []string{"core", "hydra.keys.get"},
	}

	// Skip verifying
	// TODO: Remove when Hydra implements passing key-cert pairs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, client)

	token, err := credentials.Token(ctx)
	if err != nil {
		panic(err)
	}

	c := credentials.Client(ctx)
	// Here are the tokens we got from Hydra
	fmt.Printf("Client Credentials flow completed, got:\n")
	fmt.Printf("AccessToken:  %s\n", token.AccessToken)
	fmt.Printf("TokenType:    %s\n", token.TokenType)
	fmt.Printf("RefreshToken: %s\n", token.RefreshToken)
	fmt.Printf("Expiry:       %s\n", token.Expiry.Local())
	fmt.Printf("\n")
	fmt.Printf("Getting the JWK (needed for JWT verification):\n")
	key := getKey(c, "consent.challenge", "public")
	fmt.Printf("Key           %s\n", key)

	// r := gin.Default()
	// r.GET("/", handleConsent)

	// r.Run(":3000")
}
