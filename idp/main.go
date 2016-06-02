package main

import (
	_ "fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	_ "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
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

	// client := credentials.Client(context.Background())
	_, err = credentials.Token(context.Background())
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/", handleConsent)

	r.Run(":3000")
}
